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
		locName := getLocNameFromMsg(playerMsg, playersayevents.SaveLocation)

		out, err := w.DataGet("entity", playerName)
		if err != nil {
			handleError(w, err)
			return ""
		}

		pos := out.Pos
		loc := saveLocation(playerName, locName, pos)

		return fmt.Sprintf("Saved location %s with position X Pos %f, Y Pos %f, Z Pos %f",
			loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)

	case strings.Contains(playerMsg, playersayevents.GetAllLocations):
		locs, err := getAllLocations(playerName)

		if err != nil {
			handleError(w, err)
			return ""
		}

		var locsMsg strings.Builder
		for i, loc := range locs {
			msg := fmt.Sprintf("%d: %s, x %f y %f z %f | ", i, loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
			locsMsg.WriteString(msg)
		}

		return locsMsg.String()

	case strings.Contains(playerMsg, playersayevents.GetLocation):
		locName := getLocNameFromMsg(playerMsg, playersayevents.GetLocation)

		loc, err := getLocation(playerName, locName)
		if err != nil {
			handleError(w, err)
			return ""
		}

		msg := fmt.Sprintf("%s, x %f y %f z %f", loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)

		return msg
	}
	return ""
}

func handleError(w *wrapper.Wrapper, err error) {
	if err != nil {
		log.Println(err)
		w.Say("An error has occured, view console")
	}
}

func getLocNameFromMsg(msg string, eventmsg string) string {
	locName := strings.TrimLeft(msg, eventmsg)
	return strings.TrimSpace(locName)
}
