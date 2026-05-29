package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/santgabo/usbfixertool-cli/internal/cli"
	"github.com/santgabo/usbfixertool-cli/internal/usbfix"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := cli.NewRootCommand(cli.Dependencies{
		Runner: usbfix.ExecRunner{},
		Stdin:  os.Stdin,
	})

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(usbfix.ExitCode(err))
	}
}
