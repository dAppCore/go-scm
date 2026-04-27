// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	// Note: errors.New is retained for stable config validation errors.
	"errors"
	// Note: os is retained for environment and home-directory config discovery before a Core runtime exists.
	"os"
	// Note: filepath is retained for OS-specific config file paths.
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigKeyURL   = "gitea.url"
	ConfigKeyToken = "gitea.token"
	DefaultURL     = "https://gitea.snider.dev"
)

func ResolveConfig(flagURL, flagToken string) (url, token string, err error) {
	home, homeErr := os.UserHomeDir()
	if homeErr == nil {
		path := filepath.Join(home, ".core", "config.yaml")
		if raw, readErr := os.ReadFile(path); readErr == nil {
			var data map[string]any
			if yamlErr := yaml.Unmarshal(raw, &data); yamlErr == nil {
				if giteaCfg, ok := data["gitea"].(map[string]any); ok {
					if v, ok := giteaCfg["url"].(string); ok {
						url = v
					}
					if v, ok := giteaCfg["token"].(string); ok {
						token = v
					}
				}
			}
		}
	}

	if v := os.Getenv("GITEA_URL"); v != "" {
		url = v
	}
	if v := os.Getenv("GITEA_TOKEN"); v != "" {
		token = v
	}
	if flagURL != "" {
		url = flagURL
	}
	if flagToken != "" {
		token = flagToken
	}
	if url == "" {
		url = DefaultURL
	}
	return url, token, nil
}

func NewFromConfig(flagURL, flagToken string) (*Client, error) {
	url, token, err := ResolveConfig(flagURL, flagToken)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, errors.New("gitea.NewFromConfig: no API token configured")
	}
	return New(url, token)
}

func SaveConfig(url, token string) error {
	if url == "" && token == "" {
		return errors.New("gitea.SaveConfig: url or token required")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".core", "config.yaml")
	payload := map[string]any{"gitea": map[string]any{}}
	if url != "" {
		payload["gitea"].(map[string]any)["url"] = url
	}
	if token != "" {
		payload["gitea"].(map[string]any)["token"] = token
	}
	raw, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}
