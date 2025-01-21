package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	cursorID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || cursorID < 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	limit, err := strconv.Atoi(chi.URLParam(r, "limit"))
	if err != nil || limit <= 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid limit parameter")
		return
	}

	reviewedParam := r.URL.Query().Get("reviewed")
	var reviewed *bool
	if reviewedParam != "" {
		parsedReviewed, err := strconv.ParseBool(reviewedParam)
		if err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid reviewed parameter, must be 'true' or 'false'")
			return
		}
		reviewed = &parsedReviewed
	}

	response, err := a.Services.PaginationUploads(r.Context(), cursorID, limit, reviewed)
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
