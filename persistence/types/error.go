package types

import (
	"errors"
	"fmt"
)

type Error interface {
	Code() Code
	error
}

type StdErr struct {
	CodeError Code
	error
}

func (se StdErr) Error() string {
	return fmt.Sprintf("CODE: %v, ERROR: %s", se.Code(), se.error.Error())
}

func (se StdErr) Code() Code {
	return se.CodeError
}

func NewError(code Code, msg string) Error {
	return StdErr{
		CodeError: code,
		error:     errors.New(msg),
	}
}

type Code float64

const (
	CodeUnknownMessageError Code = 6

	UnknownMessageError = "the message type is unrecognized"
)

func ErrUnknownMessage(msg interface{}) Error {
	return NewError(CodeUnknownMessageError, fmt.Sprintf("%s: %T", UnknownMessageError, msg))
}
