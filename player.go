package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
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
	propsSpec map[string]*prop.Prop
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
