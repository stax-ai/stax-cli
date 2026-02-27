// Package run resolves payload (file, string, or interactive), builds CloudEvents,
// and invokes the Python handler via the Knative func-python wrapper in a transient
// run directory (start server, POST CloudEvent, stop). Used by the run command.
package run

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/stax-ai/stax-cli/internal/automation"
)

const (
	defaultRunHost    = "127.0.0.1"
	defaultRunPort    = "8080"
	readinessEndpoint = "/health/readiness"
	readinessTimeout  = 60 * time.Second
	readinessInterval = 500 * time.Millisecond
)

// ResolvePayload returns the payload as a map from --payload-file or --payload.
// If both are empty, returns (nil, nil); the caller should use interactive or error.
// Order: payloadFile (if set) > payloadString (if set) > nil (caller should use interactive).
func ResolvePayload(payloadFile, payloadString string) (map[string]any, error) {
	if payloadFile != "" {
		data, err := os.ReadFile(payloadFile)
		if err != nil {
			return nil, fmt.Errorf("read payload file: %w", err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse payload JSON: %w", err)
		}
		return m, nil
	}
	if payloadString != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(payloadString), &m); err != nil {
			return nil, fmt.Errorf("parse payload JSON: %w", err)
		}
		return m, nil
	}
	return nil, nil
}

// CloudEvent is a minimal CloudEvents 1.0 structure (specversion, id, source, type, data).
// The CLI sends this as JSON to the handler's stdin.
type CloudEvent struct {
	SpecVersion string         `json:"specversion"`
	ID         string         `json:"id"`
	Source     string         `json:"source"`
	Type       string         `json:"type"`
	Data       map[string]any `json:"data,omitempty"`
}

// BuildCloudEvent creates a CloudEvent with data as the "data" field. If source or
// eventType are empty, defaults to "stax-cli/run" and "automation.trigger" respectively.
func BuildCloudEvent(data map[string]any, source, eventType string) CloudEvent {
	if source == "" {
		source = "stax-cli/run"
	}
	if eventType == "" {
		eventType = "automation.trigger"
	}
	return CloudEvent{
		SpecVersion: "1.0",
		ID:          newEventID(),
		Source:      source,
		Type:        eventType,
		Data:        data,
	}
}

// InvokeOptions configures handler invocation.
type InvokeOptions struct {
	ProjectRoot string   // Directory containing config.yml and main.py.
	Handler     string   // Module:function (e.g. "main:handler"); required for server-based run.
	Env         []string // Optional env vars; appended to os.Environ().
}

// Invoke starts the Knative func-python wrapper in a transient run directory,
// waits for readiness, POSTs the CloudEvent, then stops the process. Returns
// combined stdout/stderr from the server and the HTTP response body.
func Invoke(ctx context.Context, event CloudEvent, opts InvokeOptions) (stdout, stderr []byte, err error) {
	if opts.Handler == "" {
		return nil, nil, fmt.Errorf("handler is required (e.g. main:handler)")
	}
	mainPy := filepath.Join(opts.ProjectRoot, "main.py")
	if _, err := os.Stat(mainPy); err != nil {
		return nil, nil, fmt.Errorf("main.py not found in %s: %w", opts.ProjectRoot, err)
	}

	host := defaultRunHost
	port := defaultRunPort
	port, err = choosePort(host, port, false)
	if err != nil {
		return nil, nil, fmt.Errorf("choose port: %w", err)
	}

	runDir := RunDir(opts.ProjectRoot, port)
	if err := os.MkdirAll(filepath.Dir(runDir), 0755); err != nil {
		return nil, nil, fmt.Errorf("create run dir: %w", err)
	}
	if err := WriteRunScaffold(runDir, opts.ProjectRoot, opts.Handler); err != nil {
		return nil, nil, fmt.Errorf("scaffold run: %w", err)
	}
	defer os.RemoveAll(runDir)

	// venv
	venvCmd := exec.CommandContext(ctx, pythonCmd(), "-m", "venv", ".venv")
	venvCmd.Dir = runDir
	venvCmd.Stdout = os.Stdout
	venvCmd.Stderr = os.Stderr
	if err := venvCmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("create venv: %w", err)
	}

	pipPath := "pip"
	if os.PathSeparator == '\\' {
		pipPath = filepath.Join(".venv", "Scripts", "pip.exe")
	} else {
		pipPath = filepath.Join(".venv", "bin", "pip")
	}
	pipCmd := exec.CommandContext(ctx, filepath.Join(runDir, pipPath), "install", ".")
	pipCmd.Dir = runDir
	var pipOut, pipErr bytes.Buffer
	pipCmd.Stdout = &pipOut
	pipCmd.Stderr = &pipErr
	if err := pipCmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("pip install: %w\nstderr: %s", err, pipErr.Bytes())
	}

	listenAddr := net.JoinHostPort(host, port)
	pythonBin := filepath.Join(runDir, ".venv", "bin", "python")
	if os.PathSeparator == '\\' {
		pythonBin = filepath.Join(runDir, ".venv", "Scripts", "python.exe")
	}
	serverCmd := exec.CommandContext(ctx, pythonBin, "./service/main.py")
	serverCmd.Dir = runDir
	serverCmd.Env = append(os.Environ(), "PORT="+port, "LISTEN_ADDRESS="+listenAddr, "STAX_HANDLER="+opts.Handler, "PWD="+runDir)
	var outBuf, errBuf bytes.Buffer
	serverCmd.Stdout = &outBuf
	serverCmd.Stderr = &errBuf
	if err := serverCmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start server: %w", err)
	}
	defer func() { _ = serverCmd.Process.Signal(os.Interrupt) }()

	baseURL := "http://" + listenAddr
	if err := waitForReadiness(ctx, baseURL); err != nil {
		return outBuf.Bytes(), errBuf.Bytes(), fmt.Errorf("wait readiness: %w", err)
	}

	responseBody, postErr := postCloudEvent(ctx, baseURL, event)
	if postErr != nil {
		return outBuf.Bytes(), errBuf.Bytes(), fmt.Errorf("invoke: %w", postErr)
	}

	// Stop server
	_ = serverCmd.Process.Signal(os.Interrupt)
	_ = serverCmd.Wait()

	// Append response body to stdout so caller sees handler result
	if len(responseBody) > 0 {
		outBuf.WriteByte('\n')
		outBuf.Write(responseBody)
	}
	return outBuf.Bytes(), errBuf.Bytes(), nil
}

func pythonCmd() string {
	if _, err := exec.LookPath("python"); err != nil {
		return "python3"
	}
	return "python"
}

func choosePort(host, preferredPort string, explicitPort bool) (string, error) {
	l, err := net.Listen("tcp", net.JoinHostPort(host, preferredPort))
	if err == nil {
		l.Close()
		return preferredPort, nil
	}
	if explicitPort {
		return "", fmt.Errorf("port %s unavailable: %w", preferredPort, err)
	}
	l, err = net.Listen("tcp", net.JoinHostPort(host, ""))
	if err != nil {
		return "", fmt.Errorf("no port available: %w", err)
	}
	defer l.Close()
	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return "", err
	}
	return port, nil
}

func waitForReadiness(ctx context.Context, baseURL string) error {
	ctx, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()
	url := baseURL + readinessEndpoint
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(readinessInterval)
				continue
			}
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(readinessInterval)
	}
}

func postCloudEvent(ctx context.Context, baseURL string, event CloudEvent) ([]byte, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("Ce-Specversion", event.SpecVersion)
	req.Header.Set("Ce-Id", event.ID)
	req.Header.Set("Ce-Source", event.Source)
	req.Header.Set("Ce-Type", event.Type)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return buf.Bytes(), nil
}

func newEventID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("stax-%d", os.Getpid())
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Run loads config (from configPath or by FindAndLoad), resolves payload (or interactive),
// builds a CloudEvent, and invokes the handler. If cfg is nil, config is loaded from configPath
// or the current/parent dirs. When interactive is true and payload is nil, PayloadFromInteractive
// is used to prompt for each config field.
func Run(ctx context.Context, configPath string, payload map[string]any, interactive bool, cfg *automation.Config) (stdout, stderr []byte, err error) {
	if cfg == nil {
		var fpath string
		if configPath != "" {
			var loadErr error
			cfg, loadErr = automation.Load(configPath)
			if loadErr != nil {
				return nil, nil, fmt.Errorf("load config: %w", loadErr)
			}
			configPath = filepath.Dir(configPath)
		} else {
			var findErr error
			cfg, fpath, findErr = automation.FindAndLoad("")
			if findErr != nil {
				return nil, nil, fmt.Errorf("find config: %w", findErr)
			}
			configPath = filepath.Dir(fpath)
		}
	}
	if err := automation.Validate(cfg); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %w", err)
	}

	if payload == nil && interactive {
		var interErr error
		payload, interErr = PayloadFromInteractive(cfg)
		if interErr != nil {
			return nil, nil, interErr
		}
	}
	if payload == nil {
		payload = make(map[string]any)
	}

	event := BuildCloudEvent(payload, "stax-cli/run", "automation.trigger")
	return Invoke(ctx, event, InvokeOptions{ProjectRoot: configPath, Handler: cfg.Handler})
}
