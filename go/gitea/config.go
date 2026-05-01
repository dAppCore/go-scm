// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	core "dappco.re/go"
	"gopkg.in/yaml.v3"
)

const (
	ConfigKeyURL   = "gitea.url"
	ConfigKeyToken = "gitea.token"
	DefaultURL     = "https://gitea.snider.dev"
)

func ResolveConfig(flagURL, flagToken string) (url, token string, err error)  /* v090-result-boundary */ {
	url, token = loadGiteaConfigValues()

	if v := core.Getenv("GITEA_URL"); v != "" {
		url = v
	}
	if v := core.Getenv("GITEA_TOKEN"); v != "" {
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

func loadGiteaConfigValues() (string, string) {
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
	giteaCfg, _ := data["gitea"].(map[string]any)
	url, _ := giteaCfg["url"].(string)
	token, _ := giteaCfg["token"].(string)
	return url, token
}

func NewFromConfig(flagURL, flagToken string) (*Client, error)  /* v090-result-boundary */ {
	url, token, err := ResolveConfig(flagURL, flagToken)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, core.E("gitea.NewFromConfig", "no API token configured", nil)
	}
	return New(url, token)
}

func SaveConfig(url, token string) error  /* v090-result-boundary */ {
	if url == "" && token == "" {
		return core.E("gitea.SaveConfig", "url or token required", nil)
	}
	homeR := core.UserHomeDir()
	if !homeR.OK {
		return homeR.Value.(error)
	}
	path := core.PathJoin(homeR.Value.(string), ".core", "config.yaml")
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
	if r := core.MkdirAll(core.PathDir(path), 0o755); !r.OK {
		return r.Value.(error)
	}
	if r := core.WriteFile(path, raw, 0o600); !r.OK {
		return r.Value.(error)
	}
	return nil
}
