package flatService

import (
	"avito/internal/domain/models"
	"avito/internal/repositories/flatRepo"
	"context"
	"errors"
	"log/slog"
)

type FlatService interface {
	Create(ctx context.Context, houseID int, flatNumber *int, price, rooms int) (*models.Flat, error)
	UpdateStatus(ctx context.Context, flatID int, newStatus string, moderatorID string) (*models.Flat, error)
}

type Service struct {
	repo        flatRepo.FlatRepo
	isModerator func(ctx context.Context) bool
	logger      *slog.Logger
}

var ErrFlatBeingModerated = errors.New("flat is already being moderated by another user")

func NewService(repo flatRepo.FlatRepo, logger *slog.Logger) FlatService {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, houseID int, flatNumber *int, price, rooms int) (*models.Flat, error) {
	const op = "flatService.Create"

	newFlat := &models.Flat{
		HouseID:    houseID,
		FlatNumber: flatNumber,
		Price:      price,
		Rooms:      rooms,
		Status:     "created",
	}

	var err error
	newFlat.ID, err = s.repo.CreateFlat(ctx, newFlat)
	if err != nil {
		s.logger.Error("Failed to create flat", slog.String("op", op), "error", err)
		return nil, err
	}

	s.logger.Debug("Flat created successfully", slog.String("op", op), slog.Int("houseID", houseID))
	if flatNumber != nil {
		s.logger.Debug("Flat number assigned", slog.Int("flatNumber", *flatNumber))
	} else {
		s.logger.Debug("Flat number is not assigned")
	}
	return newFlat, nil
}

func (s *Service) UpdateStatus(ctx context.Context, flatID int, newStatus string, moderatorID string) (*models.Flat, error) {
	const op = "flatService.UpdateStatus"

	flat, err := s.repo.GetFlatByID(ctx, flatID)
	if err != nil {
		s.logger.Error("Failed to retrieve flat", slog.String("op", op), "error", err)
		return nil, err
	}

	if flat.Status == "on moderation" && (flat.ModeratorID == nil || *flat.ModeratorID != moderatorID) {
		s.logger.Error("Flat is already being moderated by another user", slog.String("op", op))
		return nil, ErrFlatBeingModerated
	}

	flat.Status = newStatus
	if newStatus == "on moderation" {
		flat.ModeratorID = &moderatorID
	} else {
		flat.ModeratorID = nil
	}

	updatedFlat, err := s.repo.UpdateFlatStatus(ctx, flatID, newStatus, flat.ModeratorID)
	if err != nil {
		s.logger.Error("Failed to update flat status", slog.String("op", op), "error", err)
		return nil, err
	}

	s.logger.Debug("Flat status updated successfully", slog.String("op", op), slog.Int("flatID", flatID))
	return updatedFlat, nil
}
