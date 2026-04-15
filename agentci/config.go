// SPDX-License-Identifier: EUPL-1.2

// Package agentci provides configuration, security, and orchestration for AgentCI dispatch targets.
package agentci

import (
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"
	"strconv"
	stdstrings "strings"

	"dappco.re/go/core/config"
	coreerr "dappco.re/go/core/log"
)

// AgentConfig represents a single agent machine in the config file.
type AgentConfig struct {
	Host          string   `yaml:"host" mapstructure:"host"`
	QueueDir      string   `yaml:"queue_dir" mapstructure:"queue_dir"`
	ForgejoUser   string   `yaml:"forgejo_user" mapstructure:"forgejo_user"`
	Model         string   `yaml:"model" mapstructure:"model"`                   // primary AI model
	Runner        string   `yaml:"runner" mapstructure:"runner"`                 // runner binary: claude, codex, gemini
	VerifyModel   string   `yaml:"verify_model" mapstructure:"verify_model"`     // secondary model for dual-run
	SecurityLevel string   `yaml:"security_level" mapstructure:"security_level"` // low, high
	Roles         []string `yaml:"roles" mapstructure:"roles"`
	DualRun       bool     `yaml:"dual_run" mapstructure:"dual_run"`
	Active        bool     `yaml:"active" mapstructure:"active"`
}

// ClothoConfig controls the orchestration strategy.
type ClothoConfig struct {
	Strategy            string  `yaml:"strategy" mapstructure:"strategy"`                         // direct, clotho-verified
	ValidationThreshold float64 `yaml:"validation_threshold" mapstructure:"validation_threshold"` // divergence limit (0.0-1.0)
	SigningKeyPath      string  `yaml:"signing_key_path" mapstructure:"signing_key_path"`
}

// LoadAgents reads agent targets from config and returns a map of AgentConfig.
// Returns an empty map (not an error) if no agents are configured.
// Usage: LoadAgents(...)
func LoadAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	if cfg == nil {
		return map[string]AgentConfig{}, nil
	}
	var agents map[string]AgentConfig
	if err := cfg.Get("agentci.agents", &agents); err != nil {
		if stdstrings.Contains(err.Error(), "key not found") {
			return map[string]AgentConfig{}, nil
		}
		return nil, coreerr.E("agentci.LoadAgents", "load agents", err)
	}
	if agents == nil {
		return map[string]AgentConfig{}, nil
	}

	// Validate and apply defaults.
	for name, ac := range agents {
		if ac.QueueDir == "" {
			ac.QueueDir = "/home/claude/ai-work/queue"
		}
		if ac.Model == "" {
			ac.Model = "sonnet"
		}
		if ac.Runner == "" {
			ac.Runner = "claude"
		}
		if ac.Active && ac.Host == "" {
			return nil, coreerr.E("agentci.LoadAgents", "agent "+name+": host is required", nil)
		}
		agents[name] = ac
	}

	return agents, nil
}

// LoadActiveAgents returns only active agents.
// Usage: LoadActiveAgents(...)
func LoadActiveAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	if cfg == nil {
		return map[string]AgentConfig{}, nil
	}
	all, err := LoadAgents(cfg)
	if err != nil {
		return nil, err
	}
	active := make(map[string]AgentConfig)
	for name, ac := range all {
		if ac.Active {
			active[name] = ac
		}
	}
	return active, nil
}

// LoadClothoConfig loads the Clotho orchestrator settings.
// Returns sensible defaults if no config is present.
// Usage: LoadClothoConfig(...)
func LoadClothoConfig(cfg *config.Config) (ClothoConfig, error) {
	cc := ClothoConfig{
		Strategy:            "direct",
		ValidationThreshold: 0.85,
	}

	if cfg == nil {
		return cc, nil
	}

	var raw map[string]any
	if err := cfg.Get("agentci.clotho", &raw); err != nil {
		if stdstrings.Contains(err.Error(), "key not found") {
			return cc, nil
		}
		return ClothoConfig{}, coreerr.E("agentci.LoadClothoConfig", "load clotho config", err)
	}

	if strategy, ok := raw["strategy"].(string); ok && strategy != "" {
		cc.Strategy = stdstrings.TrimSpace(strategy)
	}
	if threshold, ok := raw["validation_threshold"]; ok {
		switch v := threshold.(type) {
		case float64:
			cc.ValidationThreshold = v
		case float32:
			cc.ValidationThreshold = float64(v)
		case int:
			cc.ValidationThreshold = float64(v)
		case int64:
			cc.ValidationThreshold = float64(v)
		case uint64:
			cc.ValidationThreshold = float64(v)
		case string:
			trimmed := stdstrings.TrimSpace(v)
			if trimmed == "" {
				break
			}
			parsed, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return ClothoConfig{}, coreerr.E("agentci.LoadClothoConfig", "parse validation threshold", err)
			}
			cc.ValidationThreshold = parsed
		}
	}
	if signKeyPath, ok := raw["signing_key_path"].(string); ok {
		cc.SigningKeyPath = stdstrings.TrimSpace(signKeyPath)
	}
	return cc, nil
}

// SaveAgent writes an agent config entry to the config file.
// Usage: SaveAgent(...)
func SaveAgent(cfg *config.Config, name string, ac AgentConfig) error {
	if cfg == nil {
		return coreerr.E("agentci.SaveAgent", "config is required", nil)
	}
	key := fmt.Sprintf("agentci.agents.%s", name)
	data := map[string]any{
		"host":         ac.Host,
		"queue_dir":    ac.QueueDir,
		"forgejo_user": ac.ForgejoUser,
		"active":       ac.Active,
		"dual_run":     ac.DualRun,
	}
	if ac.Model != "" {
		data["model"] = ac.Model
	}
	if ac.Runner != "" {
		data["runner"] = ac.Runner
	}
	if ac.VerifyModel != "" {
		data["verify_model"] = ac.VerifyModel
	}
	if ac.SecurityLevel != "" {
		data["security_level"] = ac.SecurityLevel
	}
	if len(ac.Roles) > 0 {
		data["roles"] = ac.Roles
	}
	return cfg.Set(key, data)
}

// RemoveAgent removes an agent from the config file.
// Usage: RemoveAgent(...)
func RemoveAgent(cfg *config.Config, name string) error {
	if cfg == nil {
		return coreerr.E("agentci.RemoveAgent", "config is required", nil)
	}
	var agents map[string]AgentConfig
	if err := cfg.Get("agentci.agents", &agents); err != nil {
		if stdstrings.Contains(err.Error(), "key not found") {
			return coreerr.E("agentci.RemoveAgent", "no agents configured", nil)
		}
		return coreerr.E("agentci.RemoveAgent", "load agents", err)
	}
	if _, ok := agents[name]; !ok {
		return coreerr.E("agentci.RemoveAgent", "agent not found: "+name, nil)
	}
	delete(agents, name)
	return cfg.Set("agentci.agents", agents)
}

// ListAgents returns all configured agents (active and inactive).
// Usage: ListAgents(...)
func ListAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	if cfg == nil {
		return map[string]AgentConfig{}, nil
	}
	var agents map[string]AgentConfig
	if err := cfg.Get("agentci.agents", &agents); err != nil {
		if stdstrings.Contains(err.Error(), "key not found") {
			return map[string]AgentConfig{}, nil
		}
		return nil, coreerr.E("agentci.ListAgents", "load agents", err)
	}
	if agents == nil {
		return map[string]AgentConfig{}, nil
	}
	return agents, nil
}
