package grpc

import (
	"context"
	"errors"
	"log/slog"
	"upgraded-calculator/gen"
	"upgraded-calculator/internal/common"
)

type CalculatorGRPC struct {
	logger *slog.Logger
}

func (ca *CalculatorGRPC) Execute(
	ctx context.Context,
	request *gen.Request,
) (response *gen.Response, err error) {
	ca.logger.Info("Processing GRPC request with request_id", "request_id", ctx.Value("request_id"))
	c := common.NewUpgradedCalculator(ca.logger, ctx.Value("request_id").(string))
	var operations []common.Operation
	for _, op := range request.GetOperation() {
		if validatedOp, err := ca.validateAndParseOperation(op); err == nil {
			operations = append(operations, *validatedOp)
		} else {
			ca.logger.Error(err.Error())
		}
	}
	result, err := c.Execute(operations)
	if err != nil {
		ca.logger.Error(err.Error())
		return nil, err
	}
	formedResponse, _ := ca.formResponse(result)
	resp := &gen.Response{Items: formedResponse}
	ca.logger.Info("Response formed", "request_id", ctx.Value("request_id").(string))
	return resp, nil
}

func (ca *CalculatorGRPC) validateAndParseOperation(op *gen.Operation) (*common.Operation, error) {
	result := common.Operation{}
	result.Var = op.Var
	switch common.OperationType(op.Type) {
	case common.CalcOperation:
		result.Type = common.OperationType(op.Type)
		if op.Op == nil {
			return nil, errors.New("operation cannot be nil")
		}
		switch common.CalcAvailableOperation(*op.Op) {
		case common.Add, common.Sub, common.Mul, common.Div:
			result.Op = common.CalcAvailableOperation(*op.Op)
		default:
			return nil, errors.New("invalid operation type from request")
		}

		switch v := op.Left.GetValue().(type) {
		case *gen.Operand_Number:
			result.Left = &common.Operand{IntValue: &v.Number, StringValue: nil}
		case *gen.Operand_Variable:
			result.Left = &common.Operand{IntValue: nil, StringValue: &v.Variable}
		}

		switch v := op.Right.GetValue().(type) {
		case *gen.Operand_Number:
			result.Right = &common.Operand{IntValue: &v.Number, StringValue: nil}
		case *gen.Operand_Variable:
			result.Right = &common.Operand{IntValue: nil, StringValue: &v.Variable}
		}
	case common.PrintOperation:
		result.Type = common.OperationType(op.Type)
	default:
		return nil, errors.New("invalid operation type from request")
	}
	return &result, nil
}

func (ca *CalculatorGRPC) formResponse(outputList []common.PrintOutput) ([]*gen.Variable, error) {
	result := make([]*gen.Variable, 0, len(outputList))
	for _, op := range outputList {
		result = append(result, &gen.Variable{Var: op.Var, Value: op.Value})
	}
	return result, nil
}
