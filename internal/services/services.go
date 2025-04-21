package services

import "mashaghel/internal/repositories"

type Service interface {
	RpcServiceService() RpcServiceService
	SystemService() SystemService
}

type service struct {
	rpcServiceService RpcServiceService
	systemService     SystemService
}

func NewService(repo repositories.Repository) Service {
	rpcServiceService := NewRpcServiceService()
	systemService := NewSystemService(repo.SystemRepository())
	return &service{
		rpcServiceService: rpcServiceService,
		systemService:     systemService,
	}
}

func (s *service) RpcServiceService() RpcServiceService {
	return s.rpcServiceService
}

func (s *service) SystemService() SystemService {
	return s.systemService
}
