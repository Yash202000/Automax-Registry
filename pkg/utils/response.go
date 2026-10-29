package utils

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error:   message,
	})
}

func InternalErrorResponse(c *fiber.Ctx, err error, userMessage string) error {
	log.Printf("registry: %s %s: %s", c.Method(), c.Path(), Redact(err.Error()))
	return ErrorResponse(c, fiber.StatusInternalServerError, userMessage)
}
