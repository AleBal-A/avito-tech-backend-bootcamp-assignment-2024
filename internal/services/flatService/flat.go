package flatService

import (
	"avito/internal/domain/models"
	"avito/internal/repositories/flatRepo"
	"context"
	"log/slog"
)

type FlatService interface {
	Create(ctx context.Context, houseID int, flatNumber *int, price, rooms int) (*models.Flat, error)
	UpdateStatus(ctx context.Context, flatID int, status string) (*models.Flat, error)
	GetFlatsByHouseID(ctx context.Context, houseID int) ([]models.Flat, error)
}

type Service struct {
	repo        flatRepo.FlatRepo
	isModerator func(ctx context.Context) bool
	logger      *slog.Logger
}

func NewService(repo flatRepo.FlatRepo, logger *slog.Logger) *Service {
	return &Service{
		repo:        repo,
		isModerator: isModerator,
		logger:      logger,
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

func (s *Service) UpdateStatus(ctx context.Context, flatID int, status string) (*models.Flat, error) {
	const op = "flatService.UpdateStatus"

	flat, err := s.repo.UpdateFlatStatus(ctx, flatID, status)
	if err != nil {
		s.logger.Error("Failed to update flat status", slog.String("op", op), slog.Int("flatID", flatID), slog.String("status", status))
		return nil, err
	}

	s.logger.Info("Flat status updated successfully", slog.String("op", op), slog.Int("flatID", flat.ID), slog.String("status", flat.Status))
	return flat, nil
}

// TODO: прописать в мидлваре только для модеров
func (s *Service) GetFlatsByHouseID(ctx context.Context, houseID int) ([]models.Flat, error) {
	const op = "flatService.GetFlatsByHouseID"

	flats, err := s.repo.GetFlatsByHouseID(ctx, houseID)
	if err != nil {
		s.logger.Error("Failed to get flats by house ID",
			slog.String("op", op), "error", err,
			slog.Int("houseID", houseID),
		)
		return nil, err
	}

	isMod := s.isModerator(ctx)

	var result []models.Flat
	for _, flat := range flats {
		// TODO: поставить фильтрацию на уровне Repo
		if isMod || flat.Status == "approved" {
			result = append(result, *flat)
		}
	}

	return result, nil
}

func isModerator(ctx context.Context) bool {
	claims, ok := ctx.Value("user").(*models.Claims)
	if !ok || claims == nil {
		return false
	}

	return claims.Role == "moderator"
}
