package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
	hassURL    string
	dir        string
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

	if err := os.RemoveAll(b.dir); err != nil {
		log.Error("remove tempdir failed", "err", err)
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
	name := fmt.Sprintf(dbusNameFormat, os.Getpid())

	reply, err := b.conn.RequestName(name, dbus.NameFlagDoNotQueue)
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

	log.Info("exported D-bus /org/mpris/MediaPlayer2", "name", name)

	return nil
}

func (b *bridge) downloadArtwork(artUrl string) string {
	if artUrl == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(artUrl))
	sumStr := base64.URLEncoding.EncodeToString(sum[:])
	file := filepath.Join(b.dir, sumStr)

	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		return fmt.Sprintf("file://%s", file)
	}

	out, err := os.Create(file)
	if err != nil {
		log.Error("failed to create temp file for download artwork image", "err", err)

		return ""
	}
	defer out.Close()

	resp, err := http.Get(fmt.Sprintf("%s%s", b.hassURL, artUrl))
	if err != nil {
		log.Error("download art work failed", "err", err)

		return ""
	}
	defer resp.Body.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Error("copy art work to file failed", "err", err)

		return ""
	}

	return fmt.Sprintf("file://%s", file)
}

func (b *bridge) update(msg hassmessage.Message) {
	if !strings.HasPrefix(msg.Event.Data.EntityID, "media_player.") && !msg.Event.IsMusicPlayer() {
		return
	}

	parseState := func(state string) dbus.Variant {
		// sees https://www.home-assistant.io/integrations/media_player#the-state-of-a-media-player
		switch state {
		case "playing":
			return dbus.MakeVariant(playbackPlaying)
		case "paused", "buffering":
			return dbus.MakeVariant(playbackPaused)
		default:
			return dbus.MakeVariant(playbackStopped)
		}
	}

	parseLoopStatus := func(loopSts string) dbus.Variant {
		switch loopSts {
		case "all":
			return dbus.MakeVariant(loopPlaylist)
		case "one":
			return dbus.MakeVariant(loopTrack)
		case "none":
			fallthrough
		default:
			return dbus.MakeVariant(loopNone)
		}
	}

	parseShuffle := func(shuffleState *bool) dbus.Variant {
		var status bool
		if shuffleState != nil {
			status = *shuffleState
		}
		return dbus.MakeVariant(status)
	}

	props := map[string]dbus.Variant{
		"PlaybackStatus": parseState(msg.Event.State()),
		"LoopStatus":     parseLoopStatus(msg.Event.LoopStatus()),
		"Shuffle":        parseShuffle(msg.Event.Shuffle()),
		"Metadata": dbus.MakeVariant(map[string]dbus.Variant{
			"mpris:length": dbus.MakeVariant(msg.Event.Duration()),
			"mpris:artUrl": dbus.MakeVariant(b.downloadArtwork(msg.Event.ArtURL())),
			"xesam:album":  dbus.MakeVariant(msg.Event.Album()),
			"xesam:artist": dbus.MakeVariant(msg.Event.Artist()),
			"xesam:title":  dbus.MakeVariant(msg.Event.Title()),
		}),
		"Volume":   dbus.MakeVariant(msg.Event.Volume()),
		"Position": dbus.MakeVariant(msg.Event.Position()),
	}

	log.Debug("update MPRIS properties", "properties", props)

	for k, v := range props {
		b.properties.SetMust(dbusPlayerIface, k, v)
	}

	b.player.setEntityID(msg.Event.Data.EntityID)
}

func newBridge(ctx context.Context, client *hassClient) (b *bridge, err error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	hassurl, err := url.Parse(os.Getenv("HASS_URI"))
	if err != nil {
		return nil, err
	}

	hassURL := fmt.Sprintf("https://%s", hassurl.Host)

	dir, err := os.MkdirTemp("", "hassbridge")
	if err != nil {
		return nil, err
	}

	return &bridge{
		ctx:     ctx,
		player:  &player{client: client},
		hassURL: hassURL,
		conn:    conn,
		dir:     dir,
	}, nil
}
