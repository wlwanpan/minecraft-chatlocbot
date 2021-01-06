package commands

import (
	"fmt"
	"log"
	"strings"

	playersayevents "github.com/Ana-Wan/minecraft-save-script/events"

	wrapper "github.com/wlwanpan/minecraft-wrapper"
	"github.com/wlwanpan/minecraft-wrapper/events"
)

// RunServer run the game server
func RunServer() *wrapper.Wrapper {
	wpr := wrapper.NewDefaultWrapper("server.jar", 1024, 1024)

	log.Println("Server loading ...")

	err := wpr.Start()
	<-wpr.Loaded()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server loaded")

	return wpr
}

// HandleGameEvents handles game events
func HandleGameEvents(w *wrapper.Wrapper) {
	// Listening to game events...
	for {
		gameEvent := <-w.GameEvents()
		log.Println(gameEvent.String())

		if gameEvent.String() == events.PlayerSay {
			res := handlePlayerSay(gameEvent, w)
			w.Say(res)
		}
	}
}

func handlePlayerSay(gameEvent events.GameEvent, w *wrapper.Wrapper) string {
	playerName := gameEvent.Data["player_name"]
	playerMsg := gameEvent.Data["player_message"]

	switch {
	case strings.Contains(playerMsg, playersayevents.SaveLocation):
		locationMsg := strings.ReplaceAll(playerMsg, playersayevents.SaveLocation, "")
		locationName := strings.TrimSpace(locationMsg)

		out, err := w.DataGet("entity", playerName)
		if err != nil {
			handleError(w, err)
			return ""
		}

		pos := out.Pos
		result := saveLocation(playerName, locationName, pos)

		if result {
			return fmt.Sprintf("Saved location %s with position X Pos %f, Y Pos %f, Z Pos %f /n",
				locationName, pos[0], pos[1], pos[2])
		}

		return "Failed to save location, please try again"

	case strings.Contains(playerMsg, playersayevents.GetAllLocations):
		locations, err := getAllLocations(playerName)

		if err != nil {
			handleError(w, err)
			return ""
		}

		var allLocMsg strings.Builder
		for i, loc := range locations {
			msg := fmt.Sprintf("%d: %s, x %f y %f z %f | ", i, loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
			allLocMsg.WriteString(msg)
		}

		return allLocMsg.String()
	}

	return ""
}

func handleError(w *wrapper.Wrapper, err error) {
	if err != nil {
		log.Println(err)
		w.Say("An error has occured, view console")
	}
}
