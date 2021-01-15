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
func RunServer(memory int, maxMemory int, pathToServerJar string) *wrapper.Wrapper {

	getDbClient()

	wpr := wrapper.NewDefaultWrapper(pathToServerJar, memory, maxMemory)

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

func handleStopGOTO(w *wrapper.Wrapper, playerName string) {
	if cancelCmd, hasCmd := runningPlayerCmds[playerName]; hasCmd {
		cancelCmd()
		delete(runningPlayerCmds, playerName)
		w.Tell(playerName, "Stopped direction to destination")
	}
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

				pos := out.Pos
				// playerDir := getHorPlayerDirection(out.Rotation[0])
				destPos := []float64{dest.XPos, dest.YPos, dest.ZPos}
				directionToGo := getDirectionToGo(out.Rotation, pos, destPos)

				if directionToGo == constants.Reached {
					handleStopGOTO(w, playerName)
					w.Tell(playerName, fmt.Sprintf(directionToGo))
					return
				}

				w.Tell(playerName, directionToGo)

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

func getDirectionToGo(playerRotationDeg []float64, playerPos []float64, destPos []float64) string {

	playerX := playerPos[2]
	playerY := playerPos[0]
	destX := destPos[2]
	destY := destPos[0]

	deltaX := destX - playerX
	deltaY := destY - playerY
	thetaRad := math.Atan2(deltaY, deltaX)
	deg := (thetaRad * (180 / math.Pi)) + 360
	destDeg := math.Mod(deg, 360)

	// log.Println(destDegFmt)

	playerHorRotationDeg := playerRotationDeg[0]
	playerFacingDeg := math.Abs(playerHorRotationDeg)

	if playerHorRotationDeg > 0 {
		playerFacingDeg = 360 - playerFacingDeg
	}

	// log.Println(playerHorRotationDeg)
	// log.Println(playerFacingDeg)

	xDiff := playerX - destX
	yDiff := playerY - destY

	if math.Abs(xDiff) < 30 && math.Abs(yDiff) < 10 {
		return constants.Reached
	}

	degDifference := playerFacingDeg - destDeg
	absDegDifference := math.Abs(degDifference)

	if absDegDifference < 30 {
		return constants.Forward
	}

	// if (absDegDifference < 180 && degDifference < 0) ||
	// 	(absDegDifference > 180 && degDifference > 0) {
	// 	return constants.Left
	// }

	// if (absDegDifference > 180 && degDifference < 0) ||
	// 	(absDegDifference < 180 && degDifference > 0) {
	// 	return constants.Right
	// }

	if absDegDifference < 180 {
		if degDifference < 0 {
			return constants.Left
		}
		return constants.Right
	}

	if absDegDifference > 180 {
		if degDifference < 0 {
			return constants.Right
		}
		return constants.Left
	}

	return constants.Back
}
