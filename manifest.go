package sdk

// PluginManifest matches the CNAK PluginManifest type (apiVersion, kind, metadata, spec).
type PluginManifest struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   PluginMetadata `json:"metadata"`
	Spec       PluginSpec     `json:"spec"`
}

// PluginMetadata contains identifying information.
type PluginMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// PluginSpec contains the plugin configuration.
type PluginSpec struct {
	MinCnakVersion string            `json:"minCnakVersion,omitempty"`
	Permissions    PluginPermissions `json:"permissions"`
	Backend        PluginBackend     `json:"backend"`
	Frontend       PluginFrontend    `json:"frontend"`
}

// PluginPermissions declares required permissions.
type PluginPermissions struct {
	Required []string `json:"required,omitempty"`
}

// PluginBackend describes the backend sidecar.
type PluginBackend struct {
	Port       int    `json:"port,omitempty"`
	HealthPath string `json:"healthPath,omitempty"`
}

// PluginFrontend describes frontend extension points.
type PluginFrontend struct {
	Assets              []string                   `json:"assets,omitempty"`
	Sidebar             []PluginSidebarItem        `json:"sidebar,omitempty"`
	MapClickHandlers    []PluginMapClickHandler    `json:"mapClickHandlers,omitempty"`
	TrackDetailSections []PluginTrackDetailSection `json:"trackDetailSections,omitempty"`
}

// PluginSidebarItem describes a sidebar navigation entry.
type PluginSidebarItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
	Route string `json:"route"`
}

// PluginMapClickHandler describes a map click action.
type PluginMapClickHandler struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// PluginTrackDetailSection describes a track detail panel extension.
type PluginTrackDetailSection struct {
	ID string `json:"id"`
}

// PluginRegistration is the message published to cnak.plugin.register.
type PluginRegistration struct {
	Manifest PluginManifest `json:"manifest"`
	URL      string         `json:"url"`
}

// manifestBuilder accumulates frontend extension points.
type manifestBuilder struct {
	sidebar             []PluginSidebarItem
	mapClickHandlers    []PluginMapClickHandler
	trackDetailSections []PluginTrackDetailSection
	assets              []string
}

// Sidebar adds a sidebar navigation entry to the plugin manifest.
func (p *Plugin) Sidebar(id, label, icon, route string) *Plugin {
	p.manifest.sidebar = append(p.manifest.sidebar, PluginSidebarItem{
		ID: id, Label: label, Icon: icon, Route: route,
	})
	return p
}

// MapClickHandler adds a map click handler to the plugin manifest.
func (p *Plugin) MapClickHandler(id, label string) *Plugin {
	p.manifest.mapClickHandlers = append(p.manifest.mapClickHandlers, PluginMapClickHandler{
		ID: id, Label: label,
	})
	return p
}

// TrackDetailSection adds a track detail panel section to the plugin manifest.
func (p *Plugin) TrackDetailSection(id string) *Plugin {
	p.manifest.trackDetailSections = append(p.manifest.trackDetailSections, PluginTrackDetailSection{
		ID: id,
	})
	return p
}

// FrontendAssets declares frontend JS/CSS asset filenames served under /assets/.
func (p *Plugin) FrontendAssets(files ...string) *Plugin {
	p.manifest.assets = append(p.manifest.assets, files...)
	return p
}

// BuildManifest constructs the full PluginManifest from config and builder calls.
func (p *Plugin) BuildManifest() PluginManifest {
	return PluginManifest{
		APIVersion: "cnak.us/v1alpha1",
		Kind:       "Plugin",
		Metadata: PluginMetadata{
			Name:        p.name,
			Version:     p.version,
			Author:      p.config.author,
			Description: p.config.description,
			Labels:      p.config.labels,
		},
		Spec: PluginSpec{
			MinCnakVersion: p.config.minCnakVersion,
			Permissions: PluginPermissions{
				Required: p.config.permissions,
			},
			Backend: PluginBackend{
				Port:       p.config.port,
				HealthPath: "/health",
			},
			Frontend: PluginFrontend{
				Assets:              p.manifest.assets,
				Sidebar:             p.manifest.sidebar,
				MapClickHandlers:    p.manifest.mapClickHandlers,
				TrackDetailSections: p.manifest.trackDetailSections,
			},
		},
	}
}
