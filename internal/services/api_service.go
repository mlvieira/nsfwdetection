package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mlvieira/nsfwdetection/internal/config"
	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/models"
	"github.com/mlvieira/nsfwdetection/internal/repositories"
	"github.com/mlvieira/nsfwdetection/internal/websockets"
)

type APIService struct {
	hub          *websockets.Hub
	repositories *repositories.Repositories
}

var jwtSecretKey = []byte(config.AppConfig.Security.JWTSecretKey)

func NewAPIService(hub *websockets.Hub, repositories *repositories.Repositories) *APIService {
	return &APIService{
		hub:          hub,
		repositories: repositories,
	}
}

func (s *APIService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repositories.User.CheckLogin(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("Invalid username or password")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		logger.Error("Failed to sign token: %v", err)
		return "", fmt.Errorf("Failed to generate token")
	}

	return tokenString, nil
}

func (s *APIService) PaginationUploads(ctx context.Context, cursorID, limit int, reviewed *bool) (models.PaginatedResponse, error) {
	uploads, err := s.repositories.Uploaded.ListUploadsCursor(ctx, cursorID, limit, reviewed)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	totalCount, err := s.repositories.Uploaded.ListTotalUploads(ctx, reviewed)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	if uploads == nil {
		uploads = []models.UploadedImage{}
	}

	return models.PaginatedResponse{
		Data:  uploads,
		Count: len(uploads),
		Total: totalCount,
	}, nil
}

func (s *APIService) LabelImage(ctx context.Context, hash string, req models.LabelRequest, hub *websockets.Hub) (models.AckResponse, error) {
	if req.Event != "rate" && req.Event != "update_rate" {
		return models.AckResponse{}, fmt.Errorf("Invalid event type")
	}

	if req.Rating != "NSFW" && req.Rating != "SFW" {
		return models.AckResponse{}, fmt.Errorf("Invalid rating")
	}

	if req.Sha256 != hash {
		return models.AckResponse{}, fmt.Errorf("Hash mismatch in URL and payload")
	}

	status := map[string]string{
		"event":  "in_progress",
		"sha256": hash,
	}
	statusMsg, _ := json.Marshal(status)
	s.hub.Broadcast <- statusMsg

	rows, err := s.repositories.Uploaded.LabelUpload(ctx, req.Sha256, req.Rating)
	if err != nil {
		return models.AckResponse{}, err
	}

	if rows == 0 {
		return models.AckResponse{}, fmt.Errorf("Image not found")
	}

	response := models.AckResponse{
		Event:  "ack_rating",
		Sha256: hash,
		Status: "success",
	}
	statusMsg, _ = json.Marshal(response)
	s.hub.Broadcast <- statusMsg

	return response, nil
}

func (s *APIService) DeleteImage(ctx context.Context, hash string, req models.LabelRequest, hub *websockets.Hub) (models.AckResponse, error) {
	if req.Event != "delete" {
		return models.AckResponse{}, fmt.Errorf("Invalid event type")
	}

	if req.Sha256 != hash {
		return models.AckResponse{}, fmt.Errorf("Hash mismatch in URL and payload")
	}

	status := map[string]string{
		"event":  "in_progress",
		"sha256": hash,
	}
	statusMsg, _ := json.Marshal(status)
	s.hub.Broadcast <- statusMsg

	path, err := s.repositories.Uploaded.GetFilePathByHash(ctx, req.Sha256)
	if err != nil {
		return models.AckResponse{}, fmt.Errorf("Failed to fetch file path")
	}

	if path == "" {
		return models.AckResponse{}, fmt.Errorf("File does not exist")
	}

	_, fileName := filepath.Split(path)
	backendPath := filepath.Join(config.AppConfig.FileHandling.UploadDir, fileName)

	if err = os.Remove(backendPath); err != nil && !os.IsNotExist(err) {
		return models.AckResponse{}, fmt.Errorf("Failed to delete file from filesystem")
	}

	rows, err := s.repositories.Uploaded.DeleteImage(ctx, req.Sha256)
	if err != nil {
		return models.AckResponse{}, fmt.Errorf("Failed to delete image from database")
	}

	if rows == 0 {
		return models.AckResponse{}, fmt.Errorf("Image not found")
	}

	response := models.AckResponse{
		Event:  "ack_delete",
		Sha256: hash,
		Status: "success",
	}
	statusMsg, _ = json.Marshal(response)
	s.hub.Broadcast <- statusMsg

	return response, nil
}
