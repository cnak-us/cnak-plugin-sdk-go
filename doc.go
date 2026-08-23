// Copyright 2026 CNAK Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package sdk builds CNAK plugins in Go.
//
// A CNAK plugin is a small service that registers itself with the CNAK
// backend, subscribes to live data over NATS, and optionally serves an HTTP
// API and frontend assets that CNAK surfaces in its UI (sidebar entries, map
// click handlers, track detail sections, docked panels).
//
// The SDK handles the plugin lifecycle so a plugin's main function stays
// declarative: create a [Plugin] with [New], attach handlers and manifest
// entries with the builder methods, then call [Plugin.Run]:
//
//	func main() {
//	    p := sdk.New("hello-world", "0.1.0",
//	        sdk.WithAuthor("CNAK Examples"),
//	        sdk.WithDescription("Logs every track update"),
//	        sdk.WithPermissions("tracks:read"),
//	    )
//
//	    p.OnTrackUpdate(func(subject string, pt sdk.Point) {
//	        log.Printf("track %s at %.5f,%.5f", pt.TrackID, pt.Latitude, pt.Longitude)
//	    })
//
//	    if err := p.Run(); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// [Plugin.Run] connects to NATS, registers the plugin's manifest with the
// backend (retrying while the backend starts up), publishes a heartbeat
// every 30 seconds, answers discovery requests, serves HTTP on the
// configured port (default 8200) with a built-in /health endpoint, and
// blocks until SIGTERM or SIGINT triggers a graceful shutdown.
//
// # Configuration
//
// Options set at construction time can also come from the environment, which
// is how CNAK's container and Kubernetes deployments configure plugins:
//
//	NATS_URL               NATS server URL (default nats://nats-server:4222)
//	NATS_AUTH_TOKEN        NATS token auth
//	NATS_CREDENTIALS_FILE  NATS JWT credentials file
//	NATS_NKEY_SEED         NATS NKey seed
//	BACKEND_URL            CNAK backend URL (default http://backend:8080)
//	SERVICE_TOKEN          service token for HTTP registration and
//	                       credential bootstrap
//	PORT                   HTTP listen port (default 8200)
//	PLUGIN_URL             externally reachable URL advertised to the backend
//
// Explicit [Option] values passed to [New] take precedence over the
// environment.
//
// # Frontend extension points
//
// Builder methods on [Plugin] declare how the plugin appears in the CNAK UI:
// [Plugin.Sidebar], [Plugin.MapClickHandler], [Plugin.TrackDetailSection],
// [Plugin.DockedPanel], and [Plugin.FrontendAssets]. Static assets are served
// from the configured assets directory (default ./frontend) under /assets/.
//
// # Advanced use
//
// [Plugin.OnNATSMessage] subscribes to arbitrary NATS subjects, and
// [Plugin.NATS] exposes the underlying connection after [Plugin.Run] for
// publishing or request/reply patterns the SDK does not wrap.
package sdk
