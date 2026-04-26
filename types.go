// Package sdk provides a Go SDK for building CNAK plugins.
package sdk

import "time"

// Point represents a geographic entity position (mirrors CNAK's pkg/types.Point).
type Point struct {
	ID        string    `json:"id"`
	TrackID   string    `json:"trackId"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude,omitempty"`
	CE        float64   `json:"ce,omitempty"`
	LE        float64   `json:"le,omitempty"`
	Speed     float64   `json:"speed,omitempty"`
	Course    float64   `json:"course,omitempty"`
	Type      string    `json:"type"`
	Callsign  string    `json:"callsign,omitempty"`
	Group     string    `json:"group,omitempty"`
	How       string    `json:"how,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Stale     time.Time `json:"stale,omitempty"`
}

// Track represents a complete track with its history.
type Track struct {
	TrackID string  `json:"trackId"`
	Points  []Point `json:"points"`
}

// CoTEvent is a Cursor on Target event as published on NATS.
type CoTEvent struct {
	UID       string    `json:"uid"`
	Type      string    `json:"type"`
	How       string    `json:"how,omitempty"`
	Callsign  string    `json:"callsign,omitempty"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude,omitempty"`
	CE        float64   `json:"ce,omitempty"`
	LE        float64   `json:"le,omitempty"`
	Speed     float64   `json:"speed,omitempty"`
	Course    float64   `json:"course,omitempty"`
	Group     string    `json:"group,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Stale     time.Time `json:"stale,omitempty"`
}
