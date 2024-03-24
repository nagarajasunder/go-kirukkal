package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nagarajasunder/go-kirukkal/internal/dtos"
	"github.com/nagarajasunder/go-kirukkal/internal/services"
)

func SetRoomRoutes(router *httprouter.Router) {
	router.GET("/subscribe/room", SubscribeToRoomUpdates)
}

func SubscribeToRoomUpdates(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Printf("Unable to subscribe to room updates err %#v", err)
		return
	}

	services.SubscribeForRoomUpdates(conn)

	resp := dtos.NetworkResponse{
		Code:    http.StatusAccepted,
		Status:  "Success",
		Message: "Subscribed to room updates",
	}

	json.NewEncoder(w).Encode(resp)
}
