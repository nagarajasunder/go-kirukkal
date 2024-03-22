package main

import (
	"fmt"
	"net/http"

	"github.com/nagarajasunder/go-kirukkal/internal/handlers"
)

func main() {
	fmt.Println("Starting the server....")
	http.ListenAndServe(":3000", handlers.New())
}
