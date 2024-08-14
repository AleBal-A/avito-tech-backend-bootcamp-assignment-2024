package authHandler

import (
	"avito/internal/domain/models"
	"avito/internal/repositories"
	"avito/internal/services/authService"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
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

func NewHandler(authService authService.AuthService, logger *slog.Logger) *Handler {
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
		http.Error(w, "User type is required", http.StatusBadRequest)
		return
	}

	if userType != "client" && userType != "moderator" {
		h.logger.Error("Invalid user type", slog.String("op", op), slog.String("user_type", userType))
		http.Error(w, "Invalid user type. Allowed values are 'client' or 'moderator'", http.StatusBadRequest)
		return
	}

	token, err := h.authService.GenerateToken("userID", userType)
	if err != nil {
		h.logger.Error("Could not generate token", slog.String("op", op), "error", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Token generated successfully in DummyLogin", slog.String("op", op), slog.String("user_type", userType))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		h.logger.Error("Failed to write response", slog.String("op", op), "error", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "authHandler.Register"

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request in Register", slog.String("op", op), "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Role != "client" && req.Role != "moderator" {
		h.logger.Error("Invalid user type", slog.String("op", op), slog.String("user_type", req.Role))
		http.Error(w, "Invalid user type. Allowed values are 'client' or 'moderator'", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, repositories.ErrUserExists) {
			h.logger.Warn("User already exists", slog.String("op", op), slog.String("email", req.Email))
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		h.logger.Error("Could not register user", slog.String("op", op), "error", err)
		http.Error(w, "Could not register user", http.StatusInternalServerError)
		return
	}

	response := struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}{
		ID:    user,
		Email: req.Email,
		Role:  req.Role,
	}

	h.logger.Info("User registered successfully", slog.String("op", op), slog.String("email", req.Email))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write response", slog.String("op", op), "error", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "authHandler.Login"

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request in Login", slog.String("op", op), "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("Could not login user", slog.String("op", op), "error", err)
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Role)
	if err != nil {
		h.logger.Error("Could not generate token in Login", slog.String("op", op), "error", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User logged in successfully", slog.String("op", op), slog.String("email", req.Email))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		h.logger.Error("Failed to write response", slog.String("op", op), "error", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (h *Handler) ValidateToken(tokenStr string) (*models.Claims, error) {
	return h.authService.ValidateToken(tokenStr)
}
