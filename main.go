package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/linnovs/hass-mpris-bridge/internal/hassmessage"
)

const (
	envkeyURI   = "HASS_URI"
	envkeyToken = "HASS_TOKEN"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	errc := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := newHASSClient(ctx)
	if err := client.connect(os.Getenv(envkeyURI), os.Getenv(envkeyToken), errc); err != nil {
		log.Error("connect to HASS websocket failed", "err", err)
		return
	}
	defer client.close()

	bdg, err := newBridge(ctx, client)
	if err != nil {
		log.Error("create new MPRIS bridge failed", "err", err)
		return
	}
	defer bdg.close()

	if err := bdg.connect(errc); err != nil {
		log.Error("connect to D-bus failed", "err", err)
		return
	}

	ch, err := client.subscribe(hassmessage.EventStateChanged)
	if err != nil {
		log.Error("subscribe to HASS state_changed event failed", "err", err)
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
		case msg := <-ch:
			bdg.update(msg)
		}
	}
}
