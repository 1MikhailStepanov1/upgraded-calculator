package grpc

import (
	"context"
	"google.golang.org/grpc"
	"upgraded-calculator/gen"
)

type serverAPI struct {
	gen.UnimplementedCalculatorServer
	calculator Calculator
}

type Calculator interface {
	Execute(
		ctx context.Context,
		request gen.Request,
	) (response gen.Response, err error)
}

func RegisterGRPCServer(server *grpc.Server, calculator Calculator) {
	gen.RegisterCalculatorServer(server, &serverAPI{calculator: calculator})
}

func (s *serverAPI) Execute(
	ctx context.Context,
	request *gen.Request,
) (response *gen.Response, err error) {
	return
}
