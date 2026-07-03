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
	Annotations map[string]string `json:"annotations,omitempty"`
}

// RestartPolicy controls supervisor behaviour when a subprocess plugin exits.
// Mirrors the backend's RestartPolicy enum. Sidecar plugins (Docker/k8s
// owned) treat this field as advisory only — the container runtime is
// authoritative for their lifecycle.
type RestartPolicy string

const (
	// RestartOnFailure (default when empty) restarts only on non-zero exit.
	RestartOnFailure RestartPolicy = "OnFailure"
	// RestartAlways restarts after every exit.
	RestartAlways RestartPolicy = "Always"
	// RestartNever disables the supervisor.
	RestartNever RestartPolicy = "Never"
)

// PluginSpec contains the plugin configuration.
type PluginSpec struct {
	MinCnakVersion string             `json:"minCnakVersion,omitempty"`
	Permissions    PluginPermissions  `json:"permissions"`
	Backend        PluginBackend      `json:"backend"`
	Frontend       PluginFrontend     `json:"frontend"`
	Resources      PluginResourceSpec `json:"resources,omitempty"`
	Signing        PluginSigningSpec  `json:"signing,omitempty"`
	// RestartPolicy is additive-in-v1alpha1: omitted manifests inherit the
	// backend default (OnFailure). Builder method: WithRestartPolicy.
	RestartPolicy RestartPolicy `json:"restartPolicy,omitempty"`
}

// PluginPermissions declares required and optional permissions.
type PluginPermissions struct {
	Required []string `json:"required,omitempty"`
	Optional []string `json:"optional,omitempty"`
}

// PluginBackend describes the backend sidecar.
type PluginBackend struct {
	Binary     string `json:"binary,omitempty"`
	Port       int    `json:"port,omitempty"`
	HealthPath string `json:"healthPath,omitempty"`
}

// PluginResourceSpec declares resource constraints. Mirrors backend's PluginResourceSpec.
type PluginResourceSpec struct {
	MaxMemoryMB   int  `json:"maxMemoryMB,omitempty"`
	NetworkEgress bool `json:"networkEgress"`
}

// PluginSigningSpec contains code signing configuration. Mirrors backend's PluginSigningSpec.
type PluginSigningSpec struct {
	PublicKey string `json:"publicKey,omitempty"`
}

// PluginFrontend describes frontend extension points.
type PluginFrontend struct {
	Assets              []string                   `json:"assets,omitempty"`
	Sidebar             []PluginSidebarItem        `json:"sidebar,omitempty"`
	MapClickHandlers    []PluginMapClickHandler    `json:"mapClickHandlers,omitempty"`
	TrackDetailSections []PluginTrackDetailSection `json:"trackDetailSections,omitempty"`
	DockedPanel         *PluginDockedPanel         `json:"dockedPanel,omitempty"`
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

// PluginDockedPanel describes a header-toggleable docked side panel.
// A button is added to the CNAK header; clicking it opens the plugin
// bundle in a resizable rail (similar to the AI advisor).
type PluginDockedPanel struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Icon            string `json:"icon,omitempty"`
	DefaultPosition string `json:"defaultPosition,omitempty"` // "left" | "right" (default "right")
	BadgePath       string `json:"badgePath,omitempty"`       // GET path under plugin's API returning { unreadCount: int }
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
	dockedPanel         *PluginDockedPanel
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

// DockedPanel registers the plugin as a header-toggleable docked side panel.
// Pass id, label, and icon name (resolved on the frontend through CNAK's
// PLUGIN_ICON_MAP). Default position is "right" — call WithDockedPanelPosition
// to override. Call WithDockedPanelBadge to declare a polled unread-count
// endpoint on the plugin's own API namespace.
//
//	sdk.New("signal-bridge", "0.1.0").
//	    DockedPanel("signal-bridge", "Signal", "SiSignal").
//	    WithDockedPanelBadge("/me")
func (p *Plugin) DockedPanel(id, label, icon string) *Plugin {
	if p.manifest.dockedPanel == nil {
		p.manifest.dockedPanel = &PluginDockedPanel{}
	}
	p.manifest.dockedPanel.ID = id
	p.manifest.dockedPanel.Label = label
	p.manifest.dockedPanel.Icon = icon
	return p
}

// WithDockedPanelPosition sets the default rail side ("left" or "right").
// Falls back to "right" when unset.
func (p *Plugin) WithDockedPanelPosition(position string) *Plugin {
	if p.manifest.dockedPanel == nil {
		p.manifest.dockedPanel = &PluginDockedPanel{}
	}
	p.manifest.dockedPanel.DefaultPosition = position
	return p
}

// WithDockedPanelBadge declares a GET path under the plugin's API namespace
// (e.g. "/me") that the CNAK header polls every 30s for an unread count.
// Response shape: { "unreadCount": <int> }.
func (p *Plugin) WithDockedPanelBadge(path string) *Plugin {
	if p.manifest.dockedPanel == nil {
		p.manifest.dockedPanel = &PluginDockedPanel{}
	}
	p.manifest.dockedPanel.BadgePath = path
	return p
}

// BuildManifest constructs the full PluginManifest from config and builder calls.
func (p *Plugin) BuildManifest() PluginManifest {
	spec := PluginSpec{
		MinCnakVersion: p.config.minCnakVersion,
		Permissions: PluginPermissions{
			Required: p.config.permissions,
			Optional: p.config.optionalPermissions,
		},
		Backend: PluginBackend{
			Binary:     p.config.binary,
			Port:       p.config.port,
			HealthPath: "/health",
		},
		Frontend: PluginFrontend{
			Assets:              p.manifest.assets,
			Sidebar:             p.manifest.sidebar,
			MapClickHandlers:    p.manifest.mapClickHandlers,
			TrackDetailSections: p.manifest.trackDetailSections,
			DockedPanel:         p.manifest.dockedPanel,
		},
		RestartPolicy: p.config.restartPolicy,
	}
	if p.config.resourcesSet {
		spec.Resources = PluginResourceSpec{
			MaxMemoryMB:   p.config.maxMemoryMB,
			NetworkEgress: p.config.networkEgress,
		}
	}
	return PluginManifest{
		APIVersion: "cnak.us/v1alpha1",
		Kind:       "Plugin",
		Metadata: PluginMetadata{
			Name:        p.name,
			Version:     p.version,
			Author:      p.config.author,
			Description: p.config.description,
			Labels:      p.config.labels,
			Annotations: p.config.annotations,
		},
		Spec: spec,
	}
}
