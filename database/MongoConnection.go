package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var BookLogsCollection *mongo.Collection
var LibrarianLogsCollection *mongo.Collection
var BorrowingLogsCollection *mongo.Collection
var AuthLogsCollection *mongo.Collection

func ConnectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI("mongodb+srv://Teja:12345@cluster0.dzji5ih.mongodb.net/?retryWrites=true&w=majority")
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	fmt.Println("Connected to MongoDB")

	MongoClient = client
	BookLogsCollection = client.Database("library_portal_logging").Collection("book_logs")
	LibrarianLogsCollection = client.Database("library_portal_logging").Collection("librarian_logs")
	BorrowingLogsCollection = client.Database("library_portal_logging").Collection("borrowing_logs")
	AuthLogsCollection = client.Database("library_portal_logging").Collection("auth_logs")

	_, err = AuthLogsCollection.InsertOne(ctx, bson.M{
		"test": "MongoDB connected and auth_logs working",
		"time": time.Now(),
	})
	if err != nil {
		log.Fatalf(" Insert test log failed: %v", err)
	} else {
		fmt.Println("Test auth log inserted into MongoDB")
	}
}
