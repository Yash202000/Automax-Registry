package repository

import (
	"context"

	"github.com/automax/registry/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ACLRepository interface {
	Create(ctx context.Context, acl *models.ChannelACL) error
	Update(ctx context.Context, acl *models.ChannelACL) error
	Delete(ctx context.Context, channelID, aclID uuid.UUID) error
	Find(ctx context.Context, channelID, aclID uuid.UUID) (*models.ChannelACL, error)
	ListByChannel(ctx context.Context, channelID uuid.UUID) ([]models.ChannelACL, error)
}

type aclRepository struct{ db *gorm.DB }

func NewACLRepository(db *gorm.DB) ACLRepository {
	return &aclRepository{db: db}
}

func (r *aclRepository) Create(ctx context.Context, acl *models.ChannelACL) error {
	return r.db.WithContext(ctx).Create(acl).Error
}

func (r *aclRepository) Update(ctx context.Context, acl *models.ChannelACL) error {
	return r.db.WithContext(ctx).Save(acl).Error
}

func (r *aclRepository) Delete(ctx context.Context, channelID, aclID uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.ChannelACL{}, "id = ? AND channel_id = ?", aclID, channelID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *aclRepository) Find(ctx context.Context, channelID, aclID uuid.UUID) (*models.ChannelACL, error) {
	var acl models.ChannelACL
	err := r.db.WithContext(ctx).First(&acl, "id = ? AND channel_id = ?", aclID, channelID).Error
	if err != nil {
		return nil, err
	}
	return &acl, nil
}

func (r *aclRepository) ListByChannel(ctx context.Context, channelID uuid.UUID) ([]models.ChannelACL, error) {
	var rows []models.ChannelACL
	err := r.db.WithContext(ctx).Where("channel_id = ?", channelID).Order("created_at ASC, id ASC").Find(&rows).Error
	return rows, err
}
