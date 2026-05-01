// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	core "dappco.re/go"
	"gopkg.in/yaml.v3"
)

func ResolveConfig(flagURL, flagToken string) (url, token string, err error)  /* v090-result-boundary */ {
	url, token = loadForgeConfigValues()
	if v := core.Getenv("FORGE_URL"); v != "" {
		url = v
	}
	if v := core.Getenv("FORGE_TOKEN"); v != "" {
		token = v
	}
	if flagURL != "" {
		url = flagURL
	}
	if flagToken != "" {
		token = flagToken
	}
	if url == "" || token == "" {
		return url, token, core.E("forge.ResolveConfig", "forge url and token are required", nil)
	}
	return url, token, nil
}

func loadForgeConfigValues() (string, string) {
	homeR := core.UserHomeDir()
	if !homeR.OK {
		return "", ""
	}
	rawR := core.ReadFile(core.PathJoin(homeR.Value.(string), ".core", "config.yaml"))
	if !rawR.OK {
		return "", ""
	}
	var data map[string]any
	if err := yaml.Unmarshal(rawR.Value.([]byte), &data); err != nil {
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
		return core.E("forge.SaveConfig", "url or token required", nil)
	}
	homeR := core.UserHomeDir()
	if !homeR.OK {
		return homeR.Value.(error)
	}
	path := core.PathJoin(homeR.Value.(string), ".core", "config.yaml")
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
	if r := core.MkdirAll(core.PathDir(path), 0o755); !r.OK {
		return r.Value.(error)
	}
	if r := core.WriteFile(path, raw, 0o600); !r.OK {
		return r.Value.(error)
	}
	return nil
}
