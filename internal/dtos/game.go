package dtos

type GameMessage struct {
	MessageType string      `json:"message_type"`
	Message     interface{} `json:"message"`
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
	Message string `json:"message"`
	Sender  string `json:"sender"`
	Time    string `json:"time"`
}
