package authService

import (
	"avito/internal/domain/models"
	"avito/internal/repositories/authRepo"

	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, email, password, role string) (string, error)
	Login(ctx context.Context, email, password string) (*models.User, error)
	GenerateToken(userID string, role string) (string, error)
	ValidateToken(tokenStr string) (*models.Claims, error)
}

type Service struct {
	repo      authRepo.AuthRepo
	jwtSecret string
	logger    *slog.Logger
}

func NewService(repo authRepo.AuthRepo, jwtSecret string, logger *slog.Logger) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

func (s *Service) Register(ctx context.Context, email, password, role string) (string, error) {
	const op = "authService.Register"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Password hashing error", slog.String("op", op), "error", err)
		return "", err
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error("Error creating user", slog.String("op", op), "error", err)
		return "", err
	}

	return userID, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*models.User, error) {
	const op = "authService.Login"

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Error getting user by email", slog.String("op", op), "error", err)
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Error("Incorrect credentials", slog.String("op", op), slog.String("email", email))
		return nil, errors.New("invalid credentials")
	}

	s.logger.Debug("Successful login", slog.String("op", op), slog.String("user_id", user.ID))

	return user, nil
}

func (s *Service) GenerateToken(userID string, role string) (string, error) {
	const op = "authService.GenerateToken"

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error("Error generating JWT token", slog.String("op", op), "error", err)
		return "", err
	}

	s.logger.Debug("JWT token successfully generated", slog.String("op", op))
	return signedToken, nil
}

func (s *Service) ValidateToken(tokenStr string) (*models.Claims, error) {
	const op = "authService.ValidateToken"

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		s.logger.Error("JWT parsing error", slog.String("op", op), "error", err)
		return nil, err
	}
	if !token.Valid {
		s.logger.Error("invalid JWT", slog.String("op", op))
		return nil, errors.New("invalid token")
	}

	s.logger.Debug("JWT validated", slog.String("op", op), slog.String("user_id", claims.UserID))

	return claims, nil
}
