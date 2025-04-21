package app

import (
	"mashaghel/handler/controllers"
	"mashaghel/internal/services"

	"go.uber.org/zap"
)

func (a *application) InitController(service services.Service, logger *zap.Logger) controllers.Controllers {
	return controllers.NewControllers(service, logger)
}
