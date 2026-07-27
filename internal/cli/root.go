// Package cli exposes the atlantis command tree.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sorafujitani/atlantis/internal/config"
	"github.com/sorafujitani/atlantis/internal/engine"
	"github.com/sorafujitani/atlantis/internal/evaluation"
	"github.com/sorafujitani/atlantis/internal/orchestration"
	"github.com/sorafujitani/atlantis/internal/state"
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

var errModelRoutingDisabled = errors.New("model routing is disabled; use the current harness directly")

type app struct {
	configPath string
	output     string
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer
}

// NewCommand constructs a command tree with injectable streams for tests.
func NewCommand(stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	a := &app{stdin: stdin, stdout: stdout, stderr: stderr, output: "plain"}
	root := &cobra.Command{
		Use:           "atlantis",
		Short:         "Agent operating mode and persistent context maintenance",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.PersistentFlags().StringVar(&a.configPath, "config", "", "config file (default: XDG config path)")
	root.PersistentFlags().StringVarP(&a.output, "output", "o", "plain", "output format: plain or json")
	root.AddCommand(
		disabledModelCommand("run [arguments...]"), a.statusCommand(false), a.statusCommand(true),
		disabledModelCommand("resume [arguments...]"), a.cancelCommand(), a.doctorCommand(),
		disabledModelCommand("eval [arguments...]"), a.brainCommand(), versionCommand(),
	)
	return root
}

func disabledModelCommand(use string) *cobra.Command {
	return &cobra.Command{
		Use:                use,
		Short:              "Disabled model-routing command",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return errModelRoutingDisabled
		},
	}
}

func (a *app) routingDependencies() (config.Config, *engine.Engine, error) {
	cfg, _, err := config.Load(a.configPath)
	if err != nil {
		return config.Config{}, nil, err
	}
	root, err := config.StateDir(cfg)
	if err != nil {
		return config.Config{}, nil, err
	}
	store, err := state.New(root)
	if err != nil {
		return config.Config{}, nil, err
	}
	return cfg, engine.New(cfg, store), nil
}

func (a *app) stateEngine() (*engine.Engine, error) {
	root, err := config.LoadStateDir(a.configPath)
	if err != nil {
		return nil, err
	}
	store, err := state.New(root)
	if err != nil {
		return nil, err
	}
	return engine.New(config.Config{}, store), nil
}

func (a *app) runCommand() *cobra.Command {
	var cwd, profile, mode, permission string
	command := &cobra.Command{
		Use:   "run <objective>",
		Short: "Run an orchestration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			_, orchestrator, err := a.routingDependencies()
			if err != nil {
				return err
			}
			if cwd == "" {
				cwd, err = os.Getwd()
				if err != nil {
					return err
				}
			}
			cwd, err = filepath.Abs(cwd)
			if err != nil {
				return err
			}
			request := orchestration.Request{
				Objective: strings.Join(args, " "), CWD: cwd, Profile: profile,
				Mode: orchestration.ExecutionMode(mode), Permission: orchestration.Permission(permission),
			}
			runID, result, err := orchestrator.Start(command.Context(), request)
			if err != nil {
				if runID != "" {
					return fmt.Errorf("run %s: %w", runID, err)
				}
				return err
			}
			return a.print(map[string]any{"run_id": runID, "result": result}, outputOf(result))
		},
	}
	command.Flags().StringVar(&cwd, "cwd", "", "working directory")
	command.Flags().StringVar(&profile, "profile", "", "orchestration profile")
	command.Flags().StringVar(&mode, "mode", "", "single, advisor, orchestrator, or hybrid")
	command.Flags().StringVar(&permission, "permission", "read", "read or write (single/advisor root assignment)")
	return command
}

func (a *app) statusCommand(inspect bool) *cobra.Command {
	name, short := "status", "Show a run snapshot"
	if inspect {
		name, short = "inspect", "Show a run snapshot and events"
	}
	return &cobra.Command{
		Use: name + " <run-id>", Short: short, Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			orchestrator, err := a.stateEngine()
			if err != nil {
				return err
			}
			snapshot, err := orchestrator.Snapshot(args[0])
			if err != nil {
				return err
			}
			if !inspect {
				return a.print(snapshot, fmt.Sprintf("%s\t%s", snapshot.RunID, snapshot.Status))
			}
			events, err := orchestrator.Events(args[0])
			if err != nil {
				return err
			}
			return a.print(map[string]any{"snapshot": snapshot, "events": events}, fmt.Sprintf("%s\t%s\t%d events", snapshot.RunID, snapshot.Status, len(events)))
		},
	}
}

func (a *app) resumeCommand() *cobra.Command {
	return &cobra.Command{
		Use: "resume <run-id>", Short: "Resume an interrupted run", Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			_, orchestrator, err := a.routingDependencies()
			if err != nil {
				return err
			}
			result, err := orchestrator.Resume(command.Context(), args[0])
			if err != nil {
				return err
			}
			return a.print(map[string]any{"run_id": args[0], "result": result}, outputOf(result))
		},
	}
}

func (a *app) cancelCommand() *cobra.Command {
	return &cobra.Command{
		Use: "cancel <run-id>", Short: "Cancel an active or interrupted run", Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			orchestrator, err := a.stateEngine()
			if err != nil {
				return err
			}
			if err := orchestrator.Cancel(args[0]); err != nil {
				return err
			}
			return a.print(map[string]any{"run_id": args[0], "cancelled": true}, "cancelled "+args[0])
		},
	}
}

func (a *app) doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use: "doctor", Short: "Inspect dormant model CLI availability", Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, orchestrator, err := a.routingDependencies()
			if err != nil {
				return err
			}
			results := orchestrator.Doctor(command.Context())
			if a.output == "json" {
				return a.print(results, "")
			}
			for _, result := range results {
				status := "missing"
				if result.Available {
					status = "ok"
				}
				if _, err := fmt.Fprintf(a.stdout, "%s\t%s\t%s\t%s\n", result.Name, status, result.Path, result.Version); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func (a *app) evalCommand() *cobra.Command {
	var profiles []string
	var cwd string
	command := &cobra.Command{
		Use: "eval <suite.json>", Short: "Evaluate orchestration profiles against a versioned suite", Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			cfg, orchestrator, err := a.routingDependencies()
			if err != nil {
				return err
			}
			suite, err := evaluation.Load(args[0])
			if err != nil {
				return err
			}
			if cwd == "" {
				cwd, err = os.Getwd()
				if err != nil {
					return err
				}
			}
			if len(profiles) == 0 {
				profiles = []string{cfg.DefaultProfile}
			}
			report := evaluation.Report{SuiteVersion: suite.Version, GeneratedAt: time.Now().UTC()}
			for _, profileName := range profiles {
				profileStart := time.Now()
				profileResult := evaluation.ProfileResult{Profile: profileName, Total: len(suite.Cases)}
				for _, testCase := range suite.Cases {
					caseStart := time.Now()
					runID, result, runErr := orchestrator.Start(command.Context(), orchestration.Request{
						Objective: testCase.Objective, CWD: cwd, Profile: profileName, Mode: testCase.Mode, Permission: testCase.Permission,
					})
					caseResult := evaluation.CaseResult{ID: testCase.ID, RunID: runID, DurationMS: time.Since(caseStart).Milliseconds()}
					if runErr != nil {
						caseResult.Error = runErr.Error()
					} else {
						caseResult.Output = outputOf(result)
						caseResult.Passed = evaluation.Score(testCase, caseResult.Output)
						if caseResult.Passed {
							profileResult.Passed++
						}
						if snapshot, snapshotErr := orchestrator.Snapshot(runID); snapshotErr == nil {
							for _, completed := range snapshot.CompletedTasks {
								evaluation.AddUsage(&caseResult.Usage, completed.Result.Usage)
							}
						}
					}
					evaluation.AddUsage(&profileResult.Usage, caseResult.Usage)
					profileResult.Cases = append(profileResult.Cases, caseResult)
				}
				profileResult.DurationMS = time.Since(profileStart).Milliseconds()
				report.Profiles = append(report.Profiles, profileResult)
			}
			if a.output == "plain" {
				for _, profile := range report.Profiles {
					if _, err := fmt.Fprintf(a.stdout, "%s\t%d/%d\t%dms\n", profile.Profile, profile.Passed, profile.Total, profile.DurationMS); err != nil {
						return err
					}
				}
				return nil
			}
			return a.print(report, "")
		},
	}
	command.Flags().StringSliceVar(&profiles, "profiles", nil, "profiles to compare")
	command.Flags().StringVar(&cwd, "cwd", "", "working directory")
	return command
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

func outputOf(result orchestration.Result) string {
	if result.Output != "" {
		return result.Output
	}
	return result.Summary
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
