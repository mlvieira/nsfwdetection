package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/models"
	"github.com/mlvieira/nsfwdetection/internal/services"
	"github.com/mlvieira/nsfwdetection/internal/utils"
)

type APIHandlers struct {
	*Handlers
	Services *services.APIService
}

func NewAPIHandlers(h *Handlers, api *services.APIService) *APIHandlers {
	return &APIHandlers{
		Handlers: h,
		Services: api,
	}
}

func (a *APIHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request: %v", err)
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	tokenString, err := a.Services.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, models.LoginResponse{Token: tokenString})

}

func (a *APIHandlers) PaginationUploads(w http.ResponseWriter, r *http.Request) {
	var req models.PaginatedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	response, err := a.Services.PaginationUploads(r.Context(), req.ID, req.Limit, req.Reviewed)
	if err != nil {
		logger.Error("Error fetching uploads: %v", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *APIHandlers) LabelImage(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	var req models.LabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	response, err := a.Services.LabelImage(r.Context(), hash, req, a.Hub)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (a *APIHandlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	var req models.LabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	response, err := a.Services.DeleteImage(r.Context(), hash, req, a.Hub)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}
