package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"upgraded-calculator/internal/common"
)

type CalculatorHTTP struct {
	logger *slog.Logger
}

func (ca *CalculatorHTTP) Execute(
	ctx context.Context,
	data []byte,
) ([]byte, error) {
	ca.logger.Info("Processing HTTP request with request_id", "request_id", ctx.Value("request_id"))
	c := common.NewUpgradedCalculator(ca.logger, ctx.Value("request_id").(string))
	var req []common.Operation
	err := json.Unmarshal(data, &req)
	if err != nil {
		ca.logger.Error(err.Error())
		return nil, err
	}

	result, err := c.Execute(req)
	if err != nil {
		ca.logger.Error(err.Error())
		return nil, err
	}

	ca.logger.Info("Request finished")
	c = nil
	formedResponse, err := json.Marshal(result)
	if err != nil {
		ca.logger.Error(err.Error())
		return nil, err
	}
	return formedResponse, nil
}
