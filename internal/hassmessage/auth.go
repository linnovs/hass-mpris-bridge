package hassmessage

// AuthRequired will sent by the server when client connected.
type AuthRequired struct {
	Type    MessageType `json:"type"`
	Version string      `json:"ha_version"`
}

// Auth should sent to server once the client received [AuthRequired] message.
type Auth struct {
	Type  MessageType `json:"type"`
	Token string      `json:"access_token"`
}

// AuthResult should be the return message after the client supply the authentication data.
type AuthResult struct {
	Type    MessageType `json:"type"`
	Version string      `json:"ha_version"`
	Message string      `json:"message"`
}
