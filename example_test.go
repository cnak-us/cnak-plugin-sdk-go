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

package sdk_test

import (
	"encoding/json"
	"log"
	"net/http"

	sdk "github.com/cnak-us/cnak-plugin-sdk-go"
)

// A minimal plugin that logs every track update.
func Example() {
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

// A plugin with a sidebar entry, its own HTTP API, and frontend assets.
func Example_frontend() {
	p := sdk.New("fleet-status", "1.2.0",
		sdk.WithAuthor("Acme"),
		sdk.WithPermissions("tracks:read", "sidebar:register"),
		sdk.WithAssetsDir("./frontend"),
	)

	p.Sidebar("fleet-status", "Fleet Status", "MdDirectionsBoat", "/plugins/fleet-status").
		FrontendAssets("fleet-status.js", "fleet-status.css").
		HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]int{"vessels": 42})
		})

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// Subscribing to geofence alerts and arbitrary NATS subjects.
func ExamplePlugin_OnNATSMessage() {
	p := sdk.New("alert-relay", "0.1.0",
		sdk.WithPermissions("geofence:read"),
	)

	p.OnGeofenceAlert(func(subject string, data []byte) {
		log.Printf("geofence alert on %s: %s", subject, data)
	})

	p.OnNATSMessage("cnak.custom.>", func(subject string, data []byte) {
		log.Printf("custom message on %s", subject)
	})

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
