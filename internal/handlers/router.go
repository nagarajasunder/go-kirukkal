package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func New() http.Handler {

	router := httprouter.New()
	setPingRoutes(router)
	setRoomRoutes(router)

	return router
}
