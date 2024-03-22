package dtos

import (
	"github.com/gorilla/websocket"
)

type Room struct {
	RoomId           string
	Players          []*Player
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
	RoomId string `json:"room_id"`
}
