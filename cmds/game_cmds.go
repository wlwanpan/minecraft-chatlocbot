package commands

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	constants "github.com/Ana-Wan/minecraft-save-script/constants"
	dbschemas "github.com/Ana-Wan/minecraft-save-script/db_schemas"

	wrapper "github.com/wlwanpan/minecraft-wrapper"
	"github.com/wlwanpan/minecraft-wrapper/events"
)

var runningPlayerCmds map[string]context.CancelFunc

// RunServer run the game server
func RunServer() *wrapper.Wrapper {

	getDbClient()

	wpr := wrapper.NewDefaultWrapper("server.jar", 1024, 1024)

	log.Println("Server loading ...")
	runningPlayerCmds = map[string]context.CancelFunc{}

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
			handlePlayerSay(gameEvent, w)
		}

		if gameEvent.String() == events.PlayerLeft {
			handleStopGOTO(w, gameEvent.Data["player_name"])
		}
	}
}

func handlePlayerSay(gameEvent events.GameEvent, w *wrapper.Wrapper) {
	playerName := gameEvent.Data["player_name"]
	playerMsg := gameEvent.Data["player_message"]

	switch {
	case strings.Contains(playerMsg, constants.SaveLocation):
		handleSaveLocation(w, playerMsg, playerName)

	case strings.Contains(playerMsg, constants.GetAllLocations):
		handleGetAllLoc(w, playerName)

	case strings.Contains(playerMsg, constants.GetLocation):
		handleGetLoc(w, playerMsg, playerName)

	case strings.Contains(playerMsg, constants.StartDirectionToDest):
		handleGOTOLoc(w, playerMsg, playerName)

	case strings.Contains(playerMsg, constants.StopDirectionToDest):
		handleStopGOTO(w, playerName)

	case strings.Contains(playerMsg, constants.DeleteLocation):
		handleDeleteLoc(w, playerMsg, playerName)
	}

	return
}

func handleSaveLocation(w *wrapper.Wrapper, playerMsg string, playerName string) {
	locName := getLocNameFromMsg(playerMsg, constants.SaveLocation)

	out, err := w.DataGet("entity", playerName)
	if err != nil {
		handleError(w, playerName, err)
		return
	}

	pos := out.Pos
	loc, err := saveLocation(playerName, locName, pos)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	msg := fmt.Sprintf("Saved location %s with position X Pos %f, Y Pos %f, Z Pos %f",
		loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)

	w.Say(msg)
}

func handleGetAllLoc(w *wrapper.Wrapper, playerName string) {
	locs, err := getAllLocations(playerName)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	for _, loc := range locs {
		msg := fmt.Sprintf("%s : x %0.1f y %0.1f z %0.1f", loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
		w.Say(msg)
	}
}

func handleGetLoc(w *wrapper.Wrapper, playerMsg string, playerName string) {
	locName := getLocNameFromMsg(playerMsg, constants.GetLocation)

	loc, err := getLocation(playerName, locName)
	if err != nil {
		handleError(w, playerName, err)
		return
	}

	msg := fmt.Sprintf("%s, x %0.1f y %0.1f z %0.1f", loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
	w.Say(msg)
}

func handleStopGOTO(w *wrapper.Wrapper, playerName string) {
	if cancelCmd, hasCmd := runningPlayerCmds[playerName]; hasCmd {
		cancelCmd()
		w.Tell(playerName, "Stopped direction to destination")
	}
}

func handleDeleteLoc(w *wrapper.Wrapper, playerMsg string, playerName string) {
	locName := getLocNameFromMsg(playerMsg, constants.DeleteLocation)

	deleteCount, err := deleteLocation(playerName, locName)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	if deleteCount == 0 {
		w.Tell(playerName, "Location not deleted, location could not be found for deletion")
		return
	}

	msg := fmt.Sprintf("Location successfully deleted")

	w.Say(msg)
}

func handleGOTOLoc(w *wrapper.Wrapper, playerMsg string, playerName string) {
	destName := getLocNameFromMsg(playerMsg, constants.StartDirectionToDest)
	dest, err := getLocation(playerName, destName)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	if _, hasCmd := runningPlayerCmds[playerName]; hasCmd {
		w.Tell(playerName, fmt.Sprintf("direction for a destination is already running"))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	runningPlayerCmds[playerName] = cancel

	go func(ctx context.Context, dest dbschemas.SavedLocation) {
		for {
			select {
			case <-time.Tick(time.Second * 2):
				out, err := w.DataGet("entity", playerName)
				if err != nil {
					handleError(w, playerName, err)
					cancel()
					return
				}

				//pos := out.Pos
				playerDir := getHorPlayerDirection(out.Rotation[0])
				//destPos := []float64{dest.XPos, dest.YPos, dest.ZPos}
				//directionToGo := getDirectionToGo(playerDir, pos, destPos)

				w.Tell(playerName, playerDir)
				//w.Tell(playerName, fmt.Sprintf("%v", directionToGo))
				// w.Tell(playerName, fmt.Sprintf("%f %f %f", dest.XPos, dest.YPos, dest.ZPos))

			case <-ctx.Done():
				return
			}
		}
	}(ctx, dest)
}

func handleError(w *wrapper.Wrapper, playerName string, err error) {
	if err != nil {
		log.Println(err)
		w.Tell(playerName, err.Error())
	}
}

func getLocNameFromMsg(msg string, eventmsg string) string {
	locName := strings.TrimPrefix(msg, eventmsg)
	log.Println(locName)
	return strings.TrimSpace(locName)
}

func getDirectionToGo(playerDir string, playerPos []float64, destPos []float64) string {

	direction := ""

	if playerPos[2] < destPos[2] {
		direction += constants.South
	} else {
		direction += constants.North
	}

	if playerPos[0] < destPos[0] {
		direction += constants.East
	} else {
		direction += constants.West
	}

	// directionMsg := ""

	// switch playerDir {
	// case "North":
	// case "South":
	// case "East":
	// case "West":
	// }

	return direction
}

func getHorPlayerDirection(horRotation float64) string {

	isNegative := false
	if horRotation < 0 {
		isNegative = true
	}

	degree := math.Abs(horRotation)

	switch {
	// North West
	case degree >= 120 && degree <= 150:
		if isNegative {
			return constants.NorthEast
		}
		return constants.NorthWest

	// North
	case degree >= 150 && degree <= 210:
		return constants.North

	// North East
	case degree >= 210 && degree <= 240:
		if isNegative {
			return constants.NorthWest
		}
		return constants.NorthEast
	// East
	case degree >= 420 && degree <= 270:
		if isNegative {
			return constants.West
		}
		return constants.East

	// South East
	case degree >= 300 && degree <= 330:
		if isNegative {
			return constants.SouthWest
		}
		return constants.SouthEast

	// South West
	case degree >= 60 && degree <= 30:
		if isNegative {
			return constants.SouthEast
		}
		return constants.SouthWest

	// West
	case degree >= 60 && degree <= 120:
		if isNegative {
			return constants.East
		}
		return constants.West

	default:
		return constants.South
	}
}
