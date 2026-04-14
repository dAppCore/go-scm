// SPDX-License-Identifier: EUPL-1.2

package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Good(t *testing.T) {
	raw := `
code: photo-browser
name: Photo Browser
version: 0.1.0
sign: dGVzdHNpZw==

layout: HLCRF
slots:
  H: nav-breadcrumb
  L: folder-tree
  C: photo-grid
  R: metadata-panel
  F: status-bar

permissions:
  read: ["./photos/"]
  write: []
  net: []
  run: []

modules:
  - core/media
  - core/fs
`
	m, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "photo-browser", m.Code)
	assert.Equal(t, "Photo Browser", m.Name)
	assert.Equal(t, "0.1.0", m.Version)
	assert.Equal(t, "dGVzdHNpZw==", m.Sign)
	assert.Equal(t, "HLCRF", m.Layout)
	assert.Equal(t, "nav-breadcrumb", m.Slots["H"])
	assert.Equal(t, "photo-grid", m.Slots["C"])
	assert.Len(t, m.Permissions.Read, 1)
	assert.Equal(t, "./photos/", m.Permissions.Read[0])
	assert.Len(t, m.Modules, 2)
}

func TestParse_Good_WithSignKey_Good(t *testing.T) {
	raw := `
code: signed-module
name: Signed Module
version: 1.0.0
sign: dGVzdHNpZw==
sign_key: 302a300506032b6570032100f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3
`

	m, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "signed-module", m.Code)
	assert.Equal(t, "dGVzdHNpZw==", m.Sign)
	assert.Equal(t, "302a300506032b6570032100f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3", m.SignKey)
}

func TestParse_Bad(t *testing.T) {
	_, err := Parse([]byte("not: valid: yaml: ["))
	assert.Error(t, err)
}

func TestManifest_SlotNames_Good(t *testing.T) {
	m := Manifest{
		Slots: map[string]string{
			"H": "nav-bar",
			"C": "main-content",
		},
	}
	names := m.SlotNames()
	assert.Contains(t, names, "nav-bar")
	assert.Contains(t, names, "main-content")
	assert.Len(t, names, 2)
}

func TestParse_Good_WithDaemons_Good(t *testing.T) {
	raw := `
code: my-service
name: My Service
description: A test service with daemons
version: 1.0.0

daemons:
  api:
    binary: ./bin/api
    args: ["--port", "8080"]
    health: /healthz
    default: true
  worker:
    binary: ./bin/worker
    args: ["--concurrency", "4"]
    health: /ready
`
	m, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "my-service", m.Code)
	assert.Equal(t, "My Service", m.Name)
	assert.Equal(t, "A test service with daemons", m.Description)
	assert.Equal(t, "1.0.0", m.Version)
	assert.Len(t, m.Daemons, 2)

	api, ok := m.Daemons["api"]
	require.True(t, ok)
	assert.Equal(t, "./bin/api", api.Binary)
	assert.Equal(t, []string{"--port", "8080"}, api.Args)
	assert.Equal(t, "/healthz", api.Health)
	assert.True(t, api.Default)

	worker, ok := m.Daemons["worker"]
	require.True(t, ok)
	assert.Equal(t, "./bin/worker", worker.Binary)
	assert.Equal(t, []string{"--concurrency", "4"}, worker.Args)
	assert.Equal(t, "/ready", worker.Health)
	assert.False(t, worker.Default)
}

func TestManifest_DefaultDaemon_Good(t *testing.T) {
	m := Manifest{
		Daemons: map[string]DaemonSpec{
			"api": {
				Binary:  "./bin/api",
				Default: true,
			},
			"worker": {
				Binary: "./bin/worker",
			},
		},
	}
	name, spec, ok := m.DefaultDaemon()
	assert.True(t, ok)
	assert.Equal(t, "api", name)
	assert.Equal(t, "./bin/api", spec.Binary)
	assert.True(t, spec.Default)
}

func TestManifest_DefaultDaemon_Bad_NoDaemons_Good(t *testing.T) {
	m := Manifest{}
	name, spec, ok := m.DefaultDaemon()
	assert.False(t, ok)
	assert.Empty(t, name)
	assert.Empty(t, spec.Binary)
}

func TestManifest_DefaultDaemon_Bad_MultipleDefaults_Good(t *testing.T) {
	m := Manifest{
		Daemons: map[string]DaemonSpec{
			"api":    {Binary: "./bin/api", Default: true},
			"worker": {Binary: "./bin/worker", Default: true},
		},
	}
	_, _, ok := m.DefaultDaemon()
	assert.False(t, ok)
}

func TestManifest_DefaultDaemon_Bad_MultipleNoneDefault_Good(t *testing.T) {
	m := Manifest{
		Daemons: map[string]DaemonSpec{
			"api":    {Binary: "./bin/api"},
			"worker": {Binary: "./bin/worker"},
		},
	}
	_, _, ok := m.DefaultDaemon()
	assert.False(t, ok)
}

func TestParse_Good_WithProviderFields_Good(t *testing.T) {
	raw := `
code: cool-widget
name: Cool Widget Dashboard
version: 1.0.0
author: someone
licence: EUPL-1.2

namespace: /api/v1/cool-widget
port: 0
binary: ./cool-widget
args: ["--verbose"]

element:
  tag: core-cool-widget
  source: ./assets/core-cool-widget.js

spec: ./openapi.json

layout: HCF
slots:
  H: toolbar
  C: dashboard
  F: status
`
	m, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "cool-widget", m.Code)
	assert.Equal(t, "Cool Widget Dashboard", m.Name)
	assert.Equal(t, "1.0.0", m.Version)
	assert.Equal(t, "someone", m.Author)
	assert.Equal(t, "EUPL-1.2", m.Licence)
	assert.Equal(t, "/api/v1/cool-widget", m.Namespace)
	assert.Equal(t, 0, m.Port)
	assert.Equal(t, "./cool-widget", m.Binary)
	assert.Equal(t, []string{"--verbose"}, m.Args)
	assert.Equal(t, "./openapi.json", m.Spec)
	require.NotNil(t, m.Element)
	assert.Equal(t, "core-cool-widget", m.Element.Tag)
	assert.Equal(t, "./assets/core-cool-widget.js", m.Element.Source)
	assert.True(t, m.IsProvider())
}

func TestManifest_IsProvider_Good(t *testing.T) {
	m := Manifest{Namespace: "/api/v1/test", Binary: "./test"}
	assert.True(t, m.IsProvider())
}

func TestManifest_IsProvider_Bad_NoNamespace_Good(t *testing.T) {
	m := Manifest{Binary: "./test"}
	assert.False(t, m.IsProvider())
}

func TestManifest_IsProvider_Bad_NoBinary_Good(t *testing.T) {
	m := Manifest{Namespace: "/api/v1/test"}
	assert.False(t, m.IsProvider())
}

func TestManifest_IsProvider_Bad_Empty_Good(t *testing.T) {
	m := Manifest{}
	assert.False(t, m.IsProvider())
}

func TestManifest_DefaultDaemon_Good_SingleImplicit_Good(t *testing.T) {
	m := Manifest{
		Daemons: map[string]DaemonSpec{
			"server": {
				Binary: "./bin/server",
				Args:   []string{"--port", "3000"},
			},
		},
	}
	name, spec, ok := m.DefaultDaemon()
	assert.True(t, ok)
	assert.Equal(t, "server", name)
	assert.Equal(t, "./bin/server", spec.Binary)
	assert.False(t, spec.Default)
}
