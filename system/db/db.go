package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/lpernett/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func Connect() bool {
	tlsConfig := load_tls_config()
	uri := get_uri()
	opts := options.Client().ApplyURI(uri).SetTLSConfig(tlsConfig)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		fmt.Printf("Error connecting db: %v\n", err)
		return false
	}
	Client = client
	return true
}
func Disconnect() {
	if err := Client.Disconnect(context.TODO()); err != nil {
		fmt.Printf("Error disconnecting db: %v\n", err)
	}
}
func GetDatabase() *mongo.Database {
	return Client.Database(os.Getenv("VERSE_DB"))
}
func ListCollections() []string {
	db := GetDatabase()
	cols, err := db.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		fmt.Printf("Error listing collections: %v", err)
	}
	return cols
}

func CreateCollection(collectionName string) bool {
	err := GetDatabase().CreateCollection(context.TODO(), collectionName)
	if err != nil {
		fmt.Printf("Error creating collection: %v\n", err)
		return false
	}
	return true
}

func DeleteCollection(collectionName string) bool {
	err := GetDatabase().Collection(collectionName).Drop(context.TODO())
	if err != nil {
		fmt.Printf("Error deleting collection: %v\n", err)
		return false
	}
	return true
}
func InsertOneIntoCollection(collectionName string, item map[string]string) bool {
	_, err := GetDatabase().Collection(collectionName).InsertOne(context.TODO(), item)
	if err != nil {
		fmt.Printf("Error writing into collection: %v\n", err)
		return false
	}
	return true
}
func UpsertItemInCollection(collectionName string, item map[string]string, key string) bool {
	filter := bson.M{key: item[key]}
	update := bson.M{
		"$set": item,
	}
	opts := options.Update().SetUpsert(true)
	_, err := GetDatabase().Collection(collectionName).UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		fmt.Printf("Error updating collection: %v\n", err)
		return false
	}
	return true
}
func FindOneFromCollection(collectionName string, item map[string]string) map[string]string {
	res := GetDatabase().Collection(collectionName).FindOne(context.TODO(), item)
	var result map[string]string
	if err := res.Decode(&result); err != nil {
		log.Println("Error decoding result:", err)
		return nil
	}
	return result
}
func load_tls_config() *tls.Config {
	caFile := "cert/fullchain.pem"
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		panic("Error: CA file must be in PEM format")
	}
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	return tlsConfig
}
func get_uri() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	username := os.Getenv("VERSE_USERNAME")
	password := os.Getenv("VERSE_PASSWORD")
	authDB := os.Getenv("VERSE_DB")
	host := os.Getenv("VERSE_HOST")
	port := os.Getenv("VERSE_PORT")

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=%s", username, password, host, port, authDB)
	return uri
}
