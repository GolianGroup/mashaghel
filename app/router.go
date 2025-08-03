package app

import (
	"mashaghel/handler/controllers"
	"mashaghel/handler/routers"

	"go.uber.org/zap"
)

func (a *application) InitRouter(logger *zap.Logger, controllers controllers.Controllers) routers.Router {
	return routers.NewRouter(controllers)
}
