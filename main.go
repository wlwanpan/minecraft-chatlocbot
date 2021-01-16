package main

import (
	"flag"
	"fmt"
	"log"

	cmds "github.com/Ana-Wan/minecraft-save-script/cmds"
	"github.com/joho/godotenv"
)

func main() {

	memPtr := flag.Int("mem", 1024, "memory")
	maxMemPtr := flag.Int("maxmem", 1024, "max memory")
	pathToServerJarPtr := flag.String("path", "server.jar", "path to server jar")

	flag.Parse()

	log.Println(fmt.Sprintf("Loading file %s", *pathToServerJarPtr))

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, server closing ...")
		return
	}

	id, idStr := cmds.GetWorldID()
	log.Println(fmt.Sprintf("World Id = %s", idStr))

	wpr := cmds.RunServer(*memPtr, *maxMemPtr, *pathToServerJarPtr)

	defer wpr.Stop()

	cmds.HandleGameEvents(wpr, id)
}
