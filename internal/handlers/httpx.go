package handlers

import (
	"errors"

	"github.com/automax/registry/internal/services"
	"github.com/automax/registry/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func writeServiceErr(c *fiber.Ctx, err error) error {
	var se *services.Error
	if errors.As(err, &se) {
		switch se.Kind {
		case services.KindInvalid:
			if len(se.Fields) > 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   se.Message,
					"errors":  se.Fields,
				})
			}
			return utils.ErrorResponse(c, fiber.StatusBadRequest, se.Message)
		case services.KindNotFound:
			return utils.ErrorResponse(c, fiber.StatusNotFound, se.Message)
		case services.KindConflict:
			return utils.ErrorResponse(c, fiber.StatusConflict, se.Message)
		}
	}
	return utils.InternalErrorResponse(c, err, "internal server error")
}

func pathID(c *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
