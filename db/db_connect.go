package dbconnect

import (
	"context"
	"log"
	"time"

	dbschemas "./db_schemas"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
)

func getDbClient() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	if client == nil {
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(
			"mongodb+srv://anael_wan:8PrO8Z1X1dfEFfkQ@datastore.nzvab.mongodb.net/minecraft?retryWrites=true&w=majority",
		))
		handleError(err)
	}

	return client
}

// SaveLocation save location
func SaveLocation(name string, xPos float64, yPos float64, zPos float64) {

	loc := dbschemas.SavedLocation{name, xPos, yPos, zPos}

	dbClient := getDbClient()
	collection := dbClient.Database("minecraft").Collection("saved-locations")

	collection.InsertOne(context.TODO(), loc)
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
