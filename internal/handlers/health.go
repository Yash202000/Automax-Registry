package handlers

import (
	"context"
	"time"

	"github.com/automax/registry/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthTargets are the process connections reported by GET /health.
type HealthTargets struct {
	DB            *gorm.DB
	Redis         *redis.Client
	NATSConnected func() bool
}

func (h HealthTargets) handle(c *fiber.Ctx) error {
	parent := c.UserContext()
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	return utils.SuccessResponse(c, fiber.StatusOK, "ok", fiber.Map{
		"db":    pingDB(ctx, h.DB),
		"redis": pingRedis(ctx, h.Redis),
		"nats":  status(h.NATSConnected != nil && h.NATSConnected()),
	})
}

func pingDB(ctx context.Context, db *gorm.DB) string {
	if db == nil {
		return "down"
	}
	sqlDB, err := db.DB()
	if err != nil {
		return "down"
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return "down"
	}
	return "up"
}

func pingRedis(ctx context.Context, client *redis.Client) string {
	if client == nil {
		return "down"
	}
	if err := client.Ping(ctx).Err(); err != nil {
		return "down"
	}
	return "up"
}

func status(up bool) string {
	if up {
		return "up"
	}
	return "down"
}
