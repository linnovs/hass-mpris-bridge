package main

import (
	"errors"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/linnovs/hass-mpris-bridge/internal/hassmessage"
)

const (
	playerMinimumRate = float64(1)
	playerMaximumRate = float64(1)
)

type playbackStatus string

const (
	playbackPlaying playbackStatus = "Playing"
	playbackPaused  playbackStatus = "Paused"
	playbackStopped playbackStatus = "Stopped"
)

type loopStatus string

const (
	loopNone     loopStatus = "None"
	loopTrack    loopStatus = "Track"
	loopPlaylist loopStatus = "Playlist"
)

type playerMetadata map[string]dbus.Variant

type player struct {
	mux       sync.Mutex
	client    *hassClient
	entityID  string
	propsSpec map[string]*prop.Prop
}

func (p *player) callService(
	service hassmessage.ServiceType,
	data *hassmessage.CommandData,
) *dbus.Error {
	p.mux.Lock()
	defer p.mux.Unlock()

	domain := hassmessage.DomainMediaPlayer
	rtResp := false

	id, msg, err := p.client.sendCommand(hassmessage.Command{
		Type:    hassmessage.TypeCallService,
		Domain:  &domain,
		Service: &service,
		Target: &hassmessage.Target{
			EntityID: &p.entityID,
		},
		ServiceData:    data,
		ReturnResponse: &rtResp,
	})
	if err != nil {
		if err == errCommandFailed {
			return dbus.MakeFailedError(errors.New(msg.Error.Message))
		}
		return dbus.MakeFailedError(err)
	}

	p.client.commandDone(id)

	return nil
}

// Pause pauses playback.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p *player) Pause() *dbus.Error {
	return p.callService(hassmessage.ServicePause, nil)
}

// PlayPause pauses playback.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p *player) PlayPause() *dbus.Error {
	return p.callService(hassmessage.ServicePlayPause, nil)
}

// Play start or resumes playback.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p *player) Play() *dbus.Error {
	return p.callService(hassmessage.ServicePlay, nil)
}

func (p *player) setEntityID(id string) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.entityID = id
}

func (p *player) props() map[string]*prop.Prop {
	if p.propsSpec != nil {
		return p.propsSpec
	}

	p.propsSpec = map[string]*prop.Prop{
		"PlaybackStatus": {Value: playbackStopped, Writable: false, Emit: prop.EmitTrue},
		"LoopStatus":     {Value: loopNone, Writable: true, Emit: prop.EmitTrue},
		"Rate":           {Value: float64(1), Writable: true, Emit: prop.EmitTrue},
		"Shuffle":        {Value: false, Writable: true, Emit: prop.EmitTrue},
		"Metadata":       {Value: playerMetadata{}, Writable: false, Emit: prop.EmitTrue},
		"Volume":         {Value: float64(0), Writable: true, Emit: prop.EmitTrue},
		"Position":       {Value: int64(0), Writable: false, Emit: prop.EmitFalse},
		"MinimumRate":    {Value: playerMinimumRate, Writable: false, Emit: prop.EmitTrue},
		"MaximumRate":    {Value: playerMaximumRate, Writable: false, Emit: prop.EmitTrue},
		"CanGoNext":      {Value: true, Writable: false, Emit: prop.EmitTrue},
		"CanGoPrevious":  {Value: true, Writable: false, Emit: prop.EmitTrue},
		"CanPlay":        {Value: true, Writable: false, Emit: prop.EmitTrue},
		"CanPause":       {Value: true, Writable: false, Emit: prop.EmitTrue},
		"CanSeek":        {Value: true, Writable: false, Emit: prop.EmitTrue},
		"CanControl":     {Value: true, Writable: false, Emit: prop.EmitFalse},
	}

	return p.propsSpec
}
