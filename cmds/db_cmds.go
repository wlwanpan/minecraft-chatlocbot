package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	dbschemas "github.com/Ana-Wan/minecraft-save-script/db_schemas"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
)

func saveLocation(worldID [16]byte, playerName string, locName string, pos []float64) (dbschemas.SavedLocation, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	loc := dbschemas.SavedLocation{
		WorldId:      worldID,
		SavedBy:      playerName,
		LocationName: locName,
		XPos:         pos[0],
		YPos:         pos[1],
		ZPos:         pos[2],
	}

	opts := options.FindOneAndUpdate()
	opts.SetUpsert(true)
	opts.SetReturnDocument(options.After)
	res := collection.FindOneAndUpdate(ctx, bson.M{"worldid": playerName, "locationname": locName, "savedby": playerName}, bson.M{"$set": loc}, opts)
	err := res.Err()

	return loc, handleDBErrors(err)
}

func getAllLocations(worldID [16]byte) ([]dbschemas.SavedLocation, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	var locs []dbschemas.SavedLocation

	resultCursor, err := collection.Find(ctx, bson.M{"worldid": worldID})
	resultCursor.All(ctx, &locs)

	return locs, handleDBErrors(err)
}

func getLocation(worldID [16]byte, locName string) (dbschemas.SavedLocation, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	var loc dbschemas.SavedLocation
	err := collection.FindOne(ctx, bson.M{"worldid": worldID, "locationname": locName}).Decode(&loc)

	return loc, handleDBErrors(err)
}

func deleteLocation(worldID [16]byte, locName string) (int64, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	res, err := collection.DeleteOne(ctx, bson.M{"worldid": worldID, "locationname": locName})

	deleteCount := res.DeletedCount

	return deleteCount, handleDBErrors(err)
}

// DB Helpers
func getSavedLocationsDatabaseCollection() *mongo.Collection {
	dbClient := getDbClient()
	return dbClient.Database("catlocs").Collection("saved-locations")
}

func getDbClient() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	if client == nil {
		username := os.Getenv("DB_USERNAME")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		uri := fmt.Sprintf("mongodb+srv://%s:%s@datastore.nzvab.mongodb.net/%s?retryWrites=true&w=majority", username, password, dbname)

		client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))

		if err != nil {
			log.Println("Failed to connect to database ...")
			log.Fatal(err)
		}
	}

	return client
}

func handleDBErrors(err error) error {
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return errors.New("Location does not exist")
		default:
			return err
		}
	}

	return err
}
