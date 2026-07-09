package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/sorafujitani/model-orchestrator/internal/orchestration"
)

type Adapter struct {
	Command    string   `toml:"command"`
	Args       []string `toml:"args"`
	InheritEnv []string `toml:"inherit_env"`
}

type Model struct {
	Adapter  string   `toml:"adapter"`
	Model    string   `toml:"model"`
	Fallback []string `toml:"fallback"`
}

type Profile struct {
	Mode            orchestration.ExecutionMode `toml:"mode"`
	Orchestrator    string                      `toml:"orchestrator"`
	Executor        string                      `toml:"executor"`
	Advisor         string                      `toml:"advisor"`
	Worker          string                      `toml:"worker"`
	Reviewer        string                      `toml:"reviewer"`
	MaxWorkers      int                         `toml:"max_workers"`
	MaxCalls        int                         `toml:"max_calls"`
	MaxAdvisorCalls int                         `toml:"max_advisor_calls"`
	MaxRetries      int                         `toml:"max_retries"`
	MaxDuration     time.Duration               `toml:"-"`
	MaxDurationText string                      `toml:"max_duration"`
}

type Config struct {
	DefaultProfile string             `toml:"default_profile"`
	StateDir       string             `toml:"state_dir"`
	Adapters       map[string]Adapter `toml:"adapters"`
	Models         map[string]Model   `toml:"models"`
	Profiles       map[string]Profile `toml:"profiles"`
}

func Default() Config {
	return Config{
		DefaultProfile: "default",
		Adapters: map[string]Adapter{
			"codex":    {Command: "codex"},
			"claude":   {Command: "claude"},
			"opencode": {Command: "opencode"},
			"cursor":   {Command: "cursor-agent"},
			"copilot":  {Command: "copilot"},
		},
		Models: map[string]Model{
			"premium":  {Adapter: "codex"},
			"standard": {Adapter: "claude", Fallback: []string{"premium"}},
		},
		Profiles: map[string]Profile{
			"default": {
				Mode:            orchestration.ModeHybrid,
				Orchestrator:    "premium",
				Executor:        "standard",
				Advisor:         "premium",
				Worker:          "standard",
				Reviewer:        "premium",
				MaxWorkers:      4,
				MaxCalls:        20,
				MaxAdvisorCalls: 1,
				MaxRetries:      1,
				MaxDuration:     30 * time.Minute,
				MaxDurationText: "30m",
			},
		},
	}
}

func Path(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	if value := os.Getenv("MODEL_ORCHESTRATOR_CONFIG"); value != "" {
		return filepath.Abs(value)
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "model-orchestrator", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", "model-orchestrator", "config.toml"), nil
}

func StateDir(cfg Config) (string, error) {
	if value := os.Getenv("MODEL_ORCHESTRATOR_STATE_DIR"); value != "" {
		return filepath.Abs(value)
	}
	if cfg.StateDir != "" {
		return filepath.Abs(cfg.StateDir)
	}
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "model-orchestrator"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".local", "state", "model-orchestrator"), nil
}

func Load(explicit string) (Config, string, error) {
	cfg := Default()
	path, err := Path(explicit)
	if err != nil {
		return Config{}, "", err
	}
	data, err := os.ReadFile(path) //nolint:gosec // reading the explicitly selected config file is intended
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := cfg.normalize(); err != nil {
				return Config{}, path, err
			}
			return cfg, path, nil
		}
		return Config{}, path, fmt.Errorf("read config: %w", err)
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.normalize(); err != nil {
		return Config{}, path, err
	}
	return cfg, path, nil
}

func (c *Config) normalize() error {
	if value := os.Getenv("MODEL_ORCHESTRATOR_PROFILE"); value != "" {
		c.DefaultProfile = value
	}
	for name, profile := range c.Profiles {
		if profile.MaxDurationText != "" {
			duration, err := time.ParseDuration(profile.MaxDurationText)
			if err != nil {
				return fmt.Errorf("profile %q max_duration: %w", name, err)
			}
			profile.MaxDuration = duration
		}
		if profile.MaxDuration <= 0 {
			profile.MaxDuration = 30 * time.Minute
		}
		c.Profiles[name] = profile
	}
	return c.Validate()
}

func (c Config) Validate() error {
	if c.DefaultProfile == "" {
		return errors.New("default_profile is required")
	}
	if _, exists := c.Profiles[c.DefaultProfile]; !exists {
		return fmt.Errorf("default profile %q does not exist", c.DefaultProfile)
	}
	for name, adapter := range c.Adapters {
		if strings.TrimSpace(adapter.Command) == "" {
			return fmt.Errorf("adapter %q command is required", name)
		}
	}
	for name, model := range c.Models {
		if _, exists := c.Adapters[model.Adapter]; !exists {
			return fmt.Errorf("model %q references unknown adapter %q", name, model.Adapter)
		}
		for _, fallback := range model.Fallback {
			if _, exists := c.Models[fallback]; !exists {
				return fmt.Errorf("model %q references unknown fallback %q", name, fallback)
			}
		}
	}
	for name, profile := range c.Profiles {
		if profile.MaxWorkers <= 0 || profile.MaxCalls <= 0 || profile.MaxAdvisorCalls < 0 || profile.MaxRetries < 0 {
			return fmt.Errorf("profile %q has invalid limits", name)
		}
		for role, alias := range map[string]string{
			"orchestrator": profile.Orchestrator,
			"executor":     profile.Executor,
			"advisor":      profile.Advisor,
			"worker":       profile.Worker,
			"reviewer":     profile.Reviewer,
		} {
			if alias == "" {
				continue
			}
			if _, exists := c.Models[alias]; !exists {
				return fmt.Errorf("profile %q role %s references unknown model %q", name, role, alias)
			}
		}
	}
	return nil
}

func (c Config) Profile(name string) (Profile, error) {
	if name == "" {
		name = c.DefaultProfile
	}
	profile, exists := c.Profiles[name]
	if !exists {
		return Profile{}, fmt.Errorf("profile %q does not exist", name)
	}
	return profile, nil
}
