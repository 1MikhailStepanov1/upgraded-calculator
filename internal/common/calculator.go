package common

import (
	"context"
	"encoding/json"
	"log/slog"
	"upgraded-calculator/gen"
)

// HTTP обвязка
type CalculatorHTTPHandler struct {
	logger *slog.Logger
}

func (a *CalculatorHTTPHandler) Execute(ctx context.Context, data []byte) ([]byte, error) {
	a.logger.Info("Processing HTTP request")

	var req gen.Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	resp, err := execute(ctx, &req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp)
}

// GRPC обвязка
type CalculatorGRPCHandler struct {
	logger *slog.Logger
}

func (c *CalculatorGRPCHandler) Execute(ctx context.Context, req *gen.Request) ([]byte, error) {
	c.logger.Info("Processing request with request_id", "request_id", ctx.Value("request_id"))

	return []byte{}, nil
}

// Общий фасад калькулятора
type CalculatorFacade struct {
	logger      *slog.Logger
	httpHandler CalculatorHTTPHandler
	grpcHandler CalculatorGRPCHandler
}

func NewCalculatorFacade(logger *slog.Logger) *CalculatorFacade {
	return &CalculatorFacade{
		logger:      logger,
		httpHandler: CalculatorHTTPHandler{logger: logger},
		grpcHandler: CalculatorGRPCHandler{logger: logger},
	}
}

func (c *CalculatorFacade) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	switch data := input.(type) {
	case []byte:
		return c.httpHandler.Execute(ctx, data)
	case *gen.Request:
		return c.grpcHandler.Execute(ctx, data)
	default:
		return nil, nil
	}
}

// Внутренняя функция, реализующая основную бизнес логику
func execute(ctx context.Context, req *gen.Request) ([]byte, error) {
	// business logic of adapter
	// Lazy init - делать мапу переменных в рамках "адаптера".
	// В рамках мапы хранятся ссылки на память, где лежат значения переменных
	// Сделать одну операцию подсчета, которая будет складывать значения по ссылкам
	// И вторая операция - формирование ответа по порядку вызовов print, разыменовывая ссылки
	return []byte{}, nil
}
