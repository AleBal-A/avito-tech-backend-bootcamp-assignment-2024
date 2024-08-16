package common

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
)

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger, statusCode int, message, operation string, err error) {
	requestID := middleware.GetReqID(r.Context())
	if requestID == "" {
		requestID = "unknown"
	}

	logger.Error(message, slog.String("op", operation), slog.String("request_id", requestID), slog.String("error", err.Error()))

	response := struct {
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
		Code      int    `json:"code"`
	}{
		Message:   "что-то пошло не так",
		RequestID: requestID,
		Code:      statusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to write error response", slog.String("op", operation), slog.String("request_id", requestID), slog.String("error", err.Error()))
	}
}
