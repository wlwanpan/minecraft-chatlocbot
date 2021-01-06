package main

import (
	"log"

	cmds "github.com/Ana-Wan/minecraft-save-script/cmds"
	"github.com/joho/godotenv"
)

func main() {
	wpr := cmds.RunServer()

	defer wpr.Stop()

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	cmds.HandleGameEvents(wpr)
}
