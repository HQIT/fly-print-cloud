package websocket

import "errors"

var (
	ErrNodeNotConnected  = errors.New("edge node not connected")
	ErrConnectionClosed  = errors.New("connection closed")
	ErrInvalidMessage    = errors.New("invalid message format")
	ErrAuthenticationFailed = errors.New("authentication failed")
)
