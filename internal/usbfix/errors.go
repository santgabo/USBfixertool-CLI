package usbfix

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

const (
	ExitOK              = 0
	ExitRuntime         = 1
	ExitUsage           = 2
	ExitCancelled       = 3
	ExitPermission      = 126
	ExitCommandNotFound = 127
	ExitInterrupted     = 130
)

type Error struct {
	Code    int
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" && e.Err != nil {
		return e.Err.Error()
	}
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewError(code int, message string, err error) error {
	return &Error{Code: code, Message: message, Err: err}
}

func Runtimef(format string, args ...any) error {
	return NewError(ExitRuntime, fmt.Sprintf(format, args...), nil)
}

func Usagef(format string, args ...any) error {
	return NewError(ExitUsage, fmt.Sprintf(format, args...), nil)
}

func Cancelledf(format string, args ...any) error {
	return NewError(ExitCancelled, fmt.Sprintf(format, args...), nil)
}

func CommandNotFoundf(format string, args ...any) error {
	return NewError(ExitCommandNotFound, fmt.Sprintf(format, args...), nil)
}

func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	if errors.Is(err, context.Canceled) {
		return ExitInterrupted
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return ExitCommandNotFound
	}
	return ExitRuntime
}
