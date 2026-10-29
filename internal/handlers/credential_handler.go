package handlers

import (
	"log"

	"github.com/automax/registry/internal/services"
	"github.com/automax/registry/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type CredentialHandler struct {
	svc *services.CredentialService
}

func NewCredentialHandler(svc *services.CredentialService) *CredentialHandler {
	return &CredentialHandler{svc: svc}
}

func (h *CredentialHandler) Issue(c *fiber.Ctx) error {
	thingID, err := pathID(c, "id")
	if err != nil {
		log.Printf("registry: credential issue rejected reason=invalid thing id value=%q", c.Params("id"))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid id")
	}
	issued, err := h.svc.Issue(c.UserContext(), thingID)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "credential issued", issued)
}

func (h *CredentialHandler) List(c *fiber.Ctx) error {
	thingID, err := pathID(c, "id")
	if err != nil {
		log.Printf("registry: credential list rejected reason=invalid thing id value=%q", c.Params("id"))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.List(c.UserContext(), thingID)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "credentials retrieved", items)
}

func (h *CredentialHandler) Revoke(c *fiber.Ctx) error {
	thingID, err := pathID(c, "id")
	if err != nil {
		log.Printf("registry: credential revoke rejected reason=invalid thing id value=%q", c.Params("id"))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid id")
	}
	credentialID, err := pathID(c, "credentialId")
	if err != nil {
		log.Printf("registry: credential revoke rejected thing_id=%s reason=invalid credential id value=%q", thingID, c.Params("credentialId"))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid id")
	}
	item, err := h.svc.Revoke(c.UserContext(), thingID, credentialID)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "credential revoked", item)
}
