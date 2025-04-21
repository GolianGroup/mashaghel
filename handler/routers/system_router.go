package routers

import (
	"mashaghel/handler/controllers"

	"github.com/gofiber/fiber/v2"
)

type SystemRouter interface {
	AddRoutes(router fiber.Router)
}

type systemRouter struct {
	Controller controllers.SystemController
}

func NewSystemRouter(controller controllers.SystemController) SystemRouter {
	return &systemRouter{Controller: controller}
}

func (r *systemRouter) AddRoutes(router fiber.Router) {
	router.Get("/api/health", r.Controller.HealthCheck)
	router.Get("/health/ready", r.Controller.ReadyCheck)
}
