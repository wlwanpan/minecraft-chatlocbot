package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	constants "github.com/Ana-Wan/minecraft-chatlocbot/constants"
	dbschemas "github.com/Ana-Wan/minecraft-chatlocbot/db_schemas"

	wrapper "github.com/wlwanpan/minecraft-wrapper"
	"github.com/wlwanpan/minecraft-wrapper/events"
)

var runningPlayerCmds map[string]context.CancelFunc

// RunServer run the game server
func RunServer(memory int, maxMemory int, pathToServerJar string) *wrapper.Wrapper {
	getDbClient()

	wpr := wrapper.NewDefaultWrapper(pathToServerJar, memory, maxMemory)

	log.Println("Loading server ...")
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
func HandleGameEvents(w *wrapper.Wrapper, worldID [16]byte) {

	// Listening to game events...
	for {
		gameEvent := <-w.GameEvents()
		log.Println(gameEvent.String())

		if gameEvent.String() == events.PlayerSay {
			handlePlayerSay(w, gameEvent, worldID)
		}

		if gameEvent.String() == events.PlayerLeft {
			handleStopGOTO(w, gameEvent.Data["player_name"])
		}
	}
}

func handlePlayerSay(w *wrapper.Wrapper, gameEvent events.GameEvent, worldID [16]byte) {
	playerName := gameEvent.Data["player_name"]
	playerMsg := gameEvent.Data["player_message"]

	switch {
	// Player save loc
	case strings.Contains(playerMsg, constants.SaveLocation):
		handleSaveLocation(w, worldID, playerMsg, playerName)

	// Player save coords
	case strings.Contains(playerMsg, constants.SaveCoords):
		handleSaveCoords(w, worldID, playerMsg, playerName)

	// Player get all locs
	case strings.Contains(playerMsg, constants.GetAllLocations):
		handleGetAllLoc(w, worldID, playerName)

	// Player get specific loc
	case strings.Contains(playerMsg, constants.GetLocation):
		handleGetLoc(w, worldID, playerMsg, playerName)

	// Player start direction to dest
	case strings.Contains(playerMsg, constants.StartDirectionToDest):
		handleGOTOLoc(w, worldID, playerMsg, playerName)

	// Player stops direction to dest
	case strings.Contains(playerMsg, constants.StopDirectionToDest):
		handleStopGOTO(w, playerName)

	// Player delete specific location
	case strings.Contains(playerMsg, constants.DeleteLocation):
		handleDeleteLoc(w, worldID, playerMsg, playerName)

	// Player deletes all location in current world
	case strings.Contains(playerMsg, constants.DeleteAllLocations):
		handleDeleteAllLocs(w, worldID, playerName)
	}

	return
}

func handleSaveLocation(w *wrapper.Wrapper, worldID [16]byte, playerMsg string, playerName string) {
	locName := getLocNameFromMsg(playerMsg, constants.SaveLocation)

	out, err := w.DataGet("entity", playerName)
	if err != nil {
		handleError(w, playerName, err)
		return
	}

	pos := out.Pos
	loc, err := saveLocation(worldID, playerName, locName, pos)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	msg := fmt.Sprintf("%s saved location %s with position X %0.1f Y %0.1f Z %0.1f",
		playerName, loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)

	w.Say(msg)
}

func handleSaveCoords(w *wrapper.Wrapper, worldID [16]byte, playerMsg string, playerName string) {
	coordsInfo := getCoordsInfoFromMsg(playerMsg, constants.SaveLocation)

	if len(coordsInfo) != 4 {
		w.Say(fmt.Sprintf("command should have the format: %s <location_name> <x_position> <y_position> <z_position>", constants.SaveCoords))
		return
	}

	xPos, errX := strconv.ParseFloat(coordsInfo[1], 64)
	if errX != nil {
		w.Say("X position should be a valid number")
		return
	}

	yPos, errY := strconv.ParseFloat(coordsInfo[2], 64)
	if errY != nil {
		w.Say("Y position should be a valid number")
		return
	}

	zPos, errY := strconv.ParseFloat(coordsInfo[3], 64)
	if errX != nil {
		w.Say("Z position should be a valid number")
		return
	}

	pos := []float64{xPos, yPos, zPos}
	loc, err := saveLocation(worldID, playerName, coordsInfo[0], pos)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	msg := fmt.Sprintf("%s saved location %s with position X %0.1f Y %0.1f Z %0.1f",
		playerName, loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)

	w.Say(msg)
}

func handleGetAllLoc(w *wrapper.Wrapper, worldID [16]byte, playerName string) {
	locs, err := getAllLocations(worldID)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	if len(locs) == 0 {
		w.Tell(playerName, "No location saved")
		return
	}

	for _, loc := range locs {
		msg := fmt.Sprintf("%s : X %0.1f Y %0.1f Z %0.1f", loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
		w.Tell(playerName, msg)
	}
}

func handleGetLoc(w *wrapper.Wrapper, worldID [16]byte, playerMsg string, playerName string) {
	locName := getLocNameFromMsg(playerMsg, constants.GetLocation)

	loc, err := getLocation(worldID, locName)
	if err != nil {
		handleError(w, playerName, err)
		return
	}

	msg := fmt.Sprintf("%s, x %0.1f y %0.1f z %0.1f", loc.LocationName, loc.XPos, loc.YPos, loc.ZPos)
	w.Tell(playerName, msg)
}

func handleDeleteLoc(w *wrapper.Wrapper, worldID [16]byte, playerMsg string, playerName string) {
	locName := getLocNameFromMsg(playerMsg, constants.DeleteLocation)

	deleteCount, err := deleteLocation(worldID, locName)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	if deleteCount == 0 {
		w.Tell(playerName, "Location not deleted, location could not be found for deletion")
		return
	}

	msg := fmt.Sprintf("Location %s successfully deleted by %s", locName, playerName)

	w.Say(msg)
}

func handleDeleteAllLocs(w *wrapper.Wrapper, worldID [16]byte, playerName string) {

	deleteCount, err := deleteAllLocations(worldID)

	if err != nil {
		handleError(w, playerName, err)
		return
	}

	if deleteCount == 0 {
		w.Tell(playerName, "No location deleted")
		return
	}

	msg := fmt.Sprintf("%d locations deleted by %s", deleteCount, playerName)

	w.Say(msg)
}

func handleStopGOTO(w *wrapper.Wrapper, playerName string) {
	if cancelCmd, hasCmd := runningPlayerCmds[playerName]; hasCmd {
		cancelCmd()
		delete(runningPlayerCmds, playerName)
		w.Tell(playerName, "Stopped direction to destination")
	}
}

func handleGOTOLoc(w *wrapper.Wrapper, worldID [16]byte, playerMsg string, playerName string) {
	destName := getLocNameFromMsg(playerMsg, constants.StartDirectionToDest)
	dest, err := getLocation(worldID, destName)

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
