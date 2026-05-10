package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
