package run

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stax-ai/stax-cli/internal/automation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePayload_FromFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "p.json")
	err := os.WriteFile(f, []byte(`{"a":1,"b":"two"}`), 0644)
	require.NoError(t, err)

	m, err := ResolvePayload(f, "")
	require.NoError(t, err)
	assert.Equal(t, float64(1), m["a"])
	assert.Equal(t, "two", m["b"])
}

func TestResolvePayload_FromString(t *testing.T) {
	m, err := ResolvePayload("", `{"x": true}`)
	require.NoError(t, err)
	assert.Equal(t, true, m["x"])
}

func TestResolvePayload_EmptyReturnsNil(t *testing.T) {
	m, err := ResolvePayload("", "")
	require.NoError(t, err)
	assert.Nil(t, m)
}

func TestResolvePayload_InvalidJSON(t *testing.T) {
	_, err := ResolvePayload("", "not json")
	require.Error(t, err)
}

func TestBuildCloudEvent(t *testing.T) {
	ev := BuildCloudEvent(map[string]any{"k": "v"}, "", "")
	assert.Equal(t, "1.0", ev.SpecVersion)
	assert.NotEmpty(t, ev.ID)
	assert.Equal(t, "stax-cli/run", ev.Source)
	assert.Equal(t, "automation.trigger", ev.Type)
	assert.Equal(t, map[string]any{"k": "v"}, ev.Data)
}

func TestBuildCloudEvent_CustomSourceType(t *testing.T) {
	ev := BuildCloudEvent(nil, "my/source", "my.type")
	assert.Equal(t, "my/source", ev.Source)
	assert.Equal(t, "my.type", ev.Type)
}

func TestPayloadFromInteractive_EmptyConfig(t *testing.T) {
	m, err := PayloadFromInteractive(nil)
	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.Empty(t, m)
}

func TestParseFieldValue(t *testing.T) {
	tests := []struct {
		raw  string
		typ  string
		want any
	}{
		{"hello", automation.FieldTypeString, "hello"},
		{"42", automation.FieldTypeInteger, int64(42)},
		{"3.14", automation.FieldTypeNumber, 3.14},
		{"true", automation.FieldTypeBoolean, true},
		{"false", automation.FieldTypeBoolean, false},
	}
	for _, tt := range tests {
		got, err := parseFieldValue(tt.raw, tt.typ)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got)
	}
}

func TestInvoke_NoMainPy(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	ev := BuildCloudEvent(map[string]any{}, "", "")

	_, _, err := Invoke(ctx, ev, InvokeOptions{ProjectRoot: dir, Handler: "main:handler"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "main.py")
}

func TestInvoke_NoHandler(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	ev := BuildCloudEvent(map[string]any{}, "", "")

	_, _, err := Invoke(ctx, ev, InvokeOptions{ProjectRoot: dir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler is required")
}

func TestRun_RequiresPayloadOrInteractive(t *testing.T) {
	// Run with nil payload and interactive=false: should still work with empty payload
	// if config is found. We need a fixture dir with config.yml and main.py to actually invoke.
	// Here we only test that Run returns error when config is not found.
	ctx := context.Background()
	dir := t.TempDir()
	_, _, err := Run(ctx, filepath.Join(dir, "config.yml"), nil, false, nil)
	require.Error(t, err)
}

func TestCloudEventJSONRoundtrip(t *testing.T) {
	ev := BuildCloudEvent(map[string]any{"a": 1}, "s", "t")
	b, err := json.Marshal(ev)
	require.NoError(t, err)
	var decoded CloudEvent
	err = json.Unmarshal(b, &decoded)
	require.NoError(t, err)
	assert.Equal(t, ev.SpecVersion, decoded.SpecVersion)
	assert.Equal(t, ev.ID, decoded.ID)
	assert.Equal(t, ev.Source, decoded.Source)
	assert.Equal(t, ev.Type, decoded.Type)
	// JSON decodes numbers as float64
	assert.Equal(t, float64(1), decoded.Data["a"])
}

func TestChoosePort_ReturnsPort(t *testing.T) {
	port, err := choosePort("127.0.0.1", defaultRunPort, false)
	require.NoError(t, err)
	assert.NotEmpty(t, port)
}

func TestChoosePort_FallbackWhenDefaultInUse(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	_, taken, _ := net.SplitHostPort(l.Addr().String())

	port, err := choosePort("127.0.0.1", taken, false)
	require.NoError(t, err)
	assert.NotEqual(t, taken, port)
	assert.NotEmpty(t, port)
}

func TestWaitForReadiness_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == readinessEndpoint {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	ctx := context.Background()
	err := waitForReadiness(ctx, srv.URL)
	require.NoError(t, err)
}

func TestPostCloudEvent_Success(t *testing.T) {
	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	ctx := context.Background()
	ev := BuildCloudEvent(map[string]any{"x": 1}, "s", "t")
	body, err := postCloudEvent(ctx, srv.URL, ev)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"ok":true`)
	assert.Contains(t, string(received), `"data"`)
}
