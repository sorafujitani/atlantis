package adapter

import (
	"context"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/sorafujitani/atlantis/internal/config"
)

type DoctorResult struct {
	Name         string       `json:"name"`
	Command      string       `json:"command"`
	Path         string       `json:"path,omitempty"`
	Version      string       `json:"version,omitempty"`
	Available    bool         `json:"available"`
	Capabilities Capabilities `json:"capabilities"`
	Error        string       `json:"error,omitempty"`
}

func Doctor(ctx context.Context, cfg config.Config) []DoctorResult {
	results := make([]DoctorResult, 0, len(cfg.Adapters))
	for name, adapterConfig := range cfg.Adapters {
		result := DoctorResult{Name: name, Command: adapterConfig.Command, Capabilities: capabilitiesFor(name, adapterConfig.Args)}
		path, err := exec.LookPath(adapterConfig.Command)
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		result.Path = path
		checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		command := exec.CommandContext(checkCtx, path, "--version") //nolint:gosec // path came from exec.LookPath for the configured adapter
		command.Env = safeEnvironment(adapterConfig.InheritEnv)
		output, err := command.CombinedOutput()
		cancel()
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		result.Available = true
		result.Version = strings.TrimSpace(string(output))
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	return results
}
