package hassmessage

import (
	"strings"
)

// Event will sent by the server after the client sent the `subscribe_events` commands.
type Event struct {
	EventType EventType `json:"event_type"`
	Data      struct {
		EntityID string `json:"entity_id"`
		State    struct {
			State      string `json:"state"`
			Attributes struct {
				ID          string  `json:"app_id"`
				Name        string  `json:"app_name"`
				Picture     string  `json:"entity_picture"`
				Album       string  `json:"media_album_name"`
				Artist      string  `json:"media_artist"`
				Duration    int64   `json:"media_duration"`
				Position    int64   `json:"media_position"`
				Title       string  `json:"media_title"`
				VolumeLevel float64 `json:"volume_level"`
				Shuffle     *bool   `json:"shuffle,omitempty"`
				Repeat      *string `json:"repeat,omitempty"`
			} `json:"attributes"`
		} `json:"new_state"`
	} `json:"data"`
}

func (e Event) State() string {
	return strings.ToLower(e.Data.State.State)
}

func (e Event) LoopStatus() string {
	if e.Data.State.Attributes.Repeat == nil {
		return ""
	}

	return strings.ToLower(*e.Data.State.Attributes.Repeat)
}

func (e Event) Shuffle() *bool {
	return e.Data.State.Attributes.Shuffle
}

func (e Event) Duration() int64 {
	return e.Data.State.Attributes.Duration
}

func (e Event) ArtURL() string {
	return e.Data.State.Attributes.Picture
}

func (e Event) Album() string {
	return e.Data.State.Attributes.Album
}

func (e Event) Artist() string {
	return e.Data.State.Attributes.Artist
}

func (e Event) Title() string {
	return e.Data.State.Attributes.Title
}

func (e Event) Volume() float64 {
	return e.Data.State.Attributes.VolumeLevel
}

func (e Event) Position() int64 {
	return e.Data.State.Attributes.Position
}
