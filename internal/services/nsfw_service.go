package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlvieira/nsfwdetection/internal/config"
	"github.com/mlvieira/nsfwdetection/internal/driver/redis"
	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/models"
	"github.com/mlvieira/nsfwdetection/internal/repositories"
	"github.com/mlvieira/nsfwdetection/internal/tfmodel"
	"github.com/mlvieira/nsfwdetection/internal/validation"
	"github.com/mlvieira/nsfwdetection/internal/websockets"
	"github.com/mlvieira/nsfwdetection/internal/worker"
)

// NSFWService defines the business logic for NSFW processing
type NSFWService struct {
	redisClient  *redis.RedisClient
	hub          *websockets.Hub
	repositories *repositories.Repositories
}

// NewNSFWService creates a new instance of NSFWService
func NewNSFWService(redisClient *redis.RedisClient, hub *websockets.Hub, repositories *repositories.Repositories) *NSFWService {
	return &NSFWService{
		redisClient:  redisClient,
		hub:          hub,
		repositories: repositories,
	}
}

// ProcessFiles handles NSFW processing for uploaded files
func (s *NSFWService) ProcessFiles(ctx context.Context, files []*multipart.FileHeader) ([]*tfmodel.Prediction, error) {
	var output []*tfmodel.Prediction

	for id, fileHeader := range files {
		fileStartTime := time.Now()
		logger.Info("Processing file: %s (ID: %d)", fileHeader.Filename, id)

		file, err := fileHeader.Open()
		if err != nil {
			logger.Error("Failed to open file: %w", err)
			output = append(output, s.createPredictionError(id, "Failed to open file", fileHeader.Filename, fileStartTime))
			continue
		}
		defer file.Close()

		sha256Hash, err := s.computeSHA256(file)
		if err != nil {
			logger.Error("Failed to compute hash: %w", err)
			output = append(output, s.createPredictionError(id, "Failed to compute hash", fileHeader.Filename, fileStartTime))
			continue
		}
		file.Seek(0, io.SeekStart)

		if err := validation.ValidateFileType(file); err != nil {
			logger.Error("Failed to validate type: %w", err)
			output = append(output, s.createPredictionError(id, err.Error(), fileHeader.Filename, fileStartTime))
			continue
		}

		cachedPrediction := s.checkCache(ctx, sha256Hash, id, fileStartTime)
		if cachedPrediction != nil {
			output = append(output, cachedPrediction)
			continue
		}

		prediction := s.processPrediction(file, fileHeader, id, sha256Hash, fileStartTime)
		if prediction == nil {
			logger.Error("Model failed to determine score: %w", fileHeader.Filename)
			output = append(output, s.createPredictionError(id, "Prediction failed", fileHeader.Filename, fileStartTime))
			continue
		}

		s.storeCache(ctx, sha256Hash, prediction)

		uploadedImage := s.CreateUploadedImage(prediction, fileHeader)

		if err = s.repositories.Uploaded.UploadImage(ctx, uploadedImage); err != nil {
			logger.Error("Failed to save uploaded image to database: %v", err)
			output = append(output, s.createPredictionError(id, "Failed to save image to database", fileHeader.Filename, fileStartTime))
			continue
		}

		s.NotifyClients(uploadedImage)

		output = append(output, prediction)
	}

	return output, nil
}

// checkCache retrieves a cached prediction result from Redis by SHA-256 hash.
func (s *NSFWService) checkCache(ctx context.Context, sha256Hash string, id int, startTime time.Time) *tfmodel.Prediction {
	cacheKey := fmt.Sprintf("nsfw:%s", sha256Hash)
	cachedResult, err := s.redisClient.GetValue(ctx, cacheKey)

	if err == nil && cachedResult != "" {
		var cachedPrediction tfmodel.Prediction

		if err := json.NewDecoder(bytes.NewReader([]byte(cachedResult))).Decode(&cachedPrediction); err == nil {
			logger.Info("Cache hit for file (SHA256: %s)", sha256Hash)
			cachedPrediction.ID = id
			cachedPrediction.Timestamp = time.Now().Unix()
			cachedPrediction.Duration = float64(time.Since(startTime).Seconds())
			return &cachedPrediction
		}
	}

	return nil
}

// storeCache saves a prediction result in Redis with a 60-minute expiration.
func (s *NSFWService) storeCache(ctx context.Context, sha256Hash string, prediction *tfmodel.Prediction) {
	cacheKey := fmt.Sprintf("nsfw:%s", sha256Hash)
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(prediction); err == nil {
		s.redisClient.SetValue(ctx, cacheKey, buffer.String(), 60*time.Minute)
	} else {
		logger.Error("Failed to store cache for key %s: %v", cacheKey, err)
	}
}

// processPrediction saves the uploaded file temporarily and submits it to the worker pool for NSFW detection.
func (s *NSFWService) processPrediction(file multipart.File, fileHeader *multipart.FileHeader, id int, sha256Hash string, startTime time.Time) *tfmodel.Prediction {
	ext := filepath.Ext(fileHeader.Filename)

	tempFile, err := os.CreateTemp(config.AppConfig.FileHandling.TempUploadDir, "upload-*"+ext)
	if err != nil {
		logger.Error("Failed to create temp file for: %s, Error: %v", fileHeader.Filename, err)
		return nil
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		logger.Error("Failed to save temp file for: %s, Error: %v", fileHeader.Filename, err)
		return nil
	}

	resultChan := make(chan *tfmodel.Prediction, 1)
	worker.SubmitJob(worker.Job{
		ID:          id,
		FilePath:    tempFile.Name(),
		ResultsChan: resultChan,
	})

	var prediction *tfmodel.Prediction
	select {
	case prediction = <-resultChan:
		logger.Info("Received result for job %d", id)
	case <-time.After(5 * time.Second):
		logger.Error("Timeout for job %d", id)
		prediction = &tfmodel.Prediction{
			ID:        id,
			Error:     "Timeout processing file",
			Success:   false,
			Timestamp: time.Now().Unix(),
			Duration:  float64(time.Since(startTime).Seconds()),
		}
	}
	close(resultChan)

	// use goroutine to move file to later rate
	go func() {
		destPath := filepath.Join(config.AppConfig.FileHandling.UploadDir, sha256Hash+ext)

		if err = os.Rename(tempFile.Name(), destPath); err != nil {
			logger.Error("Failed to move file to uploads: %v", err)
			return
		}

		logger.Info("File saved to: %s", destPath)
	}()

	prediction.SHA256 = sha256Hash
	prediction.ID = id
	prediction.Timestamp = time.Now().Unix()
	prediction.Duration = float64(time.Since(startTime).Seconds())

	return prediction
}

// createPredictionError generates a prediction result with an error message.
func (s *NSFWService) createPredictionError(id int, errorMsg, filename string, startTime time.Time) *tfmodel.Prediction {
	return &tfmodel.Prediction{
		ID:        id,
		Error:     fmt.Sprintf("%s: %s", filename, errorMsg),
		Success:   false,
		Timestamp: time.Now().Unix(),
		Duration:  float64(time.Since(startTime).Seconds()),
	}
}

// computeSHA256 calculates the SHA-256 hash of the uploaded file for caching purposes.
func (s *NSFWService) computeSHA256(file multipart.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ValidateFiles parses the multipart form and retrieves files with keys matching 'files[...]'.
func (s *NSFWService) ValidateFiles(r *http.Request) ([]*multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(config.AppConfig.FileHandling.MaxFileSizeMB); err != nil {
		logger.Error("Failed to parse form: %v", err)
		return nil, errors.New("failed to parse form data")
	}

	files := make([]*multipart.FileHeader, 0)

	for key, fileHeaders := range r.MultipartForm.File {
		if strings.HasPrefix(key, "files[") {
			files = append(files, fileHeaders...)
		}
	}

	if len(files) == 0 {
		logger.Error("No files uploaded in request.")
		return nil, errors.New("no files uploaded")
	}

	return files, nil
}

// WriteJSONResponse sends a JSON response with the given status code and prediction results.
func (s *NSFWService) WriteJSONResponse(w http.ResponseWriter, statusCode int, results []*tfmodel.Prediction) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error("Failed to encode JSON response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// SendErrorResponse sends a JSON error response with the specified status code and error message.
func (s *NSFWService) SendErrorResponse(w http.ResponseWriter, errorMsg string, startTime time.Time, statusCode int) {
	s.WriteJSONResponse(w, statusCode, []*tfmodel.Prediction{
		{
			Error:     errorMsg,
			Success:   false,
			Timestamp: time.Now().Unix(),
			Duration:  float64(time.Since(startTime).Seconds()),
		},
	})
}

// NotifyClients sends a prediction result to all connected WebSocket clients
func (s *NSFWService) NotifyClients(uploaded models.UploadedImage) {
	image := map[string]any{
		"event": "new_upload",
		"data":  uploaded,
	}

	message, err := json.Marshal(image)
	if err != nil {
		logger.Error("Failed to marshal WebSocket message: %v", err)
		return
	}
	s.hub.Broadcast <- message
}

func (s *NSFWService) CreateUploadedImage(prediction *tfmodel.Prediction, fileHeader *multipart.FileHeader) models.UploadedImage {
	label := "SFW"
	score := prediction.SFWPercentage
	if prediction.NSFWPercentage > prediction.SFWPercentage {
		label = "NSFW"
		score = prediction.NSFWPercentage
	}

	path := fmt.Sprintf("/static/uploads/%s%s", prediction.SHA256, filepath.Ext(fileHeader.Filename))

	uploadedImage := models.UploadedImage{
		FilePath:   path,
		FileHash:   prediction.SHA256,
		Label:      label,
		NewLabel:   "unlabeled",
		Confidence: score,
		Reviewed:   false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return uploadedImage
}
