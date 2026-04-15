// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"errors"
	"fmt"
	"strings"

	"dappco.re/go/core/config"
	"gopkg.in/yaml.v3"
)

// AgentConfig represents a single agent machine in the config file.
type AgentConfig struct {
	Host          string   `yaml:"host" mapstructure:"host"`
	QueueDir      string   `yaml:"queue_dir" mapstructure:"queue_dir"`
	ForgejoUser   string   `yaml:"forgejo_user" mapstructure:"forgejo_user"`
	Model         string   `yaml:"model" mapstructure:"model"`
	Runner        string   `yaml:"runner" mapstructure:"runner"`
	VerifyModel   string   `yaml:"verify_model" mapstructure:"verify_model"`
	SecurityLevel string   `yaml:"security_level" mapstructure:"security_level"`
	Roles         []string `yaml:"roles" mapstructure:"roles"`
	DualRun       bool     `yaml:"dual_run" mapstructure:"dual_run"`
	Active        bool     `yaml:"active" mapstructure:"active"`
}

// ClothoConfig controls the orchestration strategy.
type ClothoConfig struct {
	Strategy            string  `yaml:"strategy" mapstructure:"strategy"`
	ValidationThreshold float64 `yaml:"validation_threshold" mapstructure:"validation_threshold"`
	SigningKeyPath      string  `yaml:"signing_key_path" mapstructure:"signing_key_path"`
}

func defaultClothoConfig() ClothoConfig {
	return ClothoConfig{
		Strategy:            "direct",
		ValidationThreshold: 0.5,
	}
}

// LoadAgents reads agent targets from config and returns a map of AgentConfig.
// Returns an empty map (not an error) if no agents are configured.
func LoadAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	agents := make(map[string]AgentConfig)
	if cfg == nil {
		return agents, nil
	}
	if err := cfg.Get("agents", &agents); err != nil {
		if isMissingKeyError(err) {
			return agents, nil
		}
		return nil, fmt.Errorf("agentci.LoadAgents: get agents: %w", err)
	}
	if agents == nil {
		agents = make(map[string]AgentConfig)
	}
	return agents, nil
}

// ListAgents returns all configured agents (active and inactive).
func ListAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	return LoadAgents(cfg)
}

// LoadActiveAgents returns only active agents.
func LoadActiveAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	agents, err := LoadAgents(cfg)
	if err != nil {
		return nil, err
	}
	active := make(map[string]AgentConfig)
	for name, agent := range agents {
		if agent.Active {
			active[name] = agent
		}
	}
	return active, nil
}

// LoadClothoConfig loads the Clotho orchestrator settings.
// Returns sensible defaults if no config is present.
func LoadClothoConfig(cfg *config.Config) (ClothoConfig, error) {
	clotho := defaultClothoConfig()
	if cfg == nil {
		return clotho, nil
	}
	var raw map[string]any
	if err := cfg.Get("clotho", &raw); err != nil {
		if isMissingKeyError(err) {
			return clotho, nil
		}
		return clotho, fmt.Errorf("agentci.LoadClothoConfig: get clotho: %w", err)
	}
	if err := cfg.Get("clotho", &clotho); err != nil {
		return clotho, fmt.Errorf("agentci.LoadClothoConfig: decode clotho: %w", err)
	}
	if raw == nil {
		raw = map[string]any{}
	}
	if clotho.Strategy == "" {
		clotho.Strategy = defaultClothoConfig().Strategy
	}
	if err := validateClothoStrategy(clotho.Strategy); err != nil {
		return clotho, err
	}
	if _, ok := raw["validation_threshold"]; !ok {
		clotho.ValidationThreshold = defaultClothoConfig().ValidationThreshold
	}
	if clotho.ValidationThreshold < 0 || clotho.ValidationThreshold > 1 {
		return clotho, fmt.Errorf(
			"agentci.LoadClothoConfig: validation_threshold must be between 0.0 and 1.0, got %v",
			clotho.ValidationThreshold,
		)
	}
	return clotho, nil
}

func validateClothoStrategy(strategy string) error {
	switch {
	case strategy == "":
		return nil
	case strings.EqualFold(strategy, "direct"):
		return nil
	case strings.EqualFold(strategy, "clotho-verified"):
		return nil
	default:
		return fmt.Errorf(
			"agentci.LoadClothoConfig: strategy must be direct or clotho-verified, got %q",
			strategy,
		)
	}
}

// SaveAgent writes an agent config entry to the config file.
func SaveAgent(cfg *config.Config, name string, ac AgentConfig) error {
	if cfg == nil {
		return errors.New("agentci.SaveAgent: config is required")
	}
	if name == "" {
		return errors.New("agentci.SaveAgent: name is required")
	}

	agents, err := LoadAgents(cfg)
	if err != nil {
		return fmt.Errorf("agentci.SaveAgent: load agents: %w", err)
	}
	if agents == nil {
		agents = make(map[string]AgentConfig)
	}
	agents[name] = ac
	if err := cfg.Set("agents", agents); err != nil {
		return fmt.Errorf("agentci.SaveAgent: set agents: %w", err)
	}
	return cfg.Commit()
}

// RemoveAgent removes an agent from the config file.
func RemoveAgent(cfg *config.Config, name string) error {
	if cfg == nil {
		return errors.New("agentci.RemoveAgent: config is required")
	}
	if name == "" {
		return errors.New("agentci.RemoveAgent: name is required")
	}

	agents, err := LoadAgents(cfg)
	if err != nil {
		return fmt.Errorf("agentci.RemoveAgent: load agents: %w", err)
	}
	delete(agents, name)
	if err := cfg.Set("agents", agents); err != nil {
		return fmt.Errorf("agentci.RemoveAgent: set agents: %w", err)
	}
	return cfg.Commit()
}

// MarshalYAML makes the config stable when written through generic YAML paths.
func (a AgentConfig) MarshalYAML() (any, error) {
	type alias AgentConfig
	return alias(a), nil
}

// UnmarshalYAML keeps the model permissive for YAML round-tripping.
func (a *AgentConfig) UnmarshalYAML(value *yaml.Node) error {
	type alias AgentConfig
	var out alias
	if err := value.Decode(&out); err != nil {
		return err
	}
	*a = AgentConfig(out)
	return nil
}

func isMissingKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "key not found")
}
