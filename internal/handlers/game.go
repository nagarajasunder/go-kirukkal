package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/nagarajasunder/go-kirukkal/internal/services"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func setGameRoutes(router *httprouter.Router) {
	router.GET("/rooms/:name/create", CreateRoom)
	router.GET("/room/:room_id/players/:player_name/create", CreateNewPlayer)
}

func CreateRoom(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	roomName := p.ByName("name")

	service := services.NewGameService()

	resp, err := service.CreateRoom(roomName)

	if err != nil {
		http.Error(w, "Unable to create room", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)

}

func CreateNewPlayer(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	roomId := p.ByName("room_id")
	playerName := p.ByName("player_name")
	isAdminQueryParam := r.URL.Query().Get("is_admin")
	isAdmin, err := strconv.ParseBool(isAdminQueryParam)
	if err != nil {
		isAdmin = false
	}

	if roomId == "" || playerName == "" {
		http.Error(w, "Room ID & player name cannot be empty", http.StatusBadRequest)
		return
	}

	roomService := services.NewGameService()

	exists := roomService.DoesRoomExists(roomId)

	if !exists {
		err := fmt.Sprintf("No room exists with id %s", roomId)
		http.Error(w, err, http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		conn.Close()
		http.Error(w, "Unable to create new player", http.StatusInternalServerError)
		return
	}

	roomService.CreateNewPlayer(roomId, playerName, isAdmin, conn)

}
