package commands

import (
	"context"
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

func saveLocation(playerName string, locationName string, pos []float64) bool {

	loc := dbschemas.SavedLocation{playerName, locationName, pos[0], pos[1], pos[2]}

	collection := getSavedLocationsDatabaseCollection()

	result, err := collection.InsertOne(context.TODO(), loc)

	return err == nil && result != nil
}

func getAllLocations(playerName string) ([]dbschemas.SavedLocation, error) {

	collection := getSavedLocationsDatabaseCollection()
	ctx := context.TODO()

	var locations []dbschemas.SavedLocation
	resultCursor, err := collection.Find(ctx, bson.M{"playername": playerName})
	resultCursor.All(ctx, &locations)

	return locations, err
}

func getSavedLocationsDatabaseCollection() *mongo.Collection {
	dbClient := getDbClient()
	return dbClient.Database("minecraft").Collection("saved-locations")
}

func getDbClient() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	if client == nil {
		username := os.Getenv("USERNAME")
		password := os.Getenv("PASSWORD")
		dbname := os.Getenv("DB_NAME")
		uri := fmt.Sprintf("mongodb+srv://%s:%s@datastore.nzvab.mongodb.net/%s?retryWrites=true&w=majority", username, password, dbname)

		client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))

		if err != nil {
			log.Println(err)
		}
	}

	return client
}
