package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/linnovs/hass-mpris-bridge/internal/bufferpool"
	"github.com/linnovs/hass-mpris-bridge/internal/hassmessage"
)

var (
	errUnexpectedMsg = errors.New("unexpected message after command")
	errCommandFailed = errors.New("command result failed")
)

type hassClient struct {
	ctx          context.Context
	conn         *websocket.Conn
	receiversMux sync.Mutex
	receivers    map[uint64]chan hassmessage.Message
	messageID    atomic.Uint64
}

func (c *hassClient) listen(errc chan<- error) {
	for {
		var msg hassmessage.Message

		_, reader, err := c.conn.Reader(c.ctx)
		if err != nil {
			errc <- err
			return
		}

		buf := bufferpool.Get()

		if _, err := buf.ReadFrom(reader); err != nil {
			errc <- err
			return
		}

		log.Debug("read message from HASS", "message", buf)

		if err := json.Unmarshal(buf.Bytes(), &msg); err != nil {
			errc <- err
			return
		}

		if msg.Type == hassmessage.TypePong {
			log.Debug("pong message received", "id", msg.ID)
			continue
		}

		if msg.Type == hassmessage.TypeReuseID {
			errc <- errors.New("HASS websocket id reuse, should recreate connection")
			return
		}

		c.receiversMux.Lock()
		receiverCh, ok := c.receivers[msg.ID]
		c.receiversMux.Unlock()
		if !ok {
			log.Warn("message received but no subscriber", "message", msg)
			continue
		}

		select {
		case receiverCh <- msg:
		default:
			go func() { receiverCh <- msg }()
		}
	}
}

func (c *hassClient) heartbeat() {
	const interval = 45 * time.Second

	f := func() {
		id := c.incrementID()
		msg := hassmessage.Command{ID: id, Type: hassmessage.TypePing}

		if err := wsjson.Write(c.ctx, c.conn, &msg); err != nil {
			log.Error("senting ping message failed", "err", err)

			return
		}

		log.Debug("ping message sent", "id", id)
	}

	f()
	for range time.Tick(interval) {
		select {
		case <-c.ctx.Done():
			return
		default:
			f()
		}
	}
}

func (c *hassClient) connect(uri, token string, errc chan<- error) (err error) {
	conn, _, err := websocket.Dial(c.ctx, uri, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			conn.Close(websocket.StatusInternalError, err.Error())
		}
	}()

	var authRequired hassmessage.AuthRequired
	if err := wsjson.Read(c.ctx, conn, &authRequired); err != nil {
		return err
	}

	if authRequired.Type != hassmessage.TypeAuthRequired {
		err = errors.New("unexpected first message")
		return err
	}

	authMsg := hassmessage.Auth{Type: hassmessage.TypeAuth, Token: token}
	if err := wsjson.Write(c.ctx, conn, &authMsg); err != nil {
		return err
	}

	var authResult hassmessage.AuthResult
	if err := wsjson.Read(c.ctx, conn, &authResult); err != nil {
		return err
	}

	if authResult.Type != hassmessage.TypeAuthOK {
		err = fmt.Errorf("authentication failed: %s", authResult.Message)
		return err
	}

	log.Info("home assistant connected", "version", authResult.Version)
	c.conn = conn

	go c.heartbeat()
	go c.listen(errc)

	return nil
}

func (c *hassClient) incrementID() uint64 {
	return c.messageID.Add(1)
}

func (c *hassClient) commandDone(id uint64) {
	c.receiversMux.Lock()
	defer c.receiversMux.Unlock()

	delete(c.receivers, id)
}

func (c *hassClient) sendCommand(
	cmd hassmessage.Command,
) (id uint64, msg hassmessage.Message, err error) {
	ch := make(chan hassmessage.Message, 1)
	cmd.ID = c.incrementID()

	c.receiversMux.Lock()
	c.receivers[cmd.ID] = ch
	c.receiversMux.Unlock()

	if err := wsjson.Write(c.ctx, c.conn, &cmd); err != nil {
		c.commandDone(cmd.ID)
		return 0, msg, err
	}

	msg = <-ch
	if msg.Type != hassmessage.TypeResult {
		c.commandDone(cmd.ID)
		return 0, msg, errUnexpectedMsg
	}

	if !msg.Success {
		c.commandDone(cmd.ID)
		return 0, msg, errCommandFailed
	}

	return cmd.ID, msg, nil
}

func (c *hassClient) subscribe(evtType hassmessage.EventType) (<-chan hassmessage.Message, error) {
	id, msg, err := c.sendCommand(hassmessage.Command{
		Type:      hassmessage.TypeCommandSubscribeEvent,
		EventType: evtType,
	})
	if err != nil {
		if err == errCommandFailed {
			log.Error("command failed", "message", msg.Result)
		}

		return nil, err
	}

	c.receiversMux.Lock()
	defer c.receiversMux.Unlock()

	log.Info("subscribe to HASS event", "event", evtType)

	return c.receivers[id], nil
}

func (c *hassClient) close() {
	if err := c.conn.Close(websocket.StatusNormalClosure, "goodbye"); err != nil {
		log.Error("HASS websocket close failed", "err", err)
	} else {
		log.Info("closed HASS websocket connection")
	}
}

func newHASSClient(ctx context.Context) *hassClient {
	return &hassClient{
		ctx:       ctx,
		receivers: make(map[uint64]chan hassmessage.Message),
	}
}
