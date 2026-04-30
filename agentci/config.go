// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	// Note: errors.New is retained for stable validation errors in the agentci config API.
	"errors"
	// Note: fmt.Errorf is retained for wrapped config load/save errors; replacing here would add churn across load-bearing config paths.
	"fmt"
	// Note: strings helpers are retained for case-insensitive strategy matching and config error classification.
	"strings"

	"dappco.re/go/config"
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
	if r := cfg.Get("agents", &agents); !r.OK {
		if isMissingKeyError(r.Error()) {
			return agents, nil
		}
		return nil, fmt.Errorf("agentci.LoadAgents: get agents: %s", r.Error())
	}
	if agents == nil {
		agents = make(map[string]AgentConfig)
	}
	return cloneAgents(agents), nil
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
	if r := cfg.Get("clotho", &raw); !r.OK {
		if isMissingKeyError(r.Error()) {
			return clotho, nil
		}
		return clotho, fmt.Errorf("agentci.LoadClothoConfig: get clotho: %s", r.Error())
	}
	if r := cfg.Get("clotho", &clotho); !r.OK {
		return clotho, fmt.Errorf("agentci.LoadClothoConfig: decode clotho: %s", r.Error())
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
	if r := cfg.Set("agents", agents); !r.OK {
		return fmt.Errorf("agentci.SaveAgent: set agents: %s", r.Error())
	}
	if r := cfg.Commit(); !r.OK {
		return fmt.Errorf("agentci.SaveAgent: commit: %s", r.Error())
	}
	return nil
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
	if r := cfg.Set("agents", agents); !r.OK {
		return fmt.Errorf("agentci.RemoveAgent: set agents: %s", r.Error())
	}
	if r := cfg.Commit(); !r.OK {
		return fmt.Errorf("agentci.RemoveAgent: commit: %s", r.Error())
	}
	return nil
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

func isMissingKeyError(msg string) bool {
	return strings.Contains(msg, "key not found")
}

func cloneAgents(src map[string]AgentConfig) map[string]AgentConfig {
	if len(src) == 0 {
		return make(map[string]AgentConfig)
	}
	dst := make(map[string]AgentConfig, len(src))
	for name, agent := range src {
		if len(agent.Roles) > 0 {
			agent.Roles = append([]string(nil), agent.Roles...)
		}
		dst[name] = agent
	}
	return dst
}
