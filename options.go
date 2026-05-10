package sdk

// Option configures a Plugin.
type Option func(*pluginConfig)

type pluginConfig struct {
	port               int
	natsURL            string
	natsAuthToken      string
	natsCredsFile      string
	natsNKeySeed       string
	author             string
	description        string
	permissions        []string
	optionalPermissions []string
	minCnakVersion     string
	assetsDir          string
	labels             map[string]string
	annotations        map[string]string
	backendURL         string
	serviceToken       string
	binary             string
	maxMemoryMB        int
	networkEgress      bool
	resourcesSet       bool
}

func defaultConfig() pluginConfig {
	return pluginConfig{
		port:       8200,
		natsURL:    "nats://nats-server:4222",
		backendURL: "http://backend:8080",
		assetsDir:  "./frontend",
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

// WithAnnotations sets metadata annotations on the manifest.
func WithAnnotations(annotations map[string]string) Option {
	return func(c *pluginConfig) { c.annotations = annotations }
}

// WithOptionalPermissions sets optional permissions the plugin would like
// but does not require to function.
func WithOptionalPermissions(perms ...string) Option {
	return func(c *pluginConfig) { c.optionalPermissions = perms }
}

// WithBinary sets the backend sidecar binary name in the manifest.
func WithBinary(name string) Option {
	return func(c *pluginConfig) { c.binary = name }
}

// WithResources declares resource constraints (max memory in MB and whether
// network egress is permitted) for the plugin.
func WithResources(maxMB int, egress bool) Option {
	return func(c *pluginConfig) {
		c.maxMemoryMB = maxMB
		c.networkEgress = egress
		c.resourcesSet = true
	}
}

// WithNATSToken sets the NATS auth token (default from NATS_AUTH_TOKEN env).
func WithNATSToken(token string) Option {
	return func(c *pluginConfig) { c.natsAuthToken = token }
}

// WithNATSCredsFile sets the NATS credentials file path (default from NATS_CREDENTIALS_FILE env).
func WithNATSCredsFile(path string) Option {
	return func(c *pluginConfig) { c.natsCredsFile = path }
}

// WithNATSNKeySeed sets the NATS NKey seed (default from NATS_NKEY_SEED env).
func WithNATSNKeySeed(seed string) Option {
	return func(c *pluginConfig) { c.natsNKeySeed = seed }
}

// WithBackendURL sets the backend URL for credential bootstrapping (default from BACKEND_URL env).
func WithBackendURL(url string) Option {
	return func(c *pluginConfig) { c.backendURL = url }
}
