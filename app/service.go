package app

import (
	"mashaghel/internal/repositories"
	"mashaghel/internal/services"
)

func (a *application) InitServices(repository repositories.Repository) services.Service {
	return services.NewService(repository)
}
