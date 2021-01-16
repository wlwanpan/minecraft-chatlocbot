package main

import (
	"flag"
	"fmt"
	"log"

	cmds "github.com/Ana-Wan/minecraft-chatlocbot/cmds"
	"github.com/gotk3/gotk3/gtk"
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

	id := cmds.GetWorldID()
	log.Println(fmt.Sprintf("World Id = %s", id))

	// testing gui

	gtk.Init(nil)
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Simple Example")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	//

	//wpr := cmds.RunServer(*memPtr, *maxMemPtr, *pathToServerJarPtr)

	//defer wpr.Stop()

	//cmds.HandleGameEvents(wpr, id)
}
