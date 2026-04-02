// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dappco.re/go/core/scm/marketplace"
	scmapi "dappco.re/go/core/scm/pkg/api"
	"dappco.re/go/core/ws"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInstaller struct {
	updateCalls []string
}

func (f *fakeInstaller) Install(context.Context, marketplace.Module) error { return nil }

func (f *fakeInstaller) Remove(string) error { return nil }

func (f *fakeInstaller) Update(_ context.Context, code string) error {
	f.updateCalls = append(f.updateCalls, code)
	return nil
}

func (f *fakeInstaller) Installed() ([]marketplace.InstalledModule, error) { return nil, nil }

func TestScmProvider_UpdateInstalled_EmitsInstalledChangedEvent_Good(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go hub.Run(ctx)

	server := httptest.NewServer(hub.Handler())
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.WriteJSON(ws.Message{Type: ws.TypeSubscribe, Data: "scm.installed.changed"}))
	time.Sleep(50 * time.Millisecond)

	installer := &fakeInstaller{}
	p := scmapi.NewProvider(nil, installer, nil, hub)
	r := setupRouter(p)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scm/installed/demo/update", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, []string{"demo"}, installer.updateCalls)

	var msg ws.Message
	require.NoError(t, conn.ReadJSON(&msg))
	assert.Equal(t, ws.TypeEvent, msg.Type)
	assert.Equal(t, "scm.installed.changed", msg.Channel)

	data, ok := msg.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "updated", data["action"])
	assert.Equal(t, "demo", data["code"])
}
