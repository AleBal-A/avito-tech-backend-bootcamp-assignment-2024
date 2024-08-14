package houseHandler

import (
	"avito/internal/services/houseService"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
)

type HouseHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message, operation string, err error)
}

type Handler struct {
	houseService houseService.HouseService
	logger       *slog.Logger
}

func NewHandler(houseService houseService.HouseService, logger *slog.Logger) *Handler {
	return &Handler{
		houseService: houseService,
		logger:       logger,
	}
}

// TODO: only moderator
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "houseHandler.Create"

	var req struct {
		Address   string  `json:"address"`
		YearBuilt int     `json:"year_built"`
		Builder   *string `json:"builder"`
	}

	h.logger.Debug("Start of creating a house", slog.String("op", op))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid input data", op, err)
		return
	}

	h.logger.Debug("Received house creation request", slog.String("op", op), slog.String("address", req.Address))

	house, err := h.houseService.Create(r.Context(), req.Address, req.YearBuilt, req.Builder)
	if err != nil {
		if errors.Is(err, houseService.ErrValidation) {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "Validation error", op, err)
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "Server error", op, err)
		}
		return
	}

	h.logger.Info("House created successfully", slog.String("op", op), slog.Int("house_id", house.ID))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(house); err != nil {
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message, operation string, err error) {
	requestID := middleware.GetReqID(r.Context())
	if requestID == "" {
		requestID = "unknown"
	}

	h.logger.Error(message, slog.String("op", operation), slog.String("request_id", requestID), "error", err)

	response := struct {
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
		Code      int    `json:"code"`
	}{
		Message:   message,
		RequestID: requestID,
		Code:      statusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write error response", slog.String("op", operation), slog.String("request_id", requestID), "error", err)
	}
}
