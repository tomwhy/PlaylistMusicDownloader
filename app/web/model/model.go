package model

type WebsocketMessage struct {
	MessageType string      `json:"type"`
	Data        interface{} `json:"data"`
}
