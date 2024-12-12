package hassmessage

// EventType for subscribe_events's event type.
type EventType string

const (
	// EventStateChanged represent the `state_changed` event bus.
	EventStateChanged EventType = "state_changed"
)

// ServiceType represent call_service's servic action name
type ServiceType string

const (
	ServiceVolumeSet ServiceType = "volume_set"
	ServicePlayPause ServiceType = "media_play_pause"
	ServicePlay      ServiceType = "media_play"
	ServicePause     ServiceType = "media_pause"
	ServiceStop      ServiceType = "media_stop"
	ServiceNext      ServiceType = "media_next_track"
	ServicePrevious  ServiceType = "media_previous_track"
	ServiceSeek      ServiceType = "media_seek"
	ServiceShuffle   ServiceType = "shuffle_set"
	ServiceRepeat    ServiceType = "repeat_set"
)

// ServiceDomain is the domain for a command.
type ServiceDomain string

const (
	// DomainMediaPlayer is the media_player domain
	DomainMediaPlayer ServiceDomain = "media_player"
)

// CommandData represent the `service_data` in calling a service.
type CommandData struct {
	IsMuted      *bool    `json:"is_volume_muted,omitempty"`
	VolumeLevel  *float64 `json:"volume_level,omitempty"`
	SeekPosition *int     `json:"seek_position,omitempty"`
	Shuffle      *bool    `json:"shuffle,omitempty"`
	RepeatMode   *string  `json:"repeat,omitempty"`
}

// Target represent the `target` in calling a service.
type Target struct {
	EntityID *string `json:"entity_id,omitempty"`
}

// Command represent the command message for the client.
type Command struct {
	ID   uint64      `json:"id"`
	Type MessageType `json:"type"`

	// Optional params
	Domain         *ServiceDomain `json:"domain,omitempty"`
	Service        *ServiceType   `json:"service,omitempty"`
	ServiceData    *CommandData   `json:"service_data,omitempty"`
	Target         *Target        `json:"target,omitempty"`
	ReturnResponse *bool          `json:"return_response,omitempty"`
	EventType      *EventType     `json:"event_type,omitempty"`
}
