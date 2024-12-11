package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"github.com/linnovs/hass-mpris-bridge/internal/hassmessage"
)

const (
	dbusNameFormat       = dbusObjectIface + ".hassbridge.instance%d"
	dbusObjectPath       = "/org/mpris/MediaPlayer2"
	dbusObjectIface      = "org.mpris.MediaPlayer2"
	dbusPlayerIface      = dbusObjectIface + ".Player"
	dbusPropertiesIface  = "org.freedesktop.DBus.Properties"
	dbusPropChangedIface = dbusPropertiesIface + "PropertiesChanged"
	desktopName          = "HASS media_player to MPRIS Bridge"
	desktopEntry         = "hassbridge"
)

// bridge is the D-bus object implementing `org.mpris.MediaPlayer2`.
type bridge struct {
	ctx        context.Context
	player     *player
	conn       *dbus.Conn // shared connection don't close
	errc       chan<- error
	propsSpec  map[string]*prop.Prop
	properties *prop.Properties
}

// Raise do nothing.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Raise
func (b *bridge) Raise() *dbus.Error {
	return nil
}

// Quit do nothing.
// see: https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Quit
func (b *bridge) Quit() *dbus.Error {
	return nil
}

func (b *bridge) close() {
	if err := b.conn.Close(); err != nil {
		log.Error("D-bus connection close failed", "err", err)
	}
}

func (b *bridge) props() map[string]*prop.Prop {
	if b.propsSpec != nil {
		return b.propsSpec
	}

	b.propsSpec = map[string]*prop.Prop{
		"CanQuit":             {Value: false, Writable: false, Emit: prop.EmitTrue},
		"Fullscreen":          {Value: false, Writable: false, Emit: prop.EmitTrue},
		"CanSetFullscreen":    {Value: false, Writable: false, Emit: prop.EmitTrue},
		"CanRaise":            {Value: false, Writable: false, Emit: prop.EmitTrue},
		"HasTrackList":        {Value: false, Writable: false, Emit: prop.EmitTrue},
		"Identity":            {Value: desktopName, Writable: false, Emit: prop.EmitTrue},
		"DesktopEntry":        {Value: desktopEntry, Writable: false, Emit: prop.EmitTrue},
		"SupportedUriSchemes": {Value: []string{}, Writable: false, Emit: prop.EmitTrue},
		"SupportedMimeTypes":  {Value: []string{}, Writable: false, Emit: prop.EmitTrue},
	}

	return b.propsSpec
}

func (b *bridge) export() (objIface, plyIface introspect.Interface, err error) {
	if err := b.conn.Export(b, dbusObjectPath, dbusObjectIface); err != nil {
		return objIface, plyIface, err
	}

	if err := b.conn.Export(b.player, dbusObjectPath, dbusPlayerIface); err != nil {
		return objIface, plyIface, err
	}

	props, err := prop.Export(b.conn, dbusObjectPath, map[string]map[string]*prop.Prop{
		dbusObjectIface: b.props(),
		dbusPlayerIface: b.player.props(),
	})
	if err != nil {
		return objIface, plyIface, err
	}

	objIface.Methods = introspect.Methods(b)
	objIface.Properties = props.Introspection(dbusObjectIface)
	plyIface.Methods = introspect.Methods(b.player)
	plyIface.Properties = props.Introspection(dbusPlayerIface)
	b.properties = props

	return objIface, plyIface, nil
}

func (b *bridge) connect(errc chan<- error) (err error) {
	reply, err := b.conn.RequestName(
		fmt.Sprintf(dbusNameFormat, os.Getpid()),
		dbus.NameFlagDoNotQueue,
	)
	if err != nil {
		return err
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.New("D-bus name already taken")
	}

	objIface, plyIface, err := b.export()
	if err != nil {
		return err
	}

	b.errc = errc
	n := introspect.NewIntrospectable(&introspect.Node{
		Name: dbusObjectPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			objIface,
			plyIface,
		},
	})

	if err := b.conn.Export(n, dbusObjectPath, introspect.IntrospectData.Name); err != nil {
		return err
	}

	return nil
}

func (b *bridge) setPlayerProps(name string, value dbus.Variant) {
	b.properties.SetMust(dbusPlayerIface, name, value)
}

func (b *bridge) update(msg hassmessage.Message) {
	// sees https://www.home-assistant.io/integrations/media_player#the-state-of-a-media-player
	switch msg.Event.State() {
	case "playing":
		b.setPlayerProps("PlaybackStatus", dbus.MakeVariant(playbackPlaying))
	case "paused", "buffering":
		b.setPlayerProps("PlaybackStatus", dbus.MakeVariant(playbackPaused))
	default:
		b.setPlayerProps("PlaybackStatus", dbus.MakeVariant(playbackStopped))
	}

	switch msg.Event.LoopStatus() {
	case "all":
		b.setPlayerProps("LoopStatus", dbus.MakeVariant(loopPlaylist))
	case "one":
		b.setPlayerProps("LoopStatus", dbus.MakeVariant(loopTrack))
	case "none":
		b.setPlayerProps("LoopStatus", dbus.MakeVariant(loopNone))
	}

	if msg.Event.Shuffle() != nil {
		b.setPlayerProps("Shuffle", dbus.MakeVariant(*msg.Event.Shuffle()))
	}

	b.setPlayerProps("Metadata", dbus.MakeVariant(map[string]dbus.Variant{
		"mpris:length": dbus.MakeVariant(msg.Event.Duration()),
		"mpris:artUrl": dbus.MakeVariant(msg.Event.ArtURL()),
		"xesam:album":  dbus.MakeVariant(msg.Event.Album()),
		"xesam:artist": dbus.MakeVariant(msg.Event.Artist()),
		"xesam:title":  dbus.MakeVariant(msg.Event.Title()),
	}))

	b.setPlayerProps("Volume", dbus.MakeVariant(msg.Event.Volume()))
	b.setPlayerProps("Position", dbus.MakeVariant(msg.Event.Position()))
}

func newBridge(ctx context.Context) (b *bridge, err error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	return &bridge{ctx: ctx, player: &player{}, conn: conn}, nil
}
