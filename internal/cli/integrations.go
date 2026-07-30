package cli

import (
	"errors"
	"strings"

	"github.com/sorafujitani/atlantis/integrations"
	"github.com/spf13/cobra"
)

func (a *app) integrationsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "integrations",
		Short: "Install embedded agent host adapters",
	}
	var dir string
	install := &cobra.Command{
		Use:   "install [omp|pi|opencode]",
		Short: "Install adapters from this binary",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				if dir != "" {
					return errors.New("--dir requires one integration target")
				}
				paths := make([]string, 0, len(integrations.Targets()))
				for _, target := range integrations.Targets() {
					path, err := integrations.Install(target, "")
					if err != nil {
						return err
					}
					paths = append(paths, path)
				}
				return a.print(map[string]any{"installed": paths}, strings.Join(paths, "\n"))
			}
			path, err := integrations.Install(integrations.Target(args[0]), dir)
			if err != nil {
				return err
			}
			return a.print(map[string]any{"installed": []string{path}}, path)
		},
	}
	install.Flags().StringVar(&dir, "dir", "", "override the target extension directory")
	command.AddCommand(install)
	return command
}
