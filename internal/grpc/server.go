package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"upgraded-calculator/gen"
	"upgraded-calculator/internal/common"
	"upgraded-calculator/internal/config"
)

type serverAPI struct {
	gen.UnimplementedCalculatorServer
	calculator Calculator
}

type Calculator interface {
	ExecuteGRPC(
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
	resp, err := s.calculator.ExecuteGRPC(ctx, request)
	return resp, err
}

func CreateServer(
	config *config.Config,
	logger *slog.Logger,
) *grpc.Server {
	calculator := common.NewCalculatorFacade(logger)

	grpcServer := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{Timeout: config.App.GRPCTimeout}))
	RegisterGRPCServer(grpcServer, calculator)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		logger.Info("Initiating GRPC shutdown...")
		timer := time.AfterFunc(10*time.Second, func() {
			log.Println("Server couldn't stop gracefully in time. Doing force stop.")
			grpcServer.Stop()
		})
		defer timer.Stop()
		grpcServer.GracefulStop()
		logger.Info("Server stopped.")
	}()

	return grpcServer
}
