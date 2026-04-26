package sdk

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

// Plugin is the main SDK entry point. Create one with New(), configure it
// with builder methods and options, then call Run() to start.
type Plugin struct {
	name    string
	version string
	config  pluginConfig

	manifest      manifestBuilder
	routes        []route
	trackHandlers []TrackHandler
	natsHandlers  []natsSubscription

	nc     *nats.Conn
	server *http.Server
	done   chan struct{}
}

// New creates a new Plugin with the given name, version, and options.
//
//	p := sdk.New("hello-world", "0.1.0",
//	    sdk.WithAuthor("CNAK Examples"),
//	    sdk.WithPermissions("tracks:read", "sidebar:register"),
//	)
func New(name, version string, opts ...Option) *Plugin {
	cfg := defaultConfig()

	// Environment overrides before options
	if url := os.Getenv("NATS_URL"); url != "" {
		cfg.natsURL = url
	}
	if port := os.Getenv("PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			cfg.port = p
		}
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Plugin{
		name:    name,
		version: version,
		config:  cfg,
		done:    make(chan struct{}),
	}
}

// Run starts the plugin: connects to NATS, registers, starts heartbeat,
// starts the HTTP server, and blocks until SIGTERM/SIGINT.
func (p *Plugin) Run() error {
	slog.Info("starting plugin", "name", p.name, "version", p.version, "port", p.config.port)

	// Connect to NATS
	if err := p.connectNATS(); err != nil {
		return fmt.Errorf("NATS connection failed: %w", err)
	}

	// Set up NATS subscriptions
	p.setupSubscriptions()

	// Register with CNAK backend (retries in background)
	go p.register()

	// Start heartbeat (re-publishes registration every 30s)
	p.startHeartbeat()

	// Listen for discovery requests
	p.startDiscoveryListener()

	// Start HTTP server
	if err := p.startHTTPServer(); err != nil {
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	slog.Info("plugin running", "name", p.name, "url", p.pluginURL())

	// Block until shutdown signal
	p.waitForShutdown()
	return nil
}

// Shutdown gracefully stops the plugin: deregisters, stops HTTP, closes NATS.
func (p *Plugin) Shutdown() {
	slog.Info("shutting down plugin", "name", p.name)

	// Signal done to stop heartbeat
	close(p.done)

	// Deregister from CNAK
	p.deregister()

	// Shutdown HTTP server
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.server.Shutdown(ctx)
	}
}

// NATS returns the underlying NATS connection for advanced use cases.
// Returns nil if called before Run().
func (p *Plugin) NATS() *nats.Conn {
	return p.nc
}

// pluginURL returns the URL other services should use to reach this plugin.
func (p *Plugin) pluginURL() string {
	if url := os.Getenv("PLUGIN_URL"); url != "" {
		return url
	}
	hostname, _ := os.Hostname()
	return fmt.Sprintf("http://%s:%d", hostname, p.config.port)
}

// waitForShutdown blocks until SIGTERM or SIGINT, then calls Shutdown.
func (p *Plugin) waitForShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	s := <-sig
	slog.Info("received signal", "signal", s)
	p.Shutdown()
}
