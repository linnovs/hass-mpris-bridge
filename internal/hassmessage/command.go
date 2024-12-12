package hassmessage

// EventType for subscribe_events's event type.
type EventType string

const (
	// EventStateChanged represent the `state_changed` event bus.
	EventStateChanged EventType = "state_changed"
)

// Command represent the command message for the client.
type Command struct {
	ID   uint64      `json:"id"`
	Type MessageType `json:"type"`

	// Optional params
	EventType *EventType `json:"event_type,omitempty"`
}
