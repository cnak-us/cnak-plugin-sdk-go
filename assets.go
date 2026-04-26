package sdk

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// mountAssets serves static files from the configured assets directory under /assets/*.
func (p *Plugin) mountAssets(r chi.Router) {
	fs := http.StripPrefix("/assets/", http.FileServer(http.Dir(p.config.assetsDir)))
	r.Handle("/assets/*", fs)
}
