package sdk

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type route struct {
	pattern string
	handler http.HandlerFunc
}

// HandleFunc registers an HTTP handler on the plugin's router.
// Call this before Run().
func (p *Plugin) HandleFunc(pattern string, handler http.HandlerFunc) *Plugin {
	p.routes = append(p.routes, route{pattern: pattern, handler: handler})
	return p
}

// buildRouter constructs the chi router with health, assets, and user routes.
func (p *Plugin) buildRouter() http.Handler {
	r := chi.NewRouter()

	// Built-in routes
	r.Get("/health", p.healthHandler())
	p.mountAssets(r)

	// User-defined routes
	for _, rt := range p.routes {
		r.HandleFunc(rt.pattern, rt.handler)
	}

	return r
}

// startHTTPServer starts the HTTP server in a goroutine.
func (p *Plugin) startHTTPServer() error {
	addr := fmt.Sprintf(":%d", p.config.port)
	p.server = &http.Server{
		Addr:    addr,
		Handler: p.buildRouter(),
	}

	go func() {
		slog.Info("HTTP server starting", "addr", addr, "plugin", p.name)
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}
