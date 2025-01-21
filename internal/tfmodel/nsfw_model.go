package tfmodel

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wamuir/graft/tensorflow"
)

// SharedNSFWModel is the global, shared model loaded at startup
var SharedNSFWModel *Model

// Model represents the NSFW detection model
type Model struct {
	model *tensorflow.SavedModel
}

// Prediction represents the output for NSFW detection
type Prediction struct {
	ID             int     `json:"id"`              // Job ID (used by worker)
	NSFWPercentage float32 `json:"nsfw_percentage"` // NSFW percentage
	SFWPercentage  float32 `json:"sfw_percentage"`  // SFW percentage
	Duration       float64 `json:"duration"`        // Processing time in seconds
	Timestamp      int64   `json:"timestamp"`       // UNIX timestamp
	UUID           string  `json:"uuid"`            // Unique identifier
	SHA256         string  `json:"sha256"`          // SHA256 hash
	Error          string  `json:"error,omitempty"` // Error message
	Trace          string  `json:"trace,omitempty"` // Error trace
	Success        bool    `json:"success"`         // Success flag
}

// LoadModel initializes and loads the TensorFlow model
func LoadModel(modelPath string) error {
	model, err := tensorflow.LoadSavedModel(modelPath, []string{"serve"}, nil)
	if err != nil {
		return fmt.Errorf("error loading mordel: %w", err)
	}

	SharedNSFWModel = &Model{model: model}

	return nil
}

// DetectNSFW processes an image and returns its NSFW score using the model.
func (m *Model) DetectNSFW(imagePath string) (*Prediction, error) {
	startTime := time.Now()
	predictionID := uuid.New().String()

	tensor, err := PreprocessImage(imagePath)
	if err != nil {
		return &Prediction{
			UUID:      predictionID,
			Timestamp: time.Now().Unix(),
			Duration:  float64(time.Since(startTime).Seconds()),
			Error:     fmt.Sprintf("error preprocessing image: %v", err),
			Trace:     "PreprocessImage -> invalid input format",
			Success:   false,
		}, fmt.Errorf("error preprocessing image: %w", err)
	}

	output, err := m.model.Session.Run(
		map[tensorflow.Output]*tensorflow.Tensor{
			m.model.Graph.Operation("serving_default_input").Output(0): tensor,
		},
		[]tensorflow.Output{
			m.model.Graph.Operation("StatefulPartitionedCall_1").Output(0),
		},
		nil,
	)
	if err != nil {
		return &Prediction{
			UUID:      predictionID,
			Timestamp: time.Now().Unix(),
			Duration:  time.Since(startTime).Seconds(),
			Error:     fmt.Sprintf("error running inference: %v", err),
			Trace:     "Session.Run -> model execution failed",
			Success:   false,
		}, fmt.Errorf("error running inference: %w", err)
	}

	// validate results
	scores, ok := output[0].Value().([][]float32)
	if !ok || len(scores) < 1 || len(scores[0]) < 2 {
		return &Prediction{
			UUID:      predictionID,
			Timestamp: time.Now().Unix(),
			Duration:  time.Since(startTime).Seconds(),
			Error:     "invalid output format",
			Trace:     "Output parsing -> format mismatch",
			Success:   false,
		}, errors.New("invalid output format")
	}

	nsfwScore := scores[0][1]
	sfwScore := 1.0 - nsfwScore
	nsfwPercentage := nsfwScore * 100
	sfwPercentage := sfwScore * 100

	duration := time.Since(startTime).Seconds()

	return &Prediction{
		NSFWPercentage: nsfwPercentage,
		SFWPercentage:  sfwPercentage,
		Duration:       duration,
		Timestamp:      time.Now().Unix(),
		UUID:           predictionID,
		Success:        true,
	}, nil
}

// Close releases TensorFlow resources
func (m *Model) Close() {
	if m == nil || m.model == nil {
		return
	}

	m.model.Session.Close()
}
