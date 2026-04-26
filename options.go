package sdk

// Option configures a Plugin.
type Option func(*pluginConfig)

type pluginConfig struct {
	port           int
	natsURL        string
	author         string
	description    string
	permissions    []string
	minCnakVersion string
	assetsDir      string
	labels         map[string]string
}

func defaultConfig() pluginConfig {
	return pluginConfig{
		port:      8200,
		natsURL:   "nats://nats-server:4222",
		assetsDir: "./frontend",
	}
}

// WithPort sets the HTTP port the plugin listens on (default 8200).
func WithPort(port int) Option {
	return func(c *pluginConfig) { c.port = port }
}

// WithNATSURL overrides the NATS server URL (default from NATS_URL env or nats://nats-server:4222).
func WithNATSURL(url string) Option {
	return func(c *pluginConfig) { c.natsURL = url }
}

// WithAuthor sets the plugin author in the manifest metadata.
func WithAuthor(author string) Option {
	return func(c *pluginConfig) { c.author = author }
}

// WithDescription sets the plugin description in the manifest metadata.
func WithDescription(desc string) Option {
	return func(c *pluginConfig) { c.description = desc }
}

// WithPermissions sets the required permissions (e.g., "tracks:read", "sidebar:register").
func WithPermissions(perms ...string) Option {
	return func(c *pluginConfig) { c.permissions = perms }
}

// WithMinCnakVersion sets the minimum CNAK version this plugin requires.
func WithMinCnakVersion(version string) Option {
	return func(c *pluginConfig) { c.minCnakVersion = version }
}

// WithAssetsDir sets the directory for static frontend assets (default "./frontend").
func WithAssetsDir(dir string) Option {
	return func(c *pluginConfig) { c.assetsDir = dir }
}

// WithLabels sets metadata labels on the manifest.
func WithLabels(labels map[string]string) Option {
	return func(c *pluginConfig) { c.labels = labels }
}
