package services

import (
	"context"

	"github.com/automax/registry/internal/models"
	"github.com/automax/registry/internal/repository"
	"github.com/google/uuid"
)

type ACLService struct {
	acls     repository.ACLRepository
	channels repository.ChannelRepository
	things   repository.ThingRepository
}

func NewACLService(acls repository.ACLRepository, channels repository.ChannelRepository, things repository.ThingRepository) *ACLService {
	return &ACLService{acls: acls, channels: channels, things: things}
}

func (s *ACLService) Create(ctx context.Context, channelID uuid.UUID, req models.ChannelACLCreateRequest) (*models.ChannelACLResponse, error) {
	if err := s.requireChannel(ctx, channelID); err != nil {
		return nil, err
	}
	if req.ThingID == uuid.Nil {
		return nil, Invalid("validation failed", map[string]string{"thing_id": "thing_id is required"})
	}
	ok, err := s.things.Exists(ctx, req.ThingID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, Invalid("validation failed", map[string]string{"thing_id": "thing not found"})
	}

	acl := &models.ChannelACL{
		ChannelID:    channelID,
		ThingID:      req.ThingID,
		CanPublish:   req.CanPublish,
		CanSubscribe: req.CanSubscribe,
	}
	if err := s.acls.Create(ctx, acl); err != nil {
		if repository.IsUniqueViolation(err) {
			return nil, Conflict("acl entry already exists for this thing and channel")
		}
		if repository.IsForeignKeyViolation(err) {
			return nil, Invalid("validation failed", map[string]string{"thing_id": "thing not found"})
		}
		return nil, err
	}
	resp := acl.Response()
	return &resp, nil
}

func (s *ACLService) Update(ctx context.Context, channelID, aclID uuid.UUID, req models.ChannelACLUpdateRequest) (*models.ChannelACLResponse, error) {
	if err := s.requireChannel(ctx, channelID); err != nil {
		return nil, err
	}
	acl, err := s.acls.Find(ctx, channelID, aclID)
	if repository.IsNotFound(err) {
		return nil, NotFound("acl entry not found")
	}
	if err != nil {
		return nil, err
	}
	acl.CanPublish = req.CanPublish
	acl.CanSubscribe = req.CanSubscribe
	if err := s.acls.Update(ctx, acl); err != nil {
		return nil, err
	}
	resp := acl.Response()
	return &resp, nil
}

func (s *ACLService) Delete(ctx context.Context, channelID, aclID uuid.UUID) error {
	if err := s.requireChannel(ctx, channelID); err != nil {
		return err
	}
	if err := s.acls.Delete(ctx, channelID, aclID); err != nil {
		if repository.IsNotFound(err) {
			return NotFound("acl entry not found")
		}
		return err
	}
	return nil
}

func (s *ACLService) List(ctx context.Context, channelID uuid.UUID) ([]models.ChannelACLResponse, error) {
	if err := s.requireChannel(ctx, channelID); err != nil {
		return nil, err
	}
	rows, err := s.acls.ListByChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	out := make([]models.ChannelACLResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Response())
	}
	return out, nil
}

func (s *ACLService) requireChannel(ctx context.Context, id uuid.UUID) error {
	ok, err := s.channels.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return NotFound("channel not found")
	}
	return nil
}
