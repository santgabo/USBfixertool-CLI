package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
)

type actionResult struct {
	Command string `json:"command"`
	Output  string `json:"output,omitempty"`
	Skipped bool   `json:"skipped,omitempty"`
}

type operationResult struct {
	Target  string         `json:"target"`
	Mode    string         `json:"mode"`
	DryRun  bool           `json:"dryRun"`
	Actions []actionResult `json:"actions"`
}

func appendDryRun(result *operationResult, command string) {
	result.Actions = append(result.Actions, actionResult{Command: command, Skipped: true})
}

func runWithRetries(ctx context.Context, attempts int, delay time.Duration, fn func(context.Context) ([]byte, error)) ([]byte, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		out, err := fn(ctx)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if i < attempts-1 && delay > 0 {
			select {
			case <-ctx.Done():
				return out, ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return nil, lastErr
}

func renderOperationResult(cmdOutput interface {
	Write([]byte) (int, error)
}, format string, result operationResult) error {
	switch format {
	case "json":
		return writeJSON(cmdOutput, result)
	default:
		for _, action := range result.Actions {
			if action.Skipped {
				fmt.Fprintf(cmdOutput, "DRY-RUN: %s\n", action.Command)
				continue
			}
			if strings.TrimSpace(action.Output) != "" {
				fmt.Fprint(cmdOutput, action.Output)
				if !strings.HasSuffix(action.Output, "\n") {
					fmt.Fprintln(cmdOutput)
				}
			}
		}
		return nil
	}
}

func actionCommand(name string, args ...string) string {
	return usbfix.CommandString(name, args...)
}
