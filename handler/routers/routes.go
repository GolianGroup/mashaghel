package routers

import (
	"mashaghel/handler/controllers"
	"net/http"
)

type Router interface {
	AddRoutes(mux *http.ServeMux)
}

type router struct {
	systemRouter SystemRouter
	// redisClient  producers.RedisClient
	// tracer       trace.Tracer
}

func NewRouter(controllers controllers.Controllers) Router {

	return &router{
		systemRouter: NewSystemRouter(controllers.SystemController()),
		// redisClient: redisClient,
		// tracer:      tracer,
	}
}

func (r router) AddRoutes(mux *http.ServeMux) {

	// router
	// init user router, etc ...
	// rate limiter
	// CORS

	r.systemRouter.AddRoutes(mux)

}
