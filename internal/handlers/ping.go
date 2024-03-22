package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func setPingRoutes(router *httprouter.Router) {
	router.GET("/ping", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data, err := json.Marshal("pong")
		if err != nil {
			fmt.Println("Unable to marshall response err: %#v", err)
			return
		}
		w.Write(data)
	})
}
