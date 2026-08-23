package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// credentialBootstrapRequest is sent to the backend to request NATS credentials.
type credentialBootstrapRequest struct {
	PluginID    string   `json:"pluginId"`
	Permissions []string `json:"permissions"`
}

// credentialBootstrapResponse is returned by the backend with NATS credentials.
type credentialBootstrapResponse struct {
	CredsFile string `json:"credsFile"`
}

// registerOverHTTP announces this plugin's manifest to the backend over the
// service-token endpoint, before any NATS connection exists.
//
// This has to happen first in JWT mode. The backend mints a plugin's NATS
// credentials from its *registered* manifest, so a plugin that has only ever
// registered over NATS can never obtain the credentials it needs to reach
// NATS in the first place. Registering over HTTP breaks that loop; the NATS
// registration in Run() still follows and keeps the heartbeat path unchanged.
//
// Best-effort by design: on a backend that predates the endpoint this 404s,
// and the plugin carries on exactly as it did before.
func (p *Plugin) registerOverHTTP() error {
	if p.config.backendURL == "" || p.config.serviceToken == "" {
		return fmt.Errorf("BACKEND_URL and SERVICE_TOKEN required for HTTP registration")
	}

	body, err := json.Marshal(PluginRegistration{
		Manifest: p.BuildManifest(),
		URL:      p.pluginURL(),
	})
	if err != nil {
		return fmt.Errorf("marshalling registration: %w", err)
	}

	url := fmt.Sprintf("%s/internal/plugins/register", p.config.backendURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", p.config.serviceToken)

	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	for attempt := 0; attempt < 15; attempt++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("registration request failed after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("registration returned status %d: %s",
			resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	slog.Info("registered with CNAK backend over HTTP", "plugin", p.name)
	return nil
}

// bootstrapNATSCredentials requests scoped NATS JWT credentials from the backend.
// Called automatically during Run() when no NATS auth is configured and
// SERVICE_TOKEN + BACKEND_URL are available. The backend issues credentials
// scoped to the plugin's declared permissions.
//
// Flow:
//  1. Plugin POSTs manifest permissions + plugin ID to backend
//  2. Backend verifies service token, generates scoped NATS user JWT
//  3. Backend returns .creds file content
//  4. Plugin writes to temp file, sets config.natsCredsFile
func (p *Plugin) bootstrapNATSCredentials() error {
	if p.config.backendURL == "" || p.config.serviceToken == "" {
		return fmt.Errorf("BACKEND_URL and SERVICE_TOKEN required for credential bootstrapping")
	}

	reqBody := credentialBootstrapRequest{
		PluginID:    p.name,
		Permissions: p.config.permissions,
	}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/internal/plugins/credentials", p.config.backendURL)
	slog.Info("bootstrapping NATS credentials from backend",
		"plugin", p.name, "url", url)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating bootstrap request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", p.config.serviceToken)

	// Retry — backend may not be ready yet
	var resp *http.Response
	for attempt := 0; attempt < 15; attempt++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		slog.Debug("credential bootstrap attempt failed",
			"plugin", p.name, "attempt", attempt+1, "error", err)
		time.Sleep(2 * time.Second)

		// Rebuild request body for retry
		req, _ = http.NewRequest("POST", url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Service-Token", p.config.serviceToken)
	}
	if err != nil {
		return fmt.Errorf("bootstrap request failed after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bootstrap returned status %d", resp.StatusCode)
	}

	var creds credentialBootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return fmt.Errorf("decoding bootstrap response: %w", err)
	}

	// Write creds to a temp file
	credsDir := os.TempDir()
	credsPath := filepath.Join(credsDir, fmt.Sprintf("plugin-%s.creds", p.name))
	if err := os.WriteFile(credsPath, []byte(creds.CredsFile), 0600); err != nil {
		return fmt.Errorf("writing credentials file: %w", err)
	}

	p.config.natsCredsFile = credsPath
	slog.Info("NATS credentials bootstrapped",
		"plugin", p.name, "credsFile", credsPath)
	return nil
}

// hasNATSAuth returns true if any NATS auth mechanism is configured.
func (p *Plugin) hasNATSAuth() bool {
	return p.config.natsCredsFile != "" ||
		p.config.natsNKeySeed != "" ||
		p.config.natsAuthToken != ""
}
