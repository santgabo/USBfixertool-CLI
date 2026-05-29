package usbfix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type SystemRunner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err == nil {
		return out, nil
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return out, ctx.Err()
	}

	message := strings.TrimSpace(stderr.String())
	if message == "" {
		message = err.Error()
	}
	return out, NewError(exitCodeFromExecError(err), fmt.Sprintf("%s failed", CommandString(name, args...)), errors.New(message))
}

func exitCodeFromExecError(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		if code == 126 || code == 127 {
			return code
		}
		return ExitRuntime
	}
	var pathErr *exec.Error
	if errors.As(err, &pathErr) {
		return ExitCommandNotFound
	}
	return ExitRuntime
}

func CommandString(name string, args ...string) string {
	if len(args) == 0 {
		return name
	}
	parts := append([]string{name}, args...)
	return strings.Join(parts, " ")
}
