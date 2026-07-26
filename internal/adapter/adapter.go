package adapter

import (
	"context"
	"fmt"

	"github.com/sorafujitani/atlantis/internal/config"
	"github.com/sorafujitani/atlantis/internal/orchestration"
)

type Capabilities struct {
	JSONStream    bool `json:"json_stream"`
	NativeSchema  bool `json:"native_schema"`
	Resume        bool `json:"resume"`
	Usage         bool `json:"usage"`
	Read          bool `json:"read"`
	Write         bool `json:"write"`
	FinalTextOnly bool `json:"final_text_only"`
}

type Invocation struct {
	Assignment orchestration.Assignment
	Prompt     string
	Model      string
	Session    *orchestration.NativeSessionRef
}

type Runner interface {
	Name() string
	Capabilities() Capabilities
	Run(context.Context, Invocation) (orchestration.ExecutionResult, error)
}

type Factory struct{ cfg config.Config }

func NewFactory(cfg config.Config) *Factory { return &Factory{cfg: cfg} }

func (f *Factory) Runner(modelAlias string) (Runner, config.Model, error) {
	model, exists := f.cfg.Models[modelAlias]
	if !exists {
		return nil, config.Model{}, fmt.Errorf("model alias %q does not exist", modelAlias)
	}
	adapterConfig, exists := f.cfg.Adapters[model.Adapter]
	if !exists {
		return nil, config.Model{}, fmt.Errorf("adapter %q does not exist", model.Adapter)
	}
	runner, err := NewProcessRunner(model.Adapter, adapterConfig)
	if err != nil {
		return nil, config.Model{}, err
	}
	return runner, model, nil
}

func (f *Factory) Chain(modelAlias string) ([]string, error) {
	seen := map[string]bool{}
	var chain []string
	var walk func(string) error
	walk = func(alias string) error {
		if seen[alias] {
			return nil
		}
		seen[alias] = true
		model, exists := f.cfg.Models[alias]
		if !exists {
			return fmt.Errorf("model alias %q does not exist", alias)
		}
		chain = append(chain, alias)
		for _, fallback := range model.Fallback {
			if err := walk(fallback); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(modelAlias); err != nil {
		return nil, err
	}
	return chain, nil
}
