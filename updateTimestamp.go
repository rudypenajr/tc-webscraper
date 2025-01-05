//go:build updatetimestamp

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
    clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	// Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to MongoDB!")

	// Get a handle for your collection
	var dbName = os.Getenv("MONGO_DB_NAME")
    var collectionName = os.Getenv("MONGO_COLLECTION")
    // collection := client.Database("tc-webscraper").Collection("episodes")
	collection := client.Database(dbName).Collection(collectionName)
	
	// Fetch all documents from the collection
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Fatal(err)
		}

		// Extract the `Date` field
		title, ok := doc["title"].(string)
		dateStr, ok := doc["date"].(string)
		if !ok || dateStr == "" {
			fmt.Println("Date not found or empty, skipping:", doc["_id"])
			continue
		}

		// Parse the date
		timestamp, err := parseDate(dateStr)
		if err != nil {
			fmt.Printf("Failed to parse date '%s': %v\n", dateStr, err)
			continue
		}

		// Update the document with the new Timestamp
		filter := bson.M{"_id": doc["_id"]}
		update := bson.M{"$set": bson.M{"timestamp": timestamp}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			fmt.Printf("Failed to update document %v: %v\n", doc["_id"], err)
			continue
		}
		fmt.Printf("Updated document %v - %v with %v timestamp %v\n", doc["_id"], title, dateStr, timestamp)
	}
}

func parseDate(dateStr string) (primitive.DateTime, error) {
	// Adjust this format to match your `Date` field
	t, err := time.Parse("January 2, 2006", dateStr)
	if err != nil {
		return 0, err
	}
	return primitive.NewDateTimeFromTime(t), nil
}
