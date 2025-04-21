package controllers

import (
	"mashaghel/internal/services"

	"go.uber.org/zap"
)

type Controllers interface {
	RpcServiceController() RpcServiceController
	SystemController() SystemController
}

type controllers struct {
	rpcServiceController RpcServiceController
	systemController     SystemController
}

func NewControllers(s services.Service, logger *zap.Logger) Controllers {
	rpcServiceController := NewRpcServiceController(s.RpcServiceService(), logger)
	systemController := NewSystemController(s.SystemService(), logger)
	return &controllers{

		rpcServiceController: rpcServiceController,
		systemController:     systemController,
	}
}

func (c *controllers) RpcServiceController() RpcServiceController {
	return c.rpcServiceController
}

func (c *controllers) SystemController() SystemController {
	return c.systemController
}
