package tfmodel

import (
	"fmt"

	"github.com/wamuir/graft/tensorflow"
)

// PreprocessImage preprocesses an image into a TensorFlow tensor
func PreprocessImage(filePath string) (*tensorflow.Tensor, error) {
	img, format, err := openImage(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening image: %w", err)
	}

	finalFrame, err := handleAnimatedFrames(filePath, img, format)
	if err != nil {
		return nil, fmt.Errorf("error handling frames: %w", err)
	}

	tensor, err := resizeAndNormalize(finalFrame)
	if err != nil {
		return nil, fmt.Errorf("error processing: %w", err)
	}

	return tensor, nil
}
