package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/linnovs/hass-mpris-bridge/internal/hassmessage"
)

func main() {
	verbose := flag.Bool("v", false, "verbose log")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	errc := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	haClient := newHASSClient(ctx)
	if err := haClient.connect(os.Getenv("HASS_URI"), os.Getenv("HASS_TOKEN"), errc); err != nil {
		log.Error("connect to HASS websocket failed", "err", err)
		return
	}
	defer haClient.close()

	msgCh, err := haClient.subscribe(hassmessage.EventStateChanged)
	if err != nil {
		log.Error("subscribe to HASS state_changed event failed", "err", err)
		return
	}

	hassBridge, err := newBridge(ctx)
	if err != nil {
		log.Error("create new MPRIS bridge failed", "err", err)
		return
	}
	defer hassBridge.close()

	if err := hassBridge.connect(errc); err != nil {
		log.Error("connect to D-bus failed", "err", err)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigs:
			log.Info("graefully shutting down now.")
			return
		case err := <-errc:
			log.Error("unexpected error occur", "err", err)
			return
		case msg := <-msgCh:
			if !strings.HasPrefix(msg.Event.Data.EntityID, "media_player.") {
				continue
			}

			log.Debug(
				"HASS media_player message received",
				"id",
				msg.ID,
				"entity_id",
				msg.Event.Data.EntityID,
				"state",
				msg.Event.Data.State.State,
				"attributes",
				msg.Event.Data.State.Attributes,
			)

			hassBridge.update(msg)
		}
	}
}
