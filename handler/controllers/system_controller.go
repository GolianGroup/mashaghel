package controllers

import (
	"time"

	"mashaghel/internal/services"

	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2"
)

type SystemController interface {
	HealthCheck(c *fiber.Ctx) error
	ReadyCheck(c *fiber.Ctx) error
}

type systemController struct {
	systemService services.SystemService
	logger        *zap.Logger
}

func NewSystemController(systemService services.SystemService, logger *zap.Logger) SystemController {
	return &systemController{systemService: systemService, logger: logger}
}

func (controller *systemController) HealthCheck(c *fiber.Ctx) error {
	controller.logger.Info("HealthCheck")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "UP",
		"time":   time.Now(),
	})
}

func (controller *systemController) ReadyCheck(c *fiber.Ctx) error {

	controller.logger.Info("ReadyCheck")
	readyCheck, errors := controller.systemService.ReadyCheck(c.Context())
	status := "READY"
	if errors != nil {
		status = "NOT_READY"
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":     status,
		"readyCheck": readyCheck,
		"time":       time.Now(),
	})
}
