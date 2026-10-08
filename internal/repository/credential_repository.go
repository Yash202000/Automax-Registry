package repository

import (
	"context"
	"time"

	"github.com/automax/registry/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CredentialRepository interface {
	Create(ctx context.Context, cred *models.DeviceCredential) error
	ListByThing(ctx context.Context, thingID uuid.UUID) ([]models.CredentialListItem, error)
	Revoke(ctx context.Context, thingID, id uuid.UUID, at time.Time) (bool, error)
	Status(ctx context.Context, thingID, id uuid.UUID) (string, error)
}

type credentialRepository struct{ db *gorm.DB }

func NewCredentialRepository(db *gorm.DB) CredentialRepository {
	return &credentialRepository{db: db}
}

func (r *credentialRepository) Create(ctx context.Context, cred *models.DeviceCredential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

func (r *credentialRepository) ListByThing(ctx context.Context, thingID uuid.UUID) ([]models.CredentialListItem, error) {
	var items []models.CredentialListItem
	err := r.db.WithContext(ctx).
		Model(&models.DeviceCredential{}).
		Select("id", "api_key", "status", "created_at", "revoked_at").
		Where("thing_id = ?", thingID).
		Order("created_at DESC, id DESC").
		Find(&items).Error
	if items == nil {
		items = []models.CredentialListItem{}
	}
	return items, err
}

func (r *credentialRepository) Revoke(ctx context.Context, thingID, id uuid.UUID, at time.Time) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&models.DeviceCredential{}).
		Where("id = ? AND thing_id = ? AND status = ?", id, thingID, "active").
		Updates(map[string]any{
			"status":     "revoked",
			"revoked_at": at,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *credentialRepository) Status(ctx context.Context, thingID, id uuid.UUID) (string, error) {
	var row struct {
		Status string
	}
	err := r.db.WithContext(ctx).
		Model(&models.DeviceCredential{}).
		Select("status").
		Where("id = ? AND thing_id = ?", id, thingID).
		Take(&row).Error
	return row.Status, err
}
