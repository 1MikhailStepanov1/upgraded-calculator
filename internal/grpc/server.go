package grpc

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log/slog"
	"upgraded-calculator/gen"
	"upgraded-calculator/internal/config"
)

type serverAPI struct {
	gen.UnimplementedCalculatorServer
	calculator Calculator
}

type Calculator interface {
	Execute(
		ctx context.Context,
		request *gen.Request,
	) (response *gen.Response, err error)
}

func RegisterGRPCServer(server *grpc.Server, calculator Calculator) {
	gen.RegisterCalculatorServer(server, &serverAPI{calculator: calculator})
}

func (s *serverAPI) Execute(
	ctx context.Context,
	request *gen.Request,
) (response *gen.Response, err error) {
	ctx = context.WithValue(ctx, "request_id", uuid.New().String())
	resp, err := s.calculator.Execute(ctx, request)
	return resp, err
}

func CreateServer(
	config *config.Config,
	logger *slog.Logger,
) *grpc.Server {
	calculator := &CalculatorGRPC{logger: logger}

	grpcServer := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{Timeout: config.App.GRPCTimeout}))
	RegisterGRPCServer(grpcServer, calculator)

	return grpcServer
}
