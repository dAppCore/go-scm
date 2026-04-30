// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	// Note: errors.New is retained for stable config validation errors.
	`errors`
	// Note: os is retained for environment and home-directory config discovery before a Core runtime exists.
	`os`
	// Note: filepath is retained for OS-specific config file paths.
	`path/filepath`

	"gopkg.in/yaml.v3"
)

func ResolveConfig(flagURL, flagToken string) (url, token string, err error)  /* v090-result-boundary */ {
	url, token = loadForgeConfigValues()
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

func loadForgeConfigValues() (string, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ""
	}
	raw, err := os.ReadFile(filepath.Join(home, ".core", "config.yaml"))
	if err != nil {
		return "", ""
	}
	var data map[string]any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return "", ""
	}
	forge, _ := data["forge"].(map[string]any)
	url, _ := forge["url"].(string)
	token, _ := forge["token"].(string)
	return url, token
}

func NewFromConfig(flagURL, flagToken string) (*Client, error)  /* v090-result-boundary */ {
	url, token, err := ResolveConfig(flagURL, flagToken)
	if err != nil {
		return nil, err
	}
	return New(url, token)
}

func SaveConfig(url, token string) error  /* v090-result-boundary */ {
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
