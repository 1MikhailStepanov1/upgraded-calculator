package common

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type OperationType string

const (
	CalcOperation  OperationType = "calc"
	PrintOperation OperationType = "print"
)

func (opType *OperationType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch OperationType(s) {
	case CalcOperation, PrintOperation:
		*opType = OperationType(s)
		return nil
	default:
		return errors.New("invalid operation type")
	}
}

type CalcAvailableOperation string

const (
	Add CalcAvailableOperation = "+"
	Sub CalcAvailableOperation = "-"
	Mul CalcAvailableOperation = "*"
	Div CalcAvailableOperation = "/"
)

func (opType *CalcAvailableOperation) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch CalcAvailableOperation(s) {
	case Add, Sub, Mul, Div:
		*opType = CalcAvailableOperation(s)
		return nil
	default:
		return errors.New("calculator unavailable operation")
	}
}

type Operand struct {
	IntValue    *int64
	StringValue *string
}

func (op *Operand) UnmarshalJSON(b []byte) error {
	var s = string(b)
	s = strings.ReplaceAll(s, "\"", "")

	if num, err := strconv.ParseInt(s, 10, 64); err == nil {
		*op = Operand{IntValue: &num, StringValue: nil}
		return nil
	} else if m, err := regexp.MatchString("^[A-z]+$", s); err == nil && m {
		*op = Operand{IntValue: nil, StringValue: &s}
		return nil
	}
	return errors.New("invalid type of operand")
}

type Operation struct {
	Type  OperationType          `json:"type"`
	Op    CalcAvailableOperation `json:"op,omitempty"`
	Var   string                 `json:"var"`
	Left  *Operand               `json:"left,omitempty"`
	Right *Operand               `json:"right,omitempty"`
}

type PrintOutput struct {
	Var   string `json:"var"`
	Value int64  `json:"value"`
}
