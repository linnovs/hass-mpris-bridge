package hassmessage

import (
	"encoding/json"
	"strings"
)

const mediaPlayerPrefix = "media_player."

type MediaPlayerData struct {
	EntityID string `json:"entity_id"`
	State    State  `json:"new_state"`
}

// Event will sent by the server after the client sent the `subscribe_events` commands.
type Event struct {
	EventType EventType       `json:"event_type"`
	Data      json.RawMessage `json:"data"`
	data      *MediaPlayerData
}

func (e *Event) parseMediaPlayerData() {
	if e.data == nil {
		e.data = &MediaPlayerData{}
		if err := json.Unmarshal(e.Data, e.data); err != nil {
			panic(err)
		}
	}
}

func (e *Event) EntityID() string {
	e.parseMediaPlayerData()
	return e.data.EntityID
}

func (e *Event) IsMediaPlayer() bool {
	e.parseMediaPlayerData()
	return strings.HasPrefix(e.data.EntityID, mediaPlayerPrefix)
}

func (e *Event) IsMusicPlayer() bool {
	e.parseMediaPlayerData()
	return e.data.State.contentType() == "music"
}

func (e *Event) State() MediaPlayerAttrState {
	e.parseMediaPlayerData()
	switch e.data.State.State {
	case "playing":
		return MediaPlayerAttrStatePlaying
	case "paused":
		return MediaPlayerAttrStatePaused
	case "standby":
		return MediaPlayerAttrStateStopped
	default:
		return MediaPlayerAttrStateIdle
	}
}

func (e *Event) Repeat() MediaPlayerAttrRepeat {
	e.parseMediaPlayerData()
	return e.data.State.repeat()
}

func (e *Event) Shuffle() bool {
	e.parseMediaPlayerData()
	return e.data.State.shuffle()
}

func (e *Event) Duration() int64 {
	e.parseMediaPlayerData()
	return e.data.State.duration()
}

func (e *Event) ArtURL() string {
	e.parseMediaPlayerData()
	return e.data.State.artUrl()
}

func (e *Event) Album() string {
	e.parseMediaPlayerData()
	return e.data.State.album()
}

func (e *Event) Artist() string {
	e.parseMediaPlayerData()
	return e.data.State.artist()
}

func (e *Event) Title() string {
	e.parseMediaPlayerData()
	return e.data.State.title()
}

func (e *Event) Volume() float64 {
	e.parseMediaPlayerData()
	return e.data.State.volume()
}

func (e *Event) Position() int64 {
	e.parseMediaPlayerData()
	return e.data.State.position()
}
