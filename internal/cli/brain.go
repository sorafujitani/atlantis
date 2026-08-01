package cli

import (
	"errors"
	"fmt"

	atlantis "github.com/sorafujitani/atlantis"
	"github.com/sorafujitani/atlantis/internal/brain"
	"github.com/spf13/cobra"
)

var errInvalidBrain = errors.New("brain validation failed")

func (a *app) brainCommand() *cobra.Command {
	var root string
	command := &cobra.Command{
		Use:   "brain",
		Short: "Manage portable agent memory",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	command.PersistentFlags().StringVar(&root, "dir", "", "brain directory (default: ATLANTIS_BRAIN_DIR or ~/brain)")
	vault := func() (*brain.Vault, error) {
		return brain.New(root)
	}

	command.AddCommand(
		&cobra.Command{
			Use: "init", Short: "Create an empty brain vault", Args: cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				selected, err := vault()
				if err != nil {
					return err
				}
				if err := selected.Init(); err != nil {
					return err
				}
				return a.print(map[string]any{"brain_dir": selected.Root, "initialized": true}, selected.Root)
			},
		},
		&cobra.Command{
			Use: "seed", Short: "Install repo-managed brain documents", Args: cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				selected, err := vault()
				if err != nil {
					return err
				}
				if err := selected.Seed(atlantis.BrainSeed()); err != nil {
					return err
				}
				if err := selected.ForceIndex(); err != nil {
					return err
				}
				return a.print(map[string]any{"brain_dir": selected.Root, "seeded": true}, "seeded "+selected.Root)
			},
		},
	)

	var indexForce bool
	indexCmd := &cobra.Command{
		Use: "index", Short: "Regenerate brain indexes", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			selected, err := vault()
			if err != nil {
				return err
			}
			if indexForce {
				err = selected.ForceIndex()
			} else {
				err = selected.Index()
			}
			if err != nil {
				return err
			}
			return a.print(map[string]any{"brain_dir": selected.Root, "indexed": true}, "indexed "+selected.Root)
		},
	}
	indexCmd.Flags().BoolVar(&indexForce, "force", false, "rebuild indexes even when the vault fingerprint is unchanged")
	command.AddCommand(indexCmd)

	var contextForce bool
	var printFingerprint bool
	contextCmd := &cobra.Command{
		Use: "context", Short: "Print refreshed agent context", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			selected, err := vault()
			if err != nil {
				return err
			}
			if printFingerprint {
				fingerprint, err := selected.Fingerprint()
				if err != nil {
					return err
				}
				return a.print(map[string]any{"brain_dir": selected.Root, "fingerprint": fingerprint}, fingerprint)
			}
			result, err := selected.ContextResult(contextForce)
			if err != nil {
				return err
			}
			if a.output != "plain" {
				return a.print(map[string]any{
					"brain_dir":   selected.Root,
					"context":     result.Context,
					"fingerprint": result.Fingerprint,
				}, result.Context)
			}
			_, err = fmt.Fprint(a.stdout, result.Context)
			return err
		},
	}
	contextCmd.Flags().BoolVar(&contextForce, "force", false, "rebuild indexes even when the vault fingerprint is unchanged")
	contextCmd.Flags().BoolVar(&printFingerprint, "print-fingerprint", false, "print only the vault source fingerprint")
	command.AddCommand(contextCmd)

	command.AddCommand(&cobra.Command{
		Use: "check", Short: "Validate links, reachability, and note size", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			selected, err := vault()
			if err != nil {
				return err
			}
			report, err := selected.Check()
			if err != nil {
				return err
			}
			plain := fmt.Sprintf("%d/%d reachable, %d broken, %d oversized", report.Reachable, report.Files, len(report.BrokenLinks), len(report.Oversized))
			if err := a.print(report, plain); err != nil {
				return err
			}
			if len(report.BrokenLinks) > 0 || len(report.Unreachable) > 0 || len(report.Oversized) > 0 {
				return errInvalidBrain
			}
			return nil
		},
	})

	plan := &cobra.Command{Use: "plan", Short: "Manage transient implementation plans"}
	plan.AddCommand(&cobra.Command{
		Use: "finish <slug>", Short: "Delete a completed plan and rebuild indexes", Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			selected, err := vault()
			if err != nil {
				return err
			}
			if err := selected.FinishPlan(args[0]); err != nil {
				return err
			}
			return a.print(map[string]any{"plan": args[0], "removed": true}, "removed "+args[0])
		},
	})
	command.AddCommand(plan)
	return command
}
