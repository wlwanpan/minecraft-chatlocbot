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
			stopDestTicker(w, gameEvent.Data["player_name"])
		}
	}
}

func handlePlayerSay(gameEvent events.GameEvent, w *wrapper.Wrapper) {
	playerName := gameEvent.Data["player_name"]
	playerMsg := gameEvent.Data["player_message"]

	out, _ := w.DataGet("entity", playerName)
	w.Say(fmt.Sprintf("%v", out.Motion))
	w.Say(fmt.Sprintf("%v", out.Rotation))
	log.Println(out.Dimension)
	log.Println(out.Rotation)

	switch {
	case strings.Contains(playerMsg, constants.SaveLocation):
		locName := getLocNameFromMsg(playerMsg, constants.SaveLocation)

		out, err := w.DataGet("entity", playerName)
		if err != nil {
			handleError(w, playerName, err)
			return
		}

		pos := out.Pos
		loc := saveLocation(playerName, locName, pos)

		msg := fmt.Sprintf("Saved location %s with position X Pos %f, Y Pos %f, Z Pos %f",
			loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
		w.Say(msg)

	case strings.Contains(playerMsg, constants.GetAllLocations):
		locs, err := getAllLocations(playerName)

		if err != nil {
			handleError(w, playerName, err)
			return
		}

		for i, loc := range locs {
			msg := fmt.Sprintf("%d: %s, x %f y %f z %f | ", i, loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
			w.Say(msg)
		}

	case strings.Contains(playerMsg, constants.GetLocation):
		locName := getLocNameFromMsg(playerMsg, constants.GetLocation)

		loc, err := getLocation(playerName, locName)
		if err != nil {
			handleError(w, playerName, err)
			return
		}

		msg := fmt.Sprintf("%s, x %f y %f z %f", loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
		w.Say(msg)

	case strings.Contains(playerMsg, constants.StartDirectionToDest):
		destName := getLocNameFromMsg(playerMsg, constants.StartDirectionToDest)
		dest, err := getLocation(playerName, destName)

		if err != nil {
			handleError(w, playerName, err)
			return
		}

		runDestTicker(w, playerName, dest)

	case strings.Contains(playerMsg, constants.StopDirectionToDest):
		stopDestTicker(w, playerName)

		w.Tell(playerName, "Stopped direction to destination")
	}
	return
}

func stopDestTicker(w *wrapper.Wrapper, playerName string) {
	if cancelCmd, hasCmd := runningPlayerCmds[playerName]; hasCmd {
		cancelCmd()
	}
}

func runDestTicker(w *wrapper.Wrapper, playerName string, dest dbschemas.SavedLocation) {
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

				pos := out.Pos
				dir := getHorPlayerDirection(out.Rotation[0])
				destPos := []float64{dest.XPos, dest.YPos, dest.ZPos}
				directionToGo := getDirectionToGo(dir, pos, destPos)

				// w.Tell(playerName, dir)
				w.Tell(playerName, fmt.Sprintf("%v", directionToGo))
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
	locName := strings.TrimLeft(msg, eventmsg)
	return strings.TrimSpace(locName)
}

func getDirectionToGo(playerDir string, playerPos []float64, destPos []float64) string {

	directionMsg := ""

	switch playerDir {
	case "North":
		if playerPos[2] < destPos[2] {
			directionMsg += "back"
		} else {
			directionMsg += "forward"
		}
	case "South":
		if playerPos[2] < destPos[2] {
			directionMsg += "forward"
		} else {
			directionMsg += "back"
		}
	case "East":
		if playerPos[2] < destPos[2] {
			directionMsg += "right"
		} else {
			directionMsg += "left"
		}
	case "West":
		if playerPos[2] < destPos[2] {
			directionMsg += "left"
		} else {
			directionMsg += "right"
		}
	}

	return directionMsg
}

func getHorPlayerDirection(horRotation float64) string {

	isNegative := false
	if horRotation < 0 {
		isNegative = true
	}

	degree := math.Abs(horRotation)

	switch {
	case degree >= 135 && degree <= 225:
		return constants.North
	case degree >= 225 && degree <= 315:
		if isNegative {
			return "West"
		}
		return "East"
	case degree >= 45 && degree <= 135:
		if isNegative {
			return "East"
		}
		return "West"
	default:
		return "South"
	}
}
