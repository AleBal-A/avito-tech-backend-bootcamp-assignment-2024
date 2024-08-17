package flatHandler

import (
	"avito/internal/custommiddleware"
	"avito/internal/domain/models"
	"avito/internal/handlers/common"
	"avito/internal/handlers/response"
	"avito/internal/services/flatService"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type FlatHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	flatService flatService.FlatService
	logger      *slog.Logger
}

func NewHandler(flatService flatService.FlatService, logger *slog.Logger) FlatHandler {
	return &Handler{
		flatService: flatService,
		logger:      logger,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "flatHandler.Create"

	var req struct {
		HouseID    int  `json:"house_id"`
		FlatNumber *int `json:"flat_number"`
		Price      int  `json:"price"`
		Rooms      int  `json:"rooms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request", slog.String("op", op), "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	flat, err := h.flatService.Create(r.Context(), req.HouseID, req.FlatNumber, req.Price, req.Rooms)
	if err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Could not create flat", op, err)
		return
	}

	resp := response.FlatResponse{
		ID:      flat.ID,
		HouseID: flat.HouseID,
		Price:   flat.Price,
		Rooms:   flat.Rooms,
		Status:  flat.Status,
	}

	h.logger.Info("Flat is created", slog.String("op", op), slog.Int("flat_id", flat.ID))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const op = "flatHandler.Update"

	var req struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request", slog.String("op", op), "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{
		"created":       true,
		"approved":      true,
		"declined":      true,
		"on moderation": true,
	}
	if !validStatuses[req.Status] {
		h.logger.Error("Invalid status value", slog.String("op", op), slog.String("status", req.Status))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(custommiddleware.ClaimsContextKey).(*models.Claims)
	if !ok || claims == nil {
		h.logger.Error("Claims are missing in context", slog.String("op", op))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	if claims.Role != "moderator" {
		h.logger.Error("User is not authorized to update flat status", slog.String("op", op), slog.String("user_id", userID))
		h.logger.Error("claims.Role DEBUG - ", slog.String("op", op), slog.String("claims.Role", claims.Role))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	flat, err := h.flatService.UpdateStatus(r.Context(), req.ID, req.Status, userID)
	if err != nil {
		if errors.Is(err, flatService.ErrFlatBeingModerated) {
			h.logger.Warn("Flat is already being moderated by another user", slog.String("op", op), slog.Int("flat_id", req.ID))
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.logger.Error("Failed to update flat status", slog.String("op", op), "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := response.FlatResponse{
		ID:      flat.ID,
		HouseID: flat.HouseID,
		Price:   flat.Price,
		Rooms:   flat.Rooms,
		Status:  flat.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}
