package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"upgraded-calculator/gen"
)

// HTTP обвязка
type CalculatorHTTPHandler struct {
	logger *slog.Logger
}

func (a *CalculatorHTTPHandler) Execute(ctx context.Context, data []byte) ([]byte, error) {
	a.logger.Info("Processing HTTP request")
	c := Calculator{logger: a.logger}

	var req []Operation
	err := json.Unmarshal(data, &req)
	if err != nil {
		a.logger.Error(err.Error())
		return nil, err
	}
	for _, op := range req {
		a.logger.Debug(fmt.Sprintf("Deserialized operation: %+v", op))
	}

	resp, err := c.Execute(ctx, req)
	if err != nil {
		a.logger.Error(err.Error())
		return nil, err
	}
	a.logger.Info("Request finished")
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

func (c *CalculatorFacade) Execute(ctx context.Context, input interface{}) ([]byte, error) {
	switch data := input.(type) {
	case []byte:
		return c.httpHandler.Execute(ctx, data)
	case *gen.Request:
		return c.grpcHandler.Execute(ctx, data)
	default:
		return nil, nil
	}
}

type Calculator struct {
	logger *slog.Logger
}

func (c *Calculator) Execute(ctx context.Context, operations []Operation) ([]PrintOutput, error) {
	// business logic of adapter
	// Lazy init - делать мапу переменных в рамках "адаптера".
	// В рамках мапы хранятся ссылки на память, где лежат значения переменных
	// Сделать одну операцию подсчета, которая будет складывать значения по ссылкам
	// И вторая операция - формирование ответа по порядку вызовов print, разыменовывая ссылки
	results := map[string]*int64{}
	var resultOrder []string
	for _, operation := range operations {
		switch operation.Type {
		case CalcOperation:
			computeResult, err := c.computeCalcOperation(&results, operation)
			if err != nil {
				// TODO сделать корректную ошибку
				fmt.Print("vse ploho")
			} else {
				results[operation.Var] = computeResult
			}
		case PrintOperation:
			resultOrder = append(resultOrder, operation.Var)
		default: // TODO сделать корректную ошибку
		}
	}
	result := []PrintOutput{}
	for _, variableToPrint := range resultOrder {
		resultVar := PrintOutput{
			Var:   variableToPrint,
			Value: *results[variableToPrint],
		}
		result = append(result, resultVar)
	}
	return result, nil
}

func (c *Calculator) computeCalcOperation(variableValues *map[string]*int64, operation Operation) (*int64, error) {
	//TODO обработка ошибок, если в мапе нет такой переменной
	var leftValue, rightValue int64
	if operation.Left.IntValue != nil {
		leftValue = *operation.Left.IntValue
	} else if operation.Left.StringValue != nil {
		leftValue = *(*variableValues)[*operation.Left.StringValue]
	} else {
		return nil, errors.New("invalid operand")
	}

	if operation.Right.IntValue != nil {
		rightValue = *operation.Right.IntValue
	} else if operation.Right.StringValue != nil {
		rightValue = *(*variableValues)[*operation.Right.StringValue]
	} else {
		return nil, errors.New("invalid operand")
	}

	result := new(int64)
	switch operation.Op {
	case "+":
		*result = leftValue + rightValue
	case "-":
		*result = leftValue - rightValue
	case "*":
		*result = leftValue * rightValue
	case "/":
		*result = leftValue / rightValue
	}
	c.logger.Debug(fmt.Sprintf("Result: %d", *result))
	return result, nil
}
