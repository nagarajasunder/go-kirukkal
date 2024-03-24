package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func New() http.Handler {

	router := httprouter.New()
	setPingRoutes(router)
	setGameRoutes(router)
	SetRoomRoutes(router)

	return router
}
