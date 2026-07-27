// Package cli exposes the atlantis command tree.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
	// Version is injected from a release tag at build time.
	Version = "dev"
	// Commit is the source revision embedded in a release build.
	Commit = "none"
	// Date is the release commit timestamp embedded at build time.
	Date = "unknown"
)

type app struct {
	output string
	stdout io.Writer
}

// NewCommand constructs a command tree with injectable streams for tests.
func NewCommand(stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	a := &app{stdout: stdout, output: "plain"}
	root := &cobra.Command{
		Use:           "atlantis",
		Short:         "Maintain portable context for coding agents",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.PersistentFlags().StringVarP(&a.output, "output", "o", "plain", "output format: plain or json")
	root.AddCommand(a.brainCommand(), versionCommand())
	return root
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use: "version", Short: "Print version", Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(command.OutOrStdout(), "atlantis %s (%s, %s)\n", Version, Commit, Date)
			return err
		},
	}
}

func (a *app) print(value any, plain string) error {
	switch a.output {
	case "json":
		encoder := json.NewEncoder(a.stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "plain":
		_, err := fmt.Fprintln(a.stdout, plain)
		return err
	default:
		return fmt.Errorf("unsupported output format %q", a.output)
	}
}

// Execute runs the CLI with explicit streams and arguments.
func Execute(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
	command := NewCommand(stdin, stdout, stderr)
	command.SetArgs(args)
	return command.ExecuteContext(ctx)
}

// ExitCode maps command errors to process exit codes.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) {
		return 130
	}
	return 1
}
