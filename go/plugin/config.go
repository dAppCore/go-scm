// SPDX-License-Identifier: EUPL-1.2

package plugin

type PluginConfig struct {
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Source      string `json:"source" yaml:"source"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	InstalledAt string `json:"installed_at" yaml:"installed_at"`
}
