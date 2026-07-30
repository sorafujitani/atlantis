// Package integrations installs host adapters embedded in the Atlantis binary.
package integrations

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed brain-context.js
var brainContextJS []byte

// Target identifies one supported agent host.
type Target string

const (
	TargetOMP      Target = "omp"
	TargetPi       Target = "pi"
	TargetOpenCode Target = "opencode"
)

// Targets returns every supported host in stable install order.
func Targets() []Target {
	return []Target{TargetOMP, TargetPi, TargetOpenCode}
}

// Install writes the embedded adapter for target and returns its path.
func Install(target Target, dir string) (string, error) {
	if dir == "" {
		var err error
		dir, err = defaultDir(target)
		if err != nil {
			return "", err
		}
	} else if !validTarget(target) {
		return "", fmt.Errorf("unsupported integration target %q", target)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create integration directory: %w", err)
	}
	legacy := filepath.Join(dir, "atlantis-brain.ts")
	if err := os.Remove(legacy); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("remove legacy adapter: %w", err)
	}
	path := filepath.Join(dir, "atlantis-brain.js")
	if err := os.WriteFile(path, brainContextJS, 0o600); err != nil {
		return "", fmt.Errorf("write integration adapter: %w", err)
	}
	return path, nil
}

func defaultDir(target Target) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	switch target {
	case TargetOMP:
		return filepath.Join(home, ".omp", "agent", "extensions"), nil
	case TargetPi:
		return filepath.Join(home, ".pi", "agent", "extensions"), nil
	case TargetOpenCode:
		return filepath.Join(home, ".config", "opencode", "plugins"), nil
	default:
		return "", fmt.Errorf("unsupported integration target %q", target)
	}
}

func validTarget(target Target) bool {
	return target == TargetOMP || target == TargetPi || target == TargetOpenCode
}
