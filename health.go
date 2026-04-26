package sdk

import (
	"encoding/json"
	"net/http"
)

// healthHandler returns the /health HTTP handler for the plugin.
func (p *Plugin) healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"plugin": p.name,
			"nats":   p.nc != nil && p.nc.IsConnected(),
		})
	}
}
