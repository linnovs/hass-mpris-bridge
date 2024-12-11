package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/linnovs/hass-mpris-bridge/internal/hassmessage"
)

type hassClient struct {
	ctx           context.Context
	conn          *websocket.Conn
	subscriberMux sync.Mutex
	subscriber    map[int64]chan hassmessage.Message
	messageID     atomic.Int64
}

func (c *hassClient) listen(errc chan<- error) {
	for {
		var msg hassmessage.Message

		if err := wsjson.Read(c.ctx, c.conn, &msg); err != nil {
			errc <- err
			return
		}

		log.Debug("read message from HASS", "message", msg)

		if msg.Type == hassmessage.TypePong {
			log.Debug("pong message received", "id", msg.ID)
			continue
		}

		c.subscriberMux.Lock()
		if ch, ok := c.subscriber[msg.ID]; ok {
			select {
			case ch <- msg:
			default:
				go func() { ch <- msg }()
			}
		} else {
			log.Warn("message received but no subscriber", "msg", msg)
		}
		c.subscriberMux.Unlock()
	}
}

func (c *hassClient) heartbeat() {
	const interval = 45 * time.Second

	ticker := time.NewTicker(interval)
	for range ticker.C {
		select {
		case <-c.ctx.Done():
			return
		default:
			id := c.messageID.Add(1)
			msg := hassmessage.Command{ID: id, Type: hassmessage.TypePing}

			if err := wsjson.Write(c.ctx, c.conn, &msg); err != nil {
				log.Error("senting ping message failed", "err", err)
				continue
			}

			log.Debug("ping message sent", "id", id)
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

func (c *hassClient) subscribe(evtType hassmessage.EventType) (<-chan hassmessage.Message, error) {
	ch := make(chan hassmessage.Message, 1)
	id := c.messageID.Add(1)

	c.subscriberMux.Lock()
	c.subscriber[id] = ch
	c.subscriberMux.Unlock()

	if err := wsjson.Write(c.ctx, c.conn, &hassmessage.Command{
		ID:        id,
		Type:      hassmessage.TypeCommandSubscribeEvent,
		EventType: &evtType,
	}); err != nil {
		return nil, err
	}

	msg := <-ch
	if msg.Type != hassmessage.TypeResult {
		log.Debug("wrong message received after subscribe_events command", "msg", msg)

		c.subscriberMux.Lock()
		delete(c.subscriber, id)
		c.subscriberMux.Unlock()

		return nil, fmt.Errorf("unexpected message type after subscribe_events: %s", msg.Type)
	}

	return ch, nil
}

func (c *hassClient) close() {
	if err := c.conn.Close(websocket.StatusNormalClosure, "goodbye"); err != nil {
		log.Error("HASS websocket close failed", "err", err)
	}
}

func newHASSClient(ctx context.Context) *hassClient {
	return &hassClient{
		ctx:        ctx,
		subscriber: make(map[int64]chan hassmessage.Message),
	}
}
