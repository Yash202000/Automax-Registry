package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/automax/registry/internal/models"
	"github.com/automax/registry/internal/repository"
	"github.com/automax/registry/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CredentialService struct {
	things repository.ThingRepository
	creds  repository.CredentialRepository
}

func NewCredentialService(things repository.ThingRepository, creds repository.CredentialRepository) *CredentialService {
	return &CredentialService{things: things, creds: creds}
}

func (s *CredentialService) Issue(ctx context.Context, thingID uuid.UUID) (*models.CredentialIssued, error) {
	if err := s.requireThing(ctx, thingID); err != nil {
		logCredentialErr("issue", thingID, uuid.Nil, err)
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		apiKey, err := randomToken(32)
		if err != nil {
			log.Printf("registry: credential issue failed thing_id=%s reason=random api_key err=%s", thingID, utils.Redact(err.Error()))
			return nil, err
		}
		apiSecret, err := randomToken(32)
		if err != nil {
			log.Printf("registry: credential issue failed thing_id=%s reason=random api_secret err=%s", thingID, utils.Redact(err.Error()))
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(apiSecret), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("registry: credential issue failed thing_id=%s reason=hash api_secret err=%s", thingID, utils.Redact(err.Error()))
			return nil, err
		}
		cred := &models.DeviceCredential{
			ThingID:       thingID,
			APIKey:        apiKey,
			APISecretHash: models.SecretHash(hash),
			Status:        "active",
		}
		if err := s.creds.Create(ctx, cred); err != nil {
			if repository.IsUniqueViolation(err) {
				lastErr = err
				log.Printf("registry: credential issue retry thing_id=%s attempt=%d reason=api_key collision", thingID, attempt+1)
				continue
			}
			log.Printf("registry: credential issue failed thing_id=%s api_key=%s err=%s", thingID, apiKey, utils.Redact(err.Error()))
			return nil, err
		}
		log.Printf("registry: credential issued thing_id=%s credential_id=%s api_key=%s status=active", thingID, cred.ID, apiKey)
		return &models.CredentialIssued{APIKey: apiKey, APISecret: apiSecret}, nil
	}
	log.Printf("registry: credential issue failed thing_id=%s reason=api_key collision retries exhausted err=%s", thingID, utils.Redact(lastErr.Error()))
	return nil, lastErr
}

func (s *CredentialService) List(ctx context.Context, thingID uuid.UUID) ([]models.CredentialListItem, error) {
	if err := s.requireThing(ctx, thingID); err != nil {
		logCredentialErr("list", thingID, uuid.Nil, err)
		return nil, err
	}
	items, err := s.creds.ListByThing(ctx, thingID)
	if err != nil {
		log.Printf("registry: credential list failed thing_id=%s err=%s", thingID, utils.Redact(err.Error()))
		return nil, err
	}
	log.Printf("registry: credential list thing_id=%s count=%d", thingID, len(items))
	return items, nil
}

func (s *CredentialService) Revoke(ctx context.Context, thingID, credentialID uuid.UUID) (*models.CredentialListItem, error) {
	if err := s.requireThing(ctx, thingID); err != nil {
		logCredentialErr("revoke", thingID, credentialID, err)
		return nil, err
	}
	revokedAt := time.Now().UTC()
	revoked, err := s.creds.Revoke(ctx, thingID, credentialID, revokedAt)
	if err != nil {
		log.Printf("registry: credential revoke failed thing_id=%s credential_id=%s err=%s", thingID, credentialID, utils.Redact(err.Error()))
		return nil, err
	}
	if !revoked {
		status, err := s.creds.Status(ctx, thingID, credentialID)
		if repository.IsNotFound(err) {
			log.Printf("registry: credential revoke rejected thing_id=%s credential_id=%s reason=not found", thingID, credentialID)
			return nil, NotFound("credential not found")
		}
		if err != nil {
			log.Printf("registry: credential revoke failed thing_id=%s credential_id=%s err=%s", thingID, credentialID, utils.Redact(err.Error()))
			return nil, err
		}
		if status == "revoked" {
			log.Printf("registry: credential revoke rejected thing_id=%s credential_id=%s reason=already revoked", thingID, credentialID)
			return nil, Conflict("credential is already revoked")
		}
		log.Printf("registry: credential revoke rejected thing_id=%s credential_id=%s reason=not found status=%s", thingID, credentialID, status)
		return nil, NotFound("credential not found")
	}
	items, err := s.creds.ListByThing(ctx, thingID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == credentialID {
			item := items[i]
			log.Printf("registry: credential revoked thing_id=%s credential_id=%s api_key=%s status=%s revoked_at=%s", thingID, item.ID, item.APIKey, item.Status, revokedAt.Format(time.RFC3339))
			return &item, nil
		}
	}
	log.Printf("registry: credential revoke failed thing_id=%s credential_id=%s reason=row missing after update", thingID, credentialID)
	return nil, NotFound("credential not found")
}

func (s *CredentialService) requireThing(ctx context.Context, thingID uuid.UUID) error {
	ok, err := s.things.Exists(ctx, thingID)
	if err != nil {
		return err
	}
	if !ok {
		return NotFound("thing not found")
	}
	return nil
}

func logCredentialErr(op string, thingID, credentialID uuid.UUID, err error) {
	var se *Error
	if errors.As(err, &se) {
		if credentialID == uuid.Nil {
			log.Printf("registry: credential %s rejected thing_id=%s reason=%s", op, thingID, se.Message)
			return
		}
		log.Printf("registry: credential %s rejected thing_id=%s credential_id=%s reason=%s", op, thingID, credentialID, se.Message)
		return
	}
	if credentialID == uuid.Nil {
		log.Printf("registry: credential %s failed thing_id=%s err=%s", op, thingID, utils.Redact(err.Error()))
		return
	}
	log.Printf("registry: credential %s failed thing_id=%s credential_id=%s err=%s", op, thingID, credentialID, utils.Redact(err.Error()))
}

func randomToken(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
