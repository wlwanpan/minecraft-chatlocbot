package main

import (
	"fmt"
	"log"

	dbconnect "./db_schemas"
	playersayevents "./events"

	wrapper "github.com/wlwanpan/minecraft-wrapper"
	"github.com/wlwanpan/minecraft-wrapper/events"
)

func main() {
	wpr := runServer()

	handleGameEvents(wpr)
}

func runServer() *wrapper.Wrapper {
	wpr := wrapper.NewDefaultWrapper("server.jar", 1024, 1024)

	wpr.RegisterStateChangeCBs(stateLog)
	wpr.Start()
	defer wpr.Stop()

	return wpr
}

func stateLog(w *wrapper.Wrapper, from events.Event, to events.Event) {
	log.Printf("%s -> %s", from, to)
}

func handleGameEvents(w *wrapper.Wrapper) {
	// Listening to game events...
	for {
		gameEvent := <-w.GameEvents()
		log.Println(gameEvent.String())

		if gameEvent.String() == events.PlayerSay {
			out, err := w.DataGet("entity", "LeanaWan")
			handleError(err)

			handlePlayerSay(gameEvent.Data["player_message"], out)
			pos := out.Pos
			w.Say(fmt.Sprintf("X Pos %f, Y Pos %f, Z Pos %f /n", pos[0], pos[1], pos[2]))
		}
	}
}

func handlePlayerSay(say string, data *wrapper.DataGetOutput) {
	switch {
	case say == playersayevents.SaveLocation:
		pos := data.Pos
		dbconnect.SaveLocation(say, pos[0], pos[1], pos[2])
	}
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
