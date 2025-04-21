package app

import (
	"mashaghel/handler/controllers"
	rpc_service "mashaghel/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (a *application) InitGRPCServer(controller controllers.Controllers, logger *zap.Logger) *grpc.Server {
	grpcServer := grpc.NewServer()
	// Register server with controller
	rpc_service.RegisterRpcServiceServer(grpcServer, controller.RpcServiceController())

	if a.config.Environment == "development" {
		reflection.Register(grpcServer)
	}

	return grpcServer
}
