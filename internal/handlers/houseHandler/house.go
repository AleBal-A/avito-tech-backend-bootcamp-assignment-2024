package houseHandler

import (
	"avito/internal/custommiddleware"
	"avito/internal/domain/models"
	"avito/internal/handlers/common"
	"avito/internal/handlers/response"
	"avito/internal/services/houseService"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
)

type HouseHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetFlatsByHouseID(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	houseService houseService.HouseService
	logger       *slog.Logger
}

func NewHandler(houseService houseService.HouseService, logger *slog.Logger) HouseHandler {
	return &Handler{
		houseService: houseService,
		logger:       logger,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "houseHandler.Create"

	var req struct {
		Address   string  `json:"address"`
		YearBuilt int     `json:"year"`
		Builder   *string `json:"developer"`
	}

	h.logger.Debug("Start of creating a house", slog.String("op", op))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid input data", slog.String("op", op), "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Debug("Received house creation request", slog.String("op", op), slog.String("address", req.Address))

	house, err := h.houseService.Create(r.Context(), req.Address, req.YearBuilt, req.Builder)
	if err != nil {
		if errors.Is(err, houseService.ErrValidation) {
			h.logger.Error("Validation error", slog.String("op", op), "error", err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Server error", op, err)
		}
		return
	}

	resp := response.HouseResponse{
		Id:        house.ID,
		Address:   house.Address,
		Year:      house.YearBuilt,
		Developer: checkString(house.Builder),
		CreatedAt: house.CreatedAt,
		UpdateAt:  house.CreatedAt,
	}

	h.logger.Info("House created successfully", slog.String("op", op), slog.Int("house_id", house.ID))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}

func (h *Handler) GetFlatsByHouseID(w http.ResponseWriter, r *http.Request) {
	const op = "houseHandler.GetFlatsByHouseID"

	houseIDStr := chi.URLParam(r, "id")
	if houseIDStr == "" {
		h.logger.Error("House ID is missing in the request", slog.String("op", op))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	houseID, err := strconv.Atoi(houseIDStr)
	if err != nil {
		h.logger.Error("Invalid house ID format", slog.String("op", op), "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(custommiddleware.ClaimsContextKey).(*models.Claims)
	if !ok || claims == nil {
		h.logger.Error("Claims are missing in context", slog.String("op", op))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	flats, err := h.houseService.GetFlatsByHouseID(r.Context(), houseID, claims.Role)
	if err != nil {
		h.logger.Error("Failed to get flats by house ID", slog.String("op", op), "error", err)
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Could not retrieve flats", op, err)
		return
	}

	var resp []response.FlatResponse
	for _, flat := range flats {
		resp = append(resp, response.FlatResponse{
			ID:      flat.ID,
			HouseID: flat.HouseID,
			Price:   flat.Price,
			Rooms:   flat.Rooms,
			Status:  flat.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"flats": resp}); err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}

func checkString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}
