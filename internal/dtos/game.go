package dtos

import "encoding/json"

type GameMessage struct {
	MessageType string          `json:"message_type"`
	Message     json.RawMessage `json:"message"`
}

type Line struct {
	StartX float64 `json:"start_x"`
	StartY float64 `json:"start_y"`
	EndX   float64 `json:"end_x"`
	EndY   float64 `json:"end_y"`
}

type ChooseGameWordsMessage struct {
	Words []string `json:"words"`
}

type Chat struct {
	Message    string `json:"message"`
	Sender     string `json:"sender"`
	TimeMillis int64  `json:"time"`
}

type GameWords struct {
	Words []string `json:"words"`
}
