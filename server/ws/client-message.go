package ws

// ClientMessage is a wrapper for a message that includes the sender
type ClientMessage struct {
	Client  *Client
	Message *Message
}
