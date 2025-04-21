package services

import (
	"context"
	dto "mashaghel/handler/dtos"
)

type RpcServiceService interface {
	SayHello(ctx context.Context, req *dto.HelloRequestDTO) (*dto.HelloReplyDTO, error)
}

type rpcServiceService struct{}

func NewRpcServiceService() RpcServiceService { // repo repositories.ExampleRepository
	return &rpcServiceService{}
}

func (s *rpcServiceService) SayHello(ctx context.Context, req *dto.HelloRequestDTO) (*dto.HelloReplyDTO, error) {
	message := "Hello, " + req.Name
	return &dto.HelloReplyDTO{
		Message: message,
	}, nil
}
