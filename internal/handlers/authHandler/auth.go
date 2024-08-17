package authHandler

import (
	"avito/internal/handlers/common"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"avito/internal/domain/models"
	"avito/internal/repositories"
	"avito/internal/services/authService"
)

type AuthHandler interface {
	DummyLogin(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	ValidateToken(tokenStr string) (*models.Claims, error)
}

type Handler struct {
	authService authService.AuthService
	logger      *slog.Logger
}

func NewHandler(authService authService.AuthService, logger *slog.Logger) AuthHandler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

// DummyLogin Упрощенный процесс получения токена для дальнейшего прохождения авторизации.
// Ex. ?user_type=client ("client" или "moderator")
func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	const op = "authHandler.DummyLogin"

	userType := r.URL.Query().Get("user_type")
	if userType == "" {
		h.logger.Error("User type is missing", slog.String("op", op))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userType != "client" && userType != "moderator" {
		h.logger.Error("Invalid user type", slog.String("op", op), slog.String("user_type", userType))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.authService.GenerateToken("userID", userType)
	if err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Could not generate token", op, err)
		return
	}

	h.logger.Info("Token generated successfully in DummyLogin", slog.String("op", op), slog.String("user_type", userType))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "authHandler.Register"

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"user_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request", slog.String("op", op), "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Role != "client" && req.Role != "moderator" {
		h.logger.Error("Invalid user type", slog.String("op", op), slog.String("role", req.Role))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, repositories.ErrUserExists) {
			common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "User already exists", op, err)
			return
		}
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Could not register user", op, err)
		return
	}

	response := struct {
		ID string `json:"user_id"`
	}{
		ID: user,
	}

	h.logger.Info("User registered successfully", slog.String("op", op), slog.String("email", req.Email))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "authHandler.Login"

	var req struct {
		Id       string `json:"id"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request", slog.String("op", op), "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(r.Context(), req.Id, req.Password)
	if err != nil {
		h.logger.Error("User not found", slog.String("op", op), slog.String("id", req.Id), "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Role)
	if err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Could not generate token", op, err)
		return
	}

	h.logger.Info("User logged in successfully", slog.String("op", op), slog.String("id", req.Id))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		common.WriteErrorResponse(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", op, err)
	}
}

func (h *Handler) ValidateToken(tokenStr string) (*models.Claims, error) {
	return h.authService.ValidateToken(tokenStr)
}
