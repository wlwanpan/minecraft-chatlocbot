package main

import (
	"flag"
	"log"

	cmds "github.com/Ana-Wan/minecraft-save-script/cmds"
	"github.com/joho/godotenv"
)

func main() {

	memPtr := flag.Int("mem", 1024, "memory")
	maxMemPtr := flag.Int("maxmem", 1024, "max memory")
	pathToServerJarPtr := flag.String("path", "server.jar", "path to server jar")

	flag.Parse()

	log.Print(*memPtr, *maxMemPtr, *pathToServerJarPtr)

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	wpr := cmds.RunServer(*memPtr, *maxMemPtr, *pathToServerJarPtr)

	defer wpr.Stop()

	cmds.HandleGameEvents(wpr)
}
