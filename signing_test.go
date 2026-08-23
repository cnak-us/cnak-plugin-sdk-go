package sdk

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
)

// The backend verifies with an independent implementation of this format
// (cnak/pkg/signing/registration.go). This expectation is duplicated verbatim
// in that package's TestCanonicalWireFormat — if the two ever disagree, every
// signed plugin is rejected at registration, so both sides pin the exact
// bytes rather than trusting a shared helper neither of them imports.
func TestCanonicalWireFormatMatchesBackend(t *testing.T) {
	c := registrationClaim{
		pluginID:    "mumble-bridge",
		version:     "0.1.0",
		permissions: []string{"sidebar:register", "nats:publish:voice.>"},
	}
	want := strings.Join([]string{
		"cnak-plugin-registration-v1",
		"mumble-bridge",
		"0.1.0",
		"nats:publish:voice.>,sidebar:register",
	}, "\n")

	if got := string(c.canonical()); got != want {
		t.Fatalf("canonical =\n%q\nwant\n%q", got, want)
	}
}

func TestCanonicalIsOrderAndDuplicateInsensitive(t *testing.T) {
	a := registrationClaim{pluginID: "p", version: "1", permissions: []string{"b", "a"}}
	b := registrationClaim{pluginID: "p", version: "1", permissions: []string{"a", "b", "a", ""}}

	if string(a.canonical()) != string(b.canonical()) {
		t.Fatalf("canonical forms differ:\n %q\n %q", a.canonical(), b.canonical())
	}
}

func TestSignRegistrationProducesVerifiableSignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generating keypair: %v", err)
	}

	p := New("mumble-bridge", "0.1.0", WithPermissions("sidebar:register", "nats:publish:voice.>"))
	p.config.signingKey = priv

	sig := p.signRegistration()
	if sig == "" {
		t.Fatal("signature is empty with a key configured")
	}
	raw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("signature is not base64: %v", err)
	}

	claim := registrationClaim{
		pluginID:    "mumble-bridge",
		version:     "0.1.0",
		permissions: []string{"sidebar:register", "nats:publish:voice.>"},
	}
	if !ed25519.Verify(pub, claim.canonical(), raw) {
		t.Fatal("signature does not verify against the signing key")
	}
}

// No key configured must mean no signature, not a broken one — a backend
// without PLUGIN_TRUSTED_KEYS has to keep accepting these.
func TestSignRegistrationEmptyWithoutKey(t *testing.T) {
	p := New("mumble-bridge", "0.1.0", WithPermissions("sidebar:register"))
	if sig := p.signRegistration(); sig != "" {
		t.Fatalf("signature = %q, want empty with no signing key", sig)
	}
}

func TestBuildRegistrationCarriesSignature(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(nil)

	p := New("mumble-bridge", "0.1.0", WithPermissions("sidebar:register"))
	p.config.signingKey = priv

	if reg := p.buildRegistration(); reg.Signature == "" {
		t.Fatal("buildRegistration dropped the signature")
	}
}

func TestParseSigningKeyAcceptsKeyAndSeed(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(nil)

	full, err := parseSigningKey(base64.StdEncoding.EncodeToString(priv))
	if err != nil {
		t.Fatalf("full key: %v", err)
	}
	seed, err := parseSigningKey(base64.StdEncoding.EncodeToString(priv.Seed()))
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if string(full) != string(seed) {
		t.Error("seed and full key expanded to different private keys")
	}

	if k, err := parseSigningKey(""); err != nil || k != nil {
		t.Errorf("empty input: got (%v, %v), want (nil, nil)", k, err)
	}
	if _, err := parseSigningKey("short"); err == nil {
		t.Error("expected an error for a malformed key")
	}
}
