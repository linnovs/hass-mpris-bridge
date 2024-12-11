package hassmessage

import "encoding/json"

// MessageType represent the type of supported message
type MessageType string

const (
	// TypeAuthRequired is the first message received by the client when connected.
	TypeAuthRequired MessageType = "auth_required"
	// TypeAuth is the authentication message client should supply on authentication phase.
	TypeAuth MessageType = "auth"
	// TypeAuthOK is the return message when the client supply a valid authentication data.
	TypeAuthOK MessageType = "auth_ok"
	// TypeAuthInvalid is the return message when the client supply wrong authentication data.
	TypeAuthInvalid MessageType = "auth_invalid"
	// TypeResult should respond after command sent.
	TypeResult MessageType = "result"
	// TypePing should sent by the client as a heartbeat.
	TypePing MessageType = "ping"
	// TypePong will return by the server as quickly as possible when it received ping message.
	TypePong MessageType = "pong"
	// TypeCommandSubscribeEvent is the command for client subscribe to event bus on the server.
	TypeCommandSubscribeEvent MessageType = "subscribe_events"
)

// Error represent the Result message type's error field.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Message represent the message that server sent to client after authentication phase.
type Message struct {
	ID   int64       `json:"id"`
	Type MessageType `json:"type"`

	// Result message type only
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"` // this depends on the result type
	Error   Error           `json:"error"`

	// Event message type only
	Event Event `json:"event"`
}
