package dtos

import (
	"github.com/gorilla/websocket"
)

type Room struct {
	RoomId           string
	RoomName         string
	Players          []*Player
	GameStatus       string
	CurrentRoundWord *string
	DrawnUsers       map[string]bool
	DrawingUser      *Player
}

type Player struct {
	PlayerId   string
	IsAdmin    bool
	PlayerName string
	PlayerConn *websocket.Conn
}

type RoomCreationSuccess struct {
	RoomId   string `json:"room_id"`
	RoomName string `json:"room_name"`
}

type RoomCreationMessage struct {
	RoomId   string `json:"room_id"`
	RoomName string `json:"room_name"`
}

type NetworkResponse struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
