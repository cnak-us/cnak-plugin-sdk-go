# cnak-plugin-sdk-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cnak-us/cnak-plugin-sdk-go.svg)](https://pkg.go.dev/github.com/cnak-us/cnak-plugin-sdk-go)

The official Go SDK for building [CNAK](https://cnak.us) plugins.

A CNAK plugin is a small service that registers itself with the CNAK backend,
subscribes to live track and event data over NATS, and can extend the CNAK UI
with sidebar entries, map click handlers, track detail sections, and docked
panels. The SDK handles the full lifecycle — NATS connection and auth,
registration, heartbeat, discovery, HTTP serving, health checks, and graceful
shutdown — so a plugin stays a few dozen lines of declarative code.

## Install

```sh
go get github.com/cnak-us/cnak-plugin-sdk-go
```

## Quick start

```go
package main

import (
    "log"

    sdk "github.com/cnak-us/cnak-plugin-sdk-go"
)

func main() {
    p := sdk.New("hello-world", "0.1.0",
        sdk.WithAuthor("CNAK Examples"),
        sdk.WithDescription("Logs every track update"),
        sdk.WithPermissions("tracks:read"),
    )

    p.OnTrackUpdate(func(subject string, pt sdk.Point) {
        log.Printf("track %s at %.5f,%.5f", pt.TrackID, pt.Latitude, pt.Longitude)
    })

    if err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

`Run()` connects to NATS, registers the plugin's manifest with the backend
(retrying while the backend starts up), publishes a heartbeat every 30
seconds, serves HTTP on port 8200 with a built-in `/health` endpoint, and
blocks until SIGTERM/SIGINT.

## Configuration

Everything can be set via options (`sdk.WithPort`, `sdk.WithNATSURL`, …) or
via the environment, which is how CNAK's container and Kubernetes deployments
configure plugins:

| Variable | Purpose | Default |
| --- | --- | --- |
| `NATS_URL` | NATS server URL | `nats://nats-server:4222` |
| `NATS_AUTH_TOKEN` | NATS token auth | — |
| `NATS_CREDENTIALS_FILE` | NATS JWT credentials file | — |
| `NATS_NKEY_SEED` | NATS NKey seed | — |
| `BACKEND_URL` | CNAK backend URL | `http://backend:8080` |
| `SERVICE_TOKEN` | Service token for HTTP registration + credential bootstrap | — |
| `PORT` | HTTP listen port | `8200` |
| `PLUGIN_URL` | Externally reachable URL advertised to the backend | `http://<hostname>:<port>` |

Explicit options passed to `sdk.New` take precedence over the environment.

## Extending the CNAK UI

Builder methods declare how the plugin appears in the frontend:

```go
p.Sidebar("fleet-status", "Fleet Status", "MdDirectionsBoat", "/plugins/fleet-status").
    FrontendAssets("fleet-status.js", "fleet-status.css").
    DockedPanel("fleet-chat", "Chat", "MdChat").
    WithDockedPanelBadge("/unread").
    HandleFunc("/api/summary", summaryHandler)
```

Static assets are served from the configured assets directory (default
`./frontend`) under `/assets/`.

## Documentation

Full API documentation is on
[pkg.go.dev](https://pkg.go.dev/github.com/cnak-us/cnak-plugin-sdk-go).

## License

Apache-2.0 — see [LICENSE](LICENSE).
