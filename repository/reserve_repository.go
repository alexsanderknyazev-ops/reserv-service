package repository

import (
	"context"
	"errors"
	"time"

	"reserv-service/models"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type ReserveRepository struct {
	db *gorm.DB
}

func NewReserveRepository(db *gorm.DB) *ReserveRepository {
	return &ReserveRepository{db: db}
}

func (r *ReserveRepository) Create(ctx context.Context, reserve *models.Reserve) error {
	return r.db.WithContext(ctx).Create(reserve).Error
}

func (r *ReserveRepository) FindByID(ctx context.Context, id int64) (*models.Reserve, error) {
	var reserve models.Reserve
	err := r.db.WithContext(ctx).First(&reserve, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &reserve, err
}

func (r *ReserveRepository) FindByUserID(ctx context.Context, userID int64) ([]models.Reserve, error) {
	var reserves []models.Reserve
	err := r.db.WithContext(ctx).
		Where("id_user = ?", userID).
		Order("created_at DESC").
		Find(&reserves).Error
	return reserves, err
}

func (r *ReserveRepository) FindActiveByItemID(ctx context.Context, itemID int64) (*models.Reserve, error) {
	var reserve models.Reserve
	err := r.db.WithContext(ctx).
		Where("id_item = ? AND status = ? AND expires_at > ?",
			itemID, "pending", time.Now()).
		First(&reserve).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &reserve, err
}

func (r *ReserveRepository) Update(ctx context.Context, reserve *models.Reserve) error {
	return r.db.WithContext(ctx).Save(reserve).Error
}

func (r *ReserveRepository) FindExpired(ctx context.Context) ([]models.Reserve, error) {
	var reserves []models.Reserve
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at <= ?", "pending", time.Now()).
		Find(&reserves).Error
	return reserves, err
}
