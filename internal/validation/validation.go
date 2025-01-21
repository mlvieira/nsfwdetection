package validation

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gabriel-vasile/mimetype"
)

// allowedTypes Supported MIME types
var allowedTypes = map[string]bool{
	"image/jpeg": true, // JPEG
	"image/png":  true, // PNG
	"image/gif":  true, // GIF
	"image/webp": true, // WebP
	"image/jp2":  true, // JPEG 2000
	"image/jxr":  true, // JPEG XR
	"image/jfif": true, // JFIF
}

// ValidateFileType Validates file type by reading header
func ValidateFileType(file multipart.File) error {
	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return fmt.Errorf("failed to detect MIME type: %w", err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	if !allowedTypes[mime.String()] {
		return fmt.Errorf("unsupported MIME type: %s", mime.String())
	}

	return nil
}
