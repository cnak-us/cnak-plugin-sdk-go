package sdk

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// connectNATS establishes a NATS connection with retry and reconnect support.
func (p *Plugin) connectNATS() error {
	nc, err := nats.Connect(p.config.natsURL,
		nats.Name(fmt.Sprintf("plugin-%s", p.name)),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return fmt.Errorf("connecting to NATS at %s: %w", p.config.natsURL, err)
	}
	p.nc = nc
	slog.Info("connected to NATS", "url", p.config.natsURL, "plugin", p.name)
	return nil
}

// register sends the plugin registration to the CNAK backend via NATS request.
// Retries up to 30 times with 2s delay to handle backend startup ordering.
func (p *Plugin) register() {
	reg := PluginRegistration{
		Manifest: p.BuildManifest(),
		URL:      p.pluginURL(),
	}
	data, _ := json.Marshal(reg)

	for i := 0; i < 30; i++ {
		resp, err := p.nc.Request("cnak.plugin.register", data, 5*time.Second)
		if err == nil {
			slog.Info("registered with CNAK backend", "plugin", p.name, "response", string(resp.Data))
			return
		}
		slog.Debug("registration attempt failed", "plugin", p.name, "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}
	slog.Warn("failed to register after 30 attempts", "plugin", p.name)
}

// startHeartbeat re-publishes the registration every 30 seconds.
func (p *Plugin) startHeartbeat() {
	reg := PluginRegistration{
		Manifest: p.BuildManifest(),
		URL:      p.pluginURL(),
	}
	data, _ := json.Marshal(reg)

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := p.nc.Publish("cnak.plugin.register", data); err != nil {
					slog.Warn("heartbeat publish failed", "plugin", p.name, "error", err)
				}
			case <-p.done:
				return
			}
		}
	}()
}

// startDiscoveryListener subscribes to cnak.plugin.discover and re-announces.
func (p *Plugin) startDiscoveryListener() {
	reg := PluginRegistration{
		Manifest: p.BuildManifest(),
		URL:      p.pluginURL(),
	}
	data, _ := json.Marshal(reg)

	_, err := p.nc.Subscribe("cnak.plugin.discover", func(msg *nats.Msg) {
		slog.Info("discovery request received, re-announcing", "plugin", p.name)
		p.nc.Publish("cnak.plugin.register", data)
	})
	if err != nil {
		slog.Warn("failed to subscribe to discovery", "plugin", p.name, "error", err)
	}
}

// deregister publishes a deregistration message and drains the NATS connection.
func (p *Plugin) deregister() {
	if p.nc == nil || !p.nc.IsConnected() {
		return
	}
	dereg, _ := json.Marshal(map[string]string{"name": p.name})
	if err := p.nc.Publish("cnak.plugin.deregister", dereg); err != nil {
		slog.Warn("deregister publish failed", "plugin", p.name, "error", err)
	}
	p.nc.Flush()
	p.nc.Close()
	slog.Info("deregistered from CNAK", "plugin", p.name)
}
