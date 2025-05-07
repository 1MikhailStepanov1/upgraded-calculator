package common

import (
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to create pointers for int64
func int64Ptr(i int64) *int64 {
	return &i
}

func TestUpgradedCalculator_ComputeOperations(t *testing.T) {
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	calculator := NewUpgradedCalculator(logger, "compute")

	operations := []Operation{
		{
			Type:  CalcOperation,
			Var:   "x",
			Left:  &Operand{IntValue: int64Ptr(1)},
			Right: &Operand{IntValue: int64Ptr(2)},
			Op:    "+",
		},
		{
			Type: PrintOperation,
			Var:  "x",
		},
		{
			Type:  CalcOperation,
			Var:   "y",
			Left:  &Operand{IntValue: int64Ptr(5)},
			Right: &Operand{IntValue: int64Ptr(4)},
			Op:    "-",
		},
	}

	expectedOutputs := []PrintOutput{
		{Var: "x", Value: 3},
	}

	actualOutputs, err := calculator.Execute(operations)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutputs, actualOutputs)
}

func TestUpgradedCalculator_DivisionByZero(t *testing.T) {
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	calculator := NewUpgradedCalculator(logger, "division_by_zero")

	operations := []Operation{
		{
			Type:  CalcOperation,
			Var:   "x",
			Left:  &Operand{IntValue: int64Ptr(10)},
			Right: &Operand{IntValue: int64Ptr(0)},
			Op:    "/",
		},
	}

	_, err := calculator.Execute(operations)
	assert.Error(t, err)
	assert.Equal(t, "division by zero", err.Error())
}

func TestUpgradedCalculator_SubscribeVariable(t *testing.T) {
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	calculator := NewUpgradedCalculator(logger, "test_request")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		value, err := calculator.subscribeVariable("y")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), value)
	}()

	err := calculator.publishVariable("y", 100)
	assert.NoError(t, err)

	wg.Wait()
}

func TestUpgradedCalculator_InvalidOperation(t *testing.T) {
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	calculator := NewUpgradedCalculator(logger, "invalid_operation")

	operations := []Operation{
		{
			Type:  CalcOperation,
			Var:   "x",
			Left:  &Operand{IntValue: int64Ptr(5)},
			Right: &Operand{IntValue: int64Ptr(3)},
			Op:    "pipipupu",
		},
	}

	_, err := calculator.Execute(operations)
	assert.Error(t, err)
	assert.Equal(t, "invalid operation", err.Error())
}
