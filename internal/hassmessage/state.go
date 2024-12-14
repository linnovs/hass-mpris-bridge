package hassmessage

import "encoding/json"

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
			panic(err)
		}
	}
}

func (s *State) contentType() string {
	s.parseAttrs()
	return s.attrs.ContentType
}

func (s *State) repeat() MediaPlayerAttrRepeat {
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

func (s *State) shuffle() bool {
	s.parseAttrs()
	return s.attrs.Shuffle
}

func (s *State) duration() int64 {
	s.parseAttrs()
	return s.attrs.Duration
}

func (s *State) artUrl() string {
	s.parseAttrs()
	return s.attrs.Picture
}

func (s *State) album() string {
	s.parseAttrs()
	return s.attrs.Album
}

func (s *State) artist() string {
	s.parseAttrs()
	return s.attrs.Artist
}

func (s *State) title() string {
	s.parseAttrs()
	return s.attrs.Title
}

func (s *State) volume() float64 {
	s.parseAttrs()
	return s.attrs.VolumeLevel
}

func (s *State) position() int64 {
	s.parseAttrs()
	return s.attrs.Position
}
