package sdk

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// registrationDomain and the canonical layout below are a WIRE CONTRACT with
// the CNAK backend (pkg/signing/registration.go, RegistrationClaim.Canonical).
// The two are deliberately independent implementations of the same few lines
// rather than a shared import — pulling github.com/cnak-us/cnak/pkg into every
// plugin to share 20 lines of string joining is a poor trade. Any change to
// the format is a breaking change and needs a new domain string on both sides.
const registrationDomain = "cnak-plugin-registration-v1"

// registrationClaim is what the publisher attests: this plugin, at this
// version, may hold these permissions. Not the whole manifest — routing and
// presentation detail is deployment-specific and would make a signature
// unusable across environments, while permissions are what actually grant
// privilege.
type registrationClaim struct {
	pluginID    string
	version     string
	permissions []string
}

// canonical renders the exact bytes that get signed:
//
//	cnak-plugin-registration-v1\n<pluginID>\n<version>\n<perm,perm,perm>
//
// Permissions are sorted and de-duplicated so declaration order cannot change
// the signature.
func (c registrationClaim) canonical() []byte {
	perms := make([]string, 0, len(c.permissions))
	seen := make(map[string]struct{}, len(c.permissions))
	for _, p := range c.permissions {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		perms = append(perms, p)
	}
	sort.Strings(perms)

	return []byte(strings.Join([]string{
		registrationDomain,
		c.pluginID,
		c.version,
		strings.Join(perms, ","),
	}, "\n"))
}

// signRegistration returns the base64 signature this plugin sends with its
// registration, or "" when no signing key is configured. An unsigned plugin
// is accepted by a backend without PLUGIN_TRUSTED_KEYS and rejected by one
// with it — the plugin does not need to know which.
func (p *Plugin) signRegistration() string {
	if len(p.config.signingKey) == 0 {
		return ""
	}
	claim := registrationClaim{
		pluginID:    p.name,
		version:     p.version,
		permissions: append(append([]string{}, p.config.permissions...), p.config.optionalPermissions...),
	}
	return base64.StdEncoding.EncodeToString(ed25519.Sign(p.config.signingKey, claim.canonical()))
}

// buildRegistration assembles the payload every registration path sends —
// NATS register, heartbeat, discovery re-announce, and the HTTP endpoint.
// Centralised so a new path cannot silently ship unsigned.
func (p *Plugin) buildRegistration() PluginRegistration {
	return PluginRegistration{
		Manifest:  p.BuildManifest(),
		URL:       p.pluginURL(),
		Signature: p.signRegistration(),
	}
}

// parseSigningKey decodes a base64 Ed25519 private key (64 bytes) or seed
// (32 bytes). Seeds are accepted because that is what most key-generation
// tooling emits, and expanding one is unambiguous.
func parseSigningKey(encoded string) (ed25519.PrivateKey, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		raw, err = base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("invalid signing key encoding: %w", err)
		}
	}
	switch len(raw) {
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(raw), nil
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(raw), nil
	default:
		return nil, fmt.Errorf("invalid signing key size: got %d, want %d (key) or %d (seed)",
			len(raw), ed25519.PrivateKeySize, ed25519.SeedSize)
	}
}
