package repository

import (
	"context"

	"github.com/automax/registry/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChannelRepository interface {
	Create(ctx context.Context, channel *models.Channel) error
	Update(ctx context.Context, channel *models.Channel) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Channel, error)
	List(ctx context.Context) ([]models.Channel, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type channelRepository struct{ db *gorm.DB }

func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(ctx context.Context, channel *models.Channel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

func (r *channelRepository) Update(ctx context.Context, channel *models.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

func (r *channelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.Channel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *channelRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	var channel models.Channel
	err := r.db.WithContext(ctx).First(&channel, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *channelRepository) List(ctx context.Context) ([]models.Channel, error) {
	var channels []models.Channel
	err := r.db.WithContext(ctx).Order("created_at ASC, id ASC").Find(&channels).Error
	return channels, err
}

func (r *channelRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Channel{}).Where("id = ?", id).Count(&n).Error
	return n > 0, err
}
