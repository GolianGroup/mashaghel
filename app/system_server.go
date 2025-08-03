package app

import (
	"mashaghel/handler/routers"
	"mashaghel/handler/server"

	"go.uber.org/zap"
)

func (a *application) InitSystemServer(logger *zap.Logger, router routers.Router) server.HttpServer {
	srv := server.NewHttpServer(*a.config, logger)

	mux := srv.Mux()
	router.AddRoutes(mux)

	return srv
}
