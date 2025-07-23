package main

import (
	"chickenswarm-server/internal/engine"
	"fmt"
	"net/http"
)

func main() {
	address := fmt.Sprintf("%s:%d", "0.0.0.0", 3001)
	game := engine.NewGame()

	http.HandleFunc("/ws", game.Server.UpgradeConnection)
	http.ListenAndServe(address, nil)
}
