package hassmessage

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
				Duration    int     `json:"media_duration"`
				Position    int     `json:"media_position"`
				Title       string  `json:"media_title"`
				VolumeLevel float32 `json:"volume_level"`
			} `json:"attributes"`
		} `json:"new_state"`
	} `json:"data"`
}
