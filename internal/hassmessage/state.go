package hassmessage

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/log"
)

type MediaPlayerAttrRepeat int

const (
	MediaPlayerAttrRepeatOff MediaPlayerAttrRepeat = 1 << iota
	MediaPlayerAttrRepeatOne
	MediaPlayerAttrRepeatAll
)

func (r MediaPlayerAttrRepeat) String() string {
	switch r {
	case MediaPlayerAttrRepeatAll:
		return "Playlist"
	case MediaPlayerAttrRepeatOne:
		return "Track"
	default:
		return "None"
	}
}

type MediaPlayerAttrState int

const (
	MediaPlayerAttrStateIdle = 1 << iota
	MediaPlayerAttrStatePlaying
	MediaPlayerAttrStatePaused
	MediaPlayerAttrStateStopped
)

func (s MediaPlayerAttrState) String() string {
	switch s {
	case MediaPlayerAttrStatePlaying:
		return "Playing"
	case MediaPlayerAttrStatePaused:
		return "Paused"
	default:
		return "Stopped"
	}
}

type MediaPlayerAttributes struct {
	ID          string  `json:"app_id"`
	Name        string  `json:"app_name"`
	Picture     string  `json:"entity_picture"`
	Album       string  `json:"media_album_name"`
	Artist      string  `json:"media_artist"`
	Duration    int64   `json:"media_duration"`
	Position    int64   `json:"media_position"`
	Title       string  `json:"media_title"`
	VolumeLevel float64 `json:"volume_level"`
	Shuffle     bool    `json:"shuffle"`
	Repeat      string  `json:"repeat"`
	ContentType string  `json:"media_content_type"`
}

type State struct {
	EntityID   string          `json:"entity_id"`
	State      string          `json:"state"`
	Attributes json.RawMessage `json:"attributes"`
	attrs      *MediaPlayerAttributes
}

func (s *State) parseAttrs() {
	if s.attrs == nil {
		s.attrs = &MediaPlayerAttributes{}
		if err := json.Unmarshal(s.Attributes, s.attrs); err != nil {
			log.Error("failed to unmarshal attributes", "attr", string(s.Attributes))
			panic(err)
		}
	}
}

const mediaPlayerPrefix = "media_player."

func (s *State) IsMediaPlayer() bool {
	return strings.HasPrefix(s.EntityID, mediaPlayerPrefix)
}

func (s *State) IsMusicPlayer() bool {
	s.parseAttrs()
	return s.attrs.ContentType == "music"
}

func (s *State) PlaybackState() MediaPlayerAttrState {
	switch s.State {
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

func (s *State) Repeat() MediaPlayerAttrRepeat {
	s.parseAttrs()
	switch s.attrs.Repeat {
	case "all":
		return MediaPlayerAttrRepeatAll
	case "one":
		return MediaPlayerAttrRepeatOne
	default:
		return MediaPlayerAttrRepeatOff
	}
}

func (s *State) Shuffle() bool {
	s.parseAttrs()
	return s.attrs.Shuffle
}

func (s *State) Duration() int64 {
	s.parseAttrs()
	return s.attrs.Duration
}

func (s *State) ArtURL() string {
	s.parseAttrs()
	return s.attrs.Picture
}

func (s *State) Album() string {
	s.parseAttrs()
	return s.attrs.Album
}

func (s *State) Artist() string {
	s.parseAttrs()
	return s.attrs.Artist
}

func (s *State) Title() string {
	s.parseAttrs()
	return s.attrs.Title
}

func (s *State) Volume() float64 {
	s.parseAttrs()
	return s.attrs.VolumeLevel
}

func (s *State) Position() int64 {
	s.parseAttrs()
	return s.attrs.Position
}
