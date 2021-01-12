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

func saveLocation(playerName string, locName string, pos []float64) (dbschemas.SavedLocation, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	loc := dbschemas.SavedLocation{playerName, locName, pos[0], pos[1], pos[2]}

	opts := options.FindOneAndUpdate()
	opts.SetUpsert(true)
	opts.SetReturnDocument(options.After)
	res := collection.FindOneAndUpdate(ctx, bson.M{"playername": playerName, "locationname": locName}, bson.M{"$set": loc}, opts)
	err := res.Err()

	return loc, handleDBErrors(err)
}

func getAllLocations(playerName string) ([]dbschemas.SavedLocation, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	var locs []dbschemas.SavedLocation

	resultCursor, err := collection.Find(ctx, bson.M{"playername": playerName})
	resultCursor.All(ctx, &locs)

	return locs, handleDBErrors(err)
}

func getLocation(playerName string, locName string) (dbschemas.SavedLocation, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	var loc dbschemas.SavedLocation
	err := collection.FindOne(ctx, bson.M{"playername": playerName, "locationname": locName}).Decode(&loc)

	return loc, handleDBErrors(err)
}

func deleteLocation(playerName string, locName string) (int64, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	log.Println("test del 123")
	log.Println(locName)
	res, err := collection.DeleteOne(ctx, bson.M{"playername": playerName, "locationname": locName})

	deleteCount := res.DeletedCount

	return deleteCount, handleDBErrors(err)
}

// Helpers
func getSavedLocationsDatabaseCollection() *mongo.Collection {
	dbClient := getDbClient()
	return dbClient.Database("minecraft").Collection("saved-locations")
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
