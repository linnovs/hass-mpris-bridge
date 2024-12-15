package hassmessage

import (
	"encoding/json"
)

type MediaPlayerData struct {
	State State `json:"new_state"`
}

// Event will sent by the server after the client sent the `subscribe_events` commands.
type Event struct {
	EventType EventType       `json:"event_type"`
	Data      json.RawMessage `json:"data"`
}
