package domain

type Event struct {
	GroupID string
	Name    string
	Payload map[string]any
	Room    string
}
