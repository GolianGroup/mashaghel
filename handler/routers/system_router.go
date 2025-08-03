package routers

import (
	"mashaghel/handler/controllers"
	"net/http"
)

type SystemRouter interface {
	AddRoutes(mux *http.ServeMux)
}

type systemRouter struct {
	controller controllers.SystemController
}

func NewSystemRouter(controller controllers.SystemController) SystemRouter {
	return &systemRouter{controller: controller}
}

func (r *systemRouter) AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/system/health", r.controller.HealthCheck)
	mux.HandleFunc("/system/ready", r.controller.ReadyCheck)
}
