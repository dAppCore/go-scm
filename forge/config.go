// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ResolveConfig(flagURL, flagToken string) (url, token string, err error) {
	home, homeErr := os.UserHomeDir()
	if homeErr == nil {
		path := filepath.Join(home, ".core", "config.yaml")
		if raw, readErr := os.ReadFile(path); readErr == nil {
			var data map[string]any
			if yamlErr := yaml.Unmarshal(raw, &data); yamlErr == nil {
				if forge, ok := data["forge"].(map[string]any); ok {
					if v, ok := forge["url"].(string); ok {
						url = v
					}
					if v, ok := forge["token"].(string); ok {
						token = v
					}
				}
			}
		}
	}
	if v := os.Getenv("FORGE_URL"); v != "" {
		url = v
	}
	if v := os.Getenv("FORGE_TOKEN"); v != "" {
		token = v
	}
	if flagURL != "" {
		url = flagURL
	}
	if flagToken != "" {
		token = flagToken
	}
	if url == "" || token == "" {
		return url, token, errors.New("forge.ResolveConfig: forge url and token are required")
	}
	return url, token, nil
}

func NewFromConfig(flagURL, flagToken string) (*Client, error) {
	url, token, err := ResolveConfig(flagURL, flagToken)
	if err != nil {
		return nil, err
	}
	return New(url, token)
}

func SaveConfig(url, token string) error {
	if url == "" && token == "" {
		return errors.New("forge.SaveConfig: url or token required")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".core", "config.yaml")
	payload := map[string]any{"forge": map[string]any{}}
	if url != "" {
		payload["forge"].(map[string]any)["url"] = url
	}
	if token != "" {
		payload["forge"].(map[string]any)["token"] = token
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
