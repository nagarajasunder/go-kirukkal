package services

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/nagarajasunder/go-kirukkal/internal/dtos"
	"github.com/nagarajasunder/go-kirukkal/internal/globals"
)

var roomUpdateListeners = []*websocket.Conn{}

type RoomService struct {
	RoomUpdateListeners []*websocket.Conn
}

func SubscribeForRoomUpdates(conn *websocket.Conn) {
	roomUpdateListeners = append(roomUpdateListeners, conn)
	//Send the available rooms list to new subscribers

	for _, room := range roomMap {
		roomMessage := dtos.RoomCreationMessage{
			RoomId:   room.RoomId,
			RoomName: room.RoomId,
		}
		SendMessageToRoomUpdateSubscribers(conn, roomMessage)
	}

}

func ReadSubscribeMessages(conn *websocket.Conn) {

	_, _, err := conn.ReadMessage()

	if err != nil {
		conn.Close()
	}
}

func SendRoomCreationUpdates(newRoom dtos.RoomCreationMessage) {

	for _, subsConn := range roomUpdateListeners {
		SendMessageToRoomUpdateSubscribers(subsConn, newRoom)
	}
}

func SendMessageToRoomUpdateSubscribers(conn *websocket.Conn, message dtos.RoomCreationMessage) {

	messageByte, err := json.Marshal(message)

	if err != nil {
		fmt.Printf("Unable to marshal message %#v error %#v", message, err)
		return
	}

	wsMessage := dtos.GameMessage{
		MessageType: globals.MESSAGE_TYPE_ROOM_CREATION,
		Message:     messageByte,
	}

	wsMessageByte, err := json.Marshal(wsMessage)

	if err != nil {
		fmt.Printf("Unable to marshal message payload %#v %#v", wsMessage, err)
		return
	}

	conn.WriteMessage(websocket.TextMessage, wsMessageByte)
}
