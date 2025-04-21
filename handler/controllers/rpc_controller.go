package controllers

import (
	"context"
	dto "mashaghel/handler/dtos"
	"mashaghel/internal/services"

	"go.uber.org/zap"

	rpc_service "mashaghel/proto"
)

type RpcServiceController interface {
	SayHello(ctx context.Context, req *rpc_service.HelloRequest) (*rpc_service.HelloReply, error)
}

type rpcServiceController struct {
	rpcServiceService services.RpcServiceService
}

func NewRpcServiceController(service services.RpcServiceService, logger *zap.Logger) RpcServiceController {
	return &rpcServiceController{
		rpcServiceService: service,
	}
}

func (c *rpcServiceController) SayHello(ctx context.Context, req *rpc_service.HelloRequest) (*rpc_service.HelloReply, error) {
	requestDTO := dto.ToHelloRequestDTO(req)

	responseDTO, err := c.rpcServiceService.SayHello(ctx, requestDTO)
	if err != nil {
		return nil, err
	}

	return responseDTO.ToHelloReply(), nil
}
