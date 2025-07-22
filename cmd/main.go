package main

import (
	"chickenswarm-server/internal/common"
	"chickenswarm-server/internal/engine"
	"flag"
	"fmt"
	"net/http"
)

func main() {
	isDevMode := flag.Bool("dev", true, "Run program in dev mode")
	flag.Parse()

	_, err := common.NewConfig(*isDevMode)
	if err != nil {
		panic(err)
	}

	address := fmt.Sprintf("%s:%d", "0.0.0.0", 3001)
	game := engine.NewGame()

	http.HandleFunc("/ws", game.Server.UpgradeConnection)
	http.ListenAndServe(address, nil)
}
