package handlers

import (
	"net/http"
	"time"

	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/services"
	"github.com/mlvieira/nsfwdetection/internal/utils"
)

type NSFWHandlers struct {
	*Handlers
	Services *services.NSFWService
}

func NewNSFWHandlers(h *Handlers, nsfw *services.NSFWService) *NSFWHandlers {
	return &NSFWHandlers{
		Handlers: h,
		Services: nsfw,
	}
}

// NSFWHandler processes HTTP requests for NSFW detection
func (n *NSFWHandlers) NSFWHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	clientIP := utils.GetClientIP(r)
	logger.Info("Request received from: %s, Method: %s", clientIP, r.Method)

	if r.Method != http.MethodPost {
		n.Services.SendErrorResponse(w, "Invalid request method", startTime, http.StatusMethodNotAllowed)
		return
	}

	files, err := n.Services.ValidateFiles(r)
	if err != nil {
		n.Services.SendErrorResponse(w, "Failed to parse form", startTime, http.StatusBadRequest)
		return
	}

	output, err := n.Services.ProcessFiles(r.Context(), files)
	if err != nil {
		n.Services.SendErrorResponse(w, "Failed to process files", startTime, http.StatusInternalServerError)
		return
	}

	n.Services.WriteJSONResponse(w, http.StatusOK, output)

}
