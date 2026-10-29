package repository

import (
	"context"
	"errors"

	"github.com/automax/registry/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrParentMissing = errors.New("parent group not found")

type InUseError struct {
	Children int64
	Members  int64
}

func (e *InUseError) Error() string { return "group is in use" }

type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	Update(ctx context.Context, group *models.Group) error
	DeleteIfUnused(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Group, error)
	List(ctx context.Context, filter models.GroupListFilter) ([]models.Group, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	CreatesCycle(ctx context.Context, groupID, parentID uuid.UUID) (bool, error)
}

type groupRepository struct{ db *gorm.DB }

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *groupRepository) Update(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteIfUnused locks the group, rejects the delete when a child group or a
// member thing still points at it, and otherwise deletes the row. The foreign
// keys are ON DELETE RESTRICT, so a concurrent insert still cannot orphan rows.
func (r *groupRepository) DeleteIfUnused(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var group models.Group
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&group, "id = ?", id).Error; err != nil {
			return err
		}
		var children, members int64
		if err := tx.Model(&models.Group{}).Where("parent_group_id = ?", id).Count(&children).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Thing{}).Where("group_id = ?", id).Count(&members).Error; err != nil {
			return err
		}
		if children > 0 || members > 0 {
			return &InUseError{Children: children, Members: members}
		}
		res := tx.Delete(&models.Group{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *groupRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Group, error) {
	var group models.Group
	err := r.db.WithContext(ctx).First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) List(ctx context.Context, filter models.GroupListFilter) ([]models.Group, error) {
	q := r.db.WithContext(ctx).Model(&models.Group{})
	if filter.RootsOnly {
		q = q.Where("parent_group_id IS NULL")
	} else if filter.ParentGroupID != nil {
		q = q.Where("parent_group_id = ?", *filter.ParentGroupID)
	}
	var groups []models.Group
	err := q.Order("name ASC, id ASC").Find(&groups).Error
	return groups, err
}

func (r *groupRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Group{}).Where("id = ?", id).Count(&n).Error
	return n > 0, err
}

// CreatesCycle reports whether setting parentID on groupID would loop.
// groupID may be uuid.Nil when the group has not been inserted yet.
func (r *groupRepository) CreatesCycle(ctx context.Context, groupID, parentID uuid.UUID) (bool, error) {
	current := parentID
	seen := map[uuid.UUID]struct{}{}
	for {
		if groupID != uuid.Nil && current == groupID {
			return true, nil
		}
		if _, ok := seen[current]; ok {
			return true, nil
		}
		seen[current] = struct{}{}

		var row struct {
			ParentGroupID *uuid.UUID
		}
		err := r.db.WithContext(ctx).Model(&models.Group{}).Select("parent_group_id").Where("id = ?", current).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrParentMissing
		}
		if err != nil {
			return false, err
		}
		if row.ParentGroupID == nil {
			return false, nil
		}
		current = *row.ParentGroupID
	}
}
