package houseService

import (
	"avito/internal/domain/models"
	"avito/internal/repositories/houseRepo"
	"context"
	"errors"
	"log/slog"
)

type HouseService interface {
	Create(ctx context.Context, address string, yearBuilt int, builder *string) (*models.House, error)
	Subscribe(ctx context.Context, houseID int, email string) error
	GetFlatsByHouseID(ctx context.Context, houseID int, role string) ([]models.Flat, error)
}

type Service struct {
	repo   houseRepo.HouseRepo
	logger *slog.Logger
}

var ErrValidation = errors.New("validation error")

func NewService(repo houseRepo.HouseRepo, logger *slog.Logger) HouseService {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) Create(ctx context.Context, address string, yearBuilt int, builder *string) (*models.House, error) {
	const op = "houseService.Create"

	newHouse := &models.House{
		Address:   address,
		YearBuilt: yearBuilt,
		Builder:   builder,
	}
	if err := s.repo.CreateHouse(ctx, newHouse); err != nil {
		s.logger.Error("Failed to create house", slog.String("op", op), "error", err)
		return nil, err
	}

	s.logger.Debug("House created successfully", slog.String("op", op), slog.Int("houseID", newHouse.ID))
	return newHouse, nil
}

func (s *Service) Subscribe(ctx context.Context, houseID int, email string) error {
	const op = "houseService.Subscribe"

	// Email validation
	if email == "" {
		s.logger.Error("Validation error: email is empty", slog.String("op", op))
		return ErrValidation
	}

	if err := s.repo.SubscribeToHouse(ctx, houseID, email); err != nil {
		s.logger.Error("Failed to subscribe to house", slog.String("op", op), "error", err, slog.Int("houseID", houseID), slog.String("email", email))
		return err
	}

	s.logger.Debug("User subscribed to house successfully", slog.String("op", op), slog.Int("houseID", houseID), slog.String("email", email))
	return nil
}

func (s *Service) GetFlatsByHouseID(ctx context.Context, houseID int, role string) ([]models.Flat, error) {
	const op = "houseService.GetFlatsByHouseID"

	flats, err := s.repo.GetFlatsByHouseID(ctx, houseID, role)
	if err != nil {
		s.logger.Error("Failed to get flats by house ID", slog.String("op", op), "error", err, slog.Int("houseID", houseID))
		return nil, err
	}

	s.logger.Debug("Returning all flats", slog.String("op", op), slog.Int("houseID", houseID))
	return flats, nil
}
