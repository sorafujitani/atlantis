// Package main provides the model-orchestrator executable.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sorafujitani/model-orchestrator/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err := cli.Execute(ctx, os.Stdin, os.Stdout, os.Stderr, os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(cli.ExitCode(err))
}
