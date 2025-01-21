package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/models"
)

func GetClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// WriteJSONError sends a JSON-formatted error response
func WriteJSONError(w http.ResponseWriter, statusCode int, message string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var detail string
	if len(detail) > 0 {
		detail = details[0]
	}

	errorResponse := models.ErrorResponse{
		Error:   message,
		Code:    statusCode,
		Details: detail,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
		WriteJSONError(w, http.StatusInternalServerError, "Failed to encode response")
	}
}
