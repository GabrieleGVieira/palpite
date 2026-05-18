package dto

type RealtimeEvent struct {
	Name    string         `json:"name"`
	Payload map[string]any `json:"payload"`
	Room    string         `json:"room,omitempty"`
}
