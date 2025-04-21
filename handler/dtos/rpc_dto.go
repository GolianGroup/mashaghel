package dto

import rpc_service "mashaghel/proto"

// HelloRequestDTO represents the domain model for hello request
type HelloRequestDTO struct {
	Name string
}

// HelloReplyDTO represents the domain model for hello response
type HelloReplyDTO struct {
	Message string
}

// ToHelloRequestDTO converts gRPC HelloRequest to domain DTO
func ToHelloRequestDTO(req *rpc_service.HelloRequest) *HelloRequestDTO {
	return &HelloRequestDTO{
		Name: req.Name,
	}
}

// ToHelloReply converts domain DTO to gRPC HelloReply
func (dto *HelloReplyDTO) ToHelloReply() *rpc_service.HelloReply {
	return &rpc_service.HelloReply{
		Message: dto.Message,
	}
}
