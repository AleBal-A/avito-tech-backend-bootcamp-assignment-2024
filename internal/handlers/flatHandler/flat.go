package flatHandler

import (
	"avito/internal/services/flatService"
	"encoding/json"
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

func NewHandler(flatService flatService.FlatService, logger *slog.Logger) *Handler {
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
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	flat, err := h.flatService.Create(r.Context(), req.HouseID, req.FlatNumber, req.Price, req.Rooms)
	if err != nil {
		h.logger.Error("Could not create flat", slog.String("op", op), "error", err)
		http.Error(w, "Could not create flat", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Flat is created", slog.String("op", op), slog.Int("flat_id", flat.ID))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(flat); err != nil {
		h.logger.Error("Failed to write response", slog.String("op", op), "error", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// TODO: прописать в мидлваре только для модеров
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const op = "flatHandler.Update"

	var req struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request", slog.String("op", op), "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	flat, err := h.flatService.UpdateStatus(r.Context(), req.ID, req.Status)
	if err != nil {
		h.logger.Error("Could not update flat status", slog.String("op", op), slog.Int("flat_id", req.ID), "error", err)
		http.Error(w, "Could not update flat status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(flat); err != nil {
		h.logger.Error("Failed to write response", slog.String("op", op), "error", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
