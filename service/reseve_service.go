package service

import (
	"context"
	"errors"
	"time"

	"reserv-service/models"
	"reserv-service/repository"
)

var (
	ErrItemAlreadyReserved = errors.New("item already reserved")
	ErrReserveNotFound     = errors.New("reserve not found")
	ErrInvalidReserve      = errors.New("invalid reserve")
)

type ReserveService struct {
	repo *repository.ReserveRepository
}

func NewReserveService(repo *repository.ReserveRepository) *ReserveService {
	return &ReserveService{repo: repo}
}

func (s *ReserveService) CreateReserve(ctx context.Context, req models.ReserveRequest) (*models.Reserve, error) {
	// Проверяем, не зарезервирован ли уже предмет
	existing, err := s.repo.FindActiveByItemID(ctx, req.IdItem)
	if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
		return nil, err
	}

	if existing != nil {
		return nil, ErrItemAlreadyReserved
	}

	// Создаем резерв
	reserve := &models.Reserve{
		IdItem:    req.IdItem,
		IdUser:    req.IdUser,
		ItemType:  req.ItemType,
		Status:    "pending",
		ExpiresAt: time.Now().Add(30 * time.Minute), // Резерв на 30 минут
	}

	err = s.repo.Create(ctx, reserve)
	if err != nil {
		return nil, err
	}

	return reserve, nil
}

func (s *ReserveService) GetUserReserves(ctx context.Context, userID int64, itemID int64) ([]models.Reserve, error) {
	return s.repo.FindByUserID(ctx, userID, itemID)
}

func (s *ReserveService) CancelReserve(ctx context.Context, reserveID int64) error {
	reserve, err := s.repo.FindByID(ctx, reserveID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return ErrReserveNotFound
		}
		return err
	}

	// Помечаем как отмененный
	reserve.Status = "cancelled"
	return s.repo.Update(ctx, reserve)
}

func (s *ReserveService) CompleteReserve(ctx context.Context, reserveID int64) error {
	reserve, err := s.repo.FindByID(ctx, reserveID)
	if err != nil {
		return err
	}

	// Помечаем как завершенный
	reserve.Status = "completed"
	return s.repo.Update(ctx, reserve)
}

// Проверяет и отменяет просроченные резервы
func (s *ReserveService) CleanupExpiredReserves(ctx context.Context) (int64, error) {
	expiredReserves, err := s.repo.FindExpired(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	for _, reserve := range expiredReserves {
		reserve.Status = "expired"
		if err := s.repo.Update(ctx, &reserve); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}
