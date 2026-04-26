package sdk

import (
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// TrackHandler is called when a track update is received.
type TrackHandler func(subject string, point Point)

// NATSHandler is called for raw NATS messages on a subscribed subject.
type NATSHandler func(subject string, data []byte)

// OnTrackUpdate subscribes to "tracks.>" and calls fn for each update.
// Call this before Run().
func (p *Plugin) OnTrackUpdate(fn TrackHandler) *Plugin {
	p.trackHandlers = append(p.trackHandlers, fn)
	return p
}

// OnGeofenceAlert subscribes to "geofence.alert.>" and calls fn for each alert.
// Call this before Run().
func (p *Plugin) OnGeofenceAlert(fn NATSHandler) *Plugin {
	p.natsHandlers = append(p.natsHandlers, natsSubscription{
		subject: "geofence.alert.>",
		handler: fn,
	})
	return p
}

// OnNATSMessage subscribes to an arbitrary NATS subject and calls fn.
// Call this before Run().
func (p *Plugin) OnNATSMessage(subject string, fn NATSHandler) *Plugin {
	p.natsHandlers = append(p.natsHandlers, natsSubscription{
		subject: subject,
		handler: fn,
	})
	return p
}

type natsSubscription struct {
	subject string
	handler NATSHandler
}

// setupSubscriptions creates all NATS subscriptions after connection.
func (p *Plugin) setupSubscriptions() {
	// Track update handlers
	if len(p.trackHandlers) > 0 {
		_, err := p.nc.Subscribe("tracks.>", func(msg *nats.Msg) {
			var pt Point
			if err := json.Unmarshal(msg.Data, &pt); err != nil {
				return
			}
			for _, fn := range p.trackHandlers {
				fn(msg.Subject, pt)
			}
		})
		if err != nil {
			slog.Warn("failed to subscribe to tracks", "plugin", p.name, "error", err)
		}
	}

	// Generic NATS handlers (includes geofence alerts)
	for _, sub := range p.natsHandlers {
		sub := sub
		_, err := p.nc.Subscribe(sub.subject, func(msg *nats.Msg) {
			sub.handler(msg.Subject, msg.Data)
		})
		if err != nil {
			slog.Warn("failed to subscribe", "plugin", p.name, "subject", sub.subject, "error", err)
		}
	}
}
