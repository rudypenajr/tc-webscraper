//go:build addEmbeddings

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Episode struct {
	ID                    	string   			`bson:"_id,omitempty"` // Unique identifier
	Url     				string 				`bson:"url,omitempty"`
	Title					string				`bson:"title,omitempty"`
	EpisodeNo 				string				`bson:"episode_no,omitempty"`
	Date					string				`bson:"date,omitempty"`
	Timestamp				primitive.DateTime	`bson:"timestamp"`  // ISO 8601 timestamp
	Guests 					[]string			`bson:"guests,omitempty"`
	Top5ComparisonYear 		string				`bson:"top_5_comparison_year,omitempty"`
	Notes  					string				`bson:"notes,omitempty"`
}

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	openaiKey := os.Getenv("OPENAI_API_KEY") // OpenAI API Key
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
	fmt.Println("✅ Connected to MongoDB!")

	// Get a handle for your collection
	dbName := os.Getenv("MONGO_DB_NAME")
	collectionName := os.Getenv("MONGO_COLLECTION")
	collection := client.Database(dbName).Collection(collectionName)

	// Fetch all documents from the collection
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	// Connect to OpenAI
	openaiClient := openai.NewClient(openaiKey)

	for cursor.Next(context.TODO()) {
	var episode Episode
	if err := cursor.Decode(&episode); err != nil {
		log.Fatal(err)
	}

	// Generate vector embedding for the episode
	embedding, err := generateEmbedding(openaiClient, episode)
	if err != nil {
		fmt.Printf("❌ Failed to generate embedding for '%s': %v\n", episode.Title, err)
		continue
	}

	// Update MongoDB with the new embedding
	filter := bson.M{"_id": episode.ID}
	update := bson.M{"$set": bson.M{"embedding": embedding}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Printf("❌ Failed to update document %v: %v\n", episode.ID, err)
		continue
	}

	fmt.Printf("✅ Updated document %v - '%s' with vector embedding\n", episode.ID, episode.Title)
}
}

func generateEmbedding(client *openai.Client, episode Episode) ([]float32, error) {
	// Combine relevant fields into a single text input
	textInput := fmt.Sprintf(
		"Title: %s. Guests: %s. Year: %s. Notes: %s",
		episode.Title,
		strings.Join(episode.Guests, ", "), // Convert guest slice to a string
		episode.Top5ComparisonYear,
		episode.Notes,
	)

	resp, err := client.CreateEmbeddings(context.TODO(), openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2, // OpenAI embedding model
		Input: []string{textInput},   // Use combined text
	})
	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}

// Function to generate embeddings using OpenAI API - Just Title
// func generateEmbedding(client *openai.Client, text string) ([]float32, error) {
// 	resp, err := client.CreateEmbeddings(context.TODO(), openai.EmbeddingRequest{
// 		Model: openai.AdaEmbeddingV2, // Adjust model as needed
// 		Input: []string{text},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp.Data[0].Embedding, nil
// }
