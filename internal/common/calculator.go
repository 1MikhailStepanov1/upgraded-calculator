package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"upgraded-calculator/internal/config"
)

type UpgradedCalculator struct {
	logger    *slog.Logger
	requestId string
	variables map[string]int64
	subs      map[string][]chan int64
	mutex     sync.Mutex
}

func NewUpgradedCalculator(
	logger *slog.Logger,
	requestId string,
) *UpgradedCalculator {
	return &UpgradedCalculator{
		logger:    logger,
		requestId: requestId,
		variables: make(map[string]int64),
		subs:      make(map[string][]chan int64),
	}
}

func (c *UpgradedCalculator) Execute(cont context.Context, operations []Operation) ([]PrintOutput, error) {
	var (
		result      []PrintOutput
		resultMu    sync.Mutex
		wg          sync.WaitGroup
		errorsCh    = make(chan error, 1)
		ctx, cancel = context.WithCancel(cont)
		cfg         = config.New()
	)
	defer cancel()
	defer close(errorsCh)

	workerCount := cfg.App.CalculatorWorkersCount
	if workerCount <= 0 {
		workerCount = 1
	}
	tasksCh := make(chan Operation, len(operations))

	c.logger.Debug("Operations to execute", "request_id", c.requestId, "length", len(operations))

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			defer wg.Done()
			for op := range tasksCh {
				select {
				case <-ctx.Done():
					return
				default:
					var err error

					switch op.Type {
					case CalcOperation:
						err = c.compute(op)
						c.logger.Debug("Compute operation", "request_id", c.requestId, "operation", op)
					case PrintOperation:
						var value int64
						value, err = c.subscribeVariable(op.Var)
						if err == nil {
							resultMu.Lock()
							result = append(result, PrintOutput{
								Var:   op.Var,
								Value: value,
							})
							resultMu.Unlock()
						}
						c.logger.Debug("Print operation", "request_id", c.requestId, "operation", op)
					default:
						err = errors.New("invalid operation")
					}

					if err != nil {
						select {
						case errorsCh <- err:
							cancel()
						default:
						}
					}
				}
			}
		}(i)
	}
	go func() {
		defer close(tasksCh)
		for _, op := range operations {
			select {
			case <-ctx.Done():
				break
			case tasksCh <- op:
			}
		}
	}()

	wg.Wait()

	c.logger.Debug("All operations executed", "request_id", c.requestId)
	select {
	case err := <-errorsCh:
		return nil, err
	default:
		return result, nil
	}
}

func (c *UpgradedCalculator) compute(operation Operation) error {
	leftValue, err := c.getOperandValue(*operation.Left)
	if err != nil {
		return err
	}
	c.logger.Debug("Operand value", "left", leftValue)

	rightValue, err := c.getOperandValue(*operation.Right)
	if err != nil {
		return err
	}

	c.logger.Debug("Operand value", "right", rightValue)

	var res int64
	switch operation.Op {
	case "+":
		res = leftValue + rightValue
	case "-":
		res = leftValue - rightValue
	case "*":
		res = leftValue * rightValue
	case "/":
		if rightValue == 0 {
			return errors.New("division by zero")
		}
		res = leftValue / rightValue
	default:
		return errors.New("invalid operation")
	}

	return c.publishVariable(operation.Var, res)
}

func (c *UpgradedCalculator) getOperandValue(op Operand) (int64, error) {
	if op.IntValue != nil {
		return *op.IntValue, nil
	}
	if op.StringValue != nil {
		return c.subscribeVariable(*op.StringValue)
	}
	return 0, errors.New("invalid operand")
}

func (c *UpgradedCalculator) subscribeVariable(name string) (int64, error) {
	c.mutex.Lock()
	if val, exists := c.variables[name]; exists {
		c.mutex.Unlock()
		return val, nil
	}
	c.mutex.Unlock()

	ch := make(chan int64, 1)

	c.mutex.Lock()
	c.subs[name] = append(c.subs[name], ch)
	c.mutex.Unlock()

	select {
	case val := <-ch:
		return val, nil
	case <-time.After(2 * time.Second):
		c.mutex.Lock()
		for i, subCh := range c.subs[name] {
			if subCh == ch {
				c.subs[name] = append(c.subs[name][:i], c.subs[name][i+1:]...)
				break
			}
		}
		if len(c.subs[name]) == 0 {
			delete(c.subs, name)
		}
		c.mutex.Unlock()
		return 0, fmt.Errorf("variable '%s' is uncomputable", name)
	}
}

func (c *UpgradedCalculator) publishVariable(name string, value int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.variables[name]; exists {
		return fmt.Errorf("variable %s already set", name)
	}

	c.variables[name] = value

	if subscribers, ok := c.subs[name]; ok {
		for _, ch := range subscribers {
			ch <- value
			close(ch)
		}
		delete(c.subs, name)
	}

	return nil
}
