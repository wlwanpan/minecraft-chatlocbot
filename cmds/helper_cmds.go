package commands

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	constants "github.com/Ana-Wan/minecraft-chatlocbot/constants"
	uuid "github.com/google/uuid"
	wrapper "github.com/wlwanpan/minecraft-wrapper"
)

// GetWorldID returns the world id if worldid.txt file exists, else creates a new worldid.txt file with a new id
func GetWorldID() [16]byte {

	file, err := os.OpenFile("worldid.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	defer file.Close()

	if err != nil {
		log.Print(err)
		return uuid.Nil
	}

	idStr := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txtLine := scanner.Text()
		if strings.Contains(txtLine, "worldId=") {
			idStr = strings.TrimPrefix(txtLine, "worldId=")
			break
		}
	}

	if idStr == "" {
		newID, _ := uuid.NewUUID()
		if _, err := file.Write([]byte(fmt.Sprintf("worldId=%s", newID.String()))); err != nil {
			file.Close()
			log.Fatal(err)
		}

		file.Close()
		return newID
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not retrieve world id, the worldid.txt file might have been previously modified\nError: %v", err))
	}

	return id
}

func handleError(w *wrapper.Wrapper, playerName string, err error) {
	if err != nil {
		log.Println(err)
		w.Tell(playerName, err.Error())
	}
}

func getLocNameFromMsg(msg string, eventmsg string) string {
	locName := strings.TrimPrefix(msg, eventmsg)

	return strings.TrimSpace(locName)
}

func getCoordsInfoFromMsg(msg string, eventmsg string) []string {
	msgSuffix := strings.TrimPrefix(msg, eventmsg)
	msgNoWhiteSpace := strings.TrimSpace(msgSuffix)

	coordsInfo := strings.Split(msgNoWhiteSpace, " ")

	return coordsInfo
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

	playerHorRotationDeg := playerRotationDeg[0]
	playerFacingDeg := math.Abs(playerHorRotationDeg)

	if playerHorRotationDeg > 0 {
		playerFacingDeg = 360 - playerFacingDeg
	}

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

	if (absDegDifference < 180 && degDifference < 0) ||
		(absDegDifference > 180 && degDifference > 0) {
		return constants.Left
	}

	if (absDegDifference > 180 && degDifference < 0) ||
		(absDegDifference < 180 && degDifference > 0) {
		return constants.Right
	}

	return constants.Back
}
