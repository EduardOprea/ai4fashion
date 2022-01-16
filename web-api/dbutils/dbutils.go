package dbutils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//const connString := os.Getenv("MONGODB_URL")

const dbDefaultUrl = "mongodb://localhost:27017"
const dbName = "ai4fashionDB"

//TODO check where the mongo client is initiated and make sure to close it
func InitiateMongoClient() *mongo.Client {
	var err error
	var client *mongo.Client
	fmt.Println("initiating database connection")
	username := os.Getenv("MONGO_ROOT_USERNAME")
	password := os.Getenv("MONGO_ROOT_PASSWORD")
	fmt.Printf("Connection username : %v ; password : %v \n", username, password)
	opts := options.Client()
	if len(username) > 0 && len(password) > 0 {
		fmt.Println("Username and password set => trying to auth")
		opts.SetAuth(options.Credential{
			Username: os.Getenv("MONGO_ROOT_USERNAME"),
			Password: os.Getenv("MONGO_ROOT_PASSWORD"),
		})
	}
	var dbConnString string
	if len(os.Getenv("MONGODB_URL")) > 0 && len(os.Getenv("MONGODB_PORT")) > 0 {
		dbConnString = fmt.Sprintf("%s:%s", os.Getenv("MONGODB_URL"), os.Getenv("MONGODB_PORT"))
	} else {
		dbConnString = dbDefaultUrl
	}
	fmt.Printf("Connecting to db using following url => %s\n", dbConnString)
	opts.ApplyURI(dbConnString)
	opts.SetMaxPoolSize(5)
	if client, err = mongo.Connect(context.Background(), opts); err != nil {
		fmt.Println("error connecting to mongoodb")
		fmt.Println(err.Error())
	}
	return client
}

func UploadFile(data []byte, filename string) {
	conn := InitiateMongoClient()
	defer conn.Disconnect(context.TODO())
	bucket, err := gridfs.NewBucket(
		conn.Database(dbName),
	)
	if err != nil {
		fmt.Printf("error creating grid fs bucket %v \n", err)
		return
	}
	uploadStream, err := bucket.OpenUploadStream(
		filename,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer uploadStream.Close()

	fileSize, err := uploadStream.Write(data)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Printf("Write file to DB was successful. File size: %d M\n", fileSize)
}
func GetImageProcessed(imageName string) []byte {
	conn := InitiateMongoClient()

	db := conn.Database(dbName)

	bucket, _ := gridfs.NewBucket(
		db,
	)
	var buf bytes.Buffer
	dStream, err := bucket.DownloadToStreamByName(imageName, &buf)
	if err != nil {
		fmt.Printf("Error getting image processed from db => %v\n", err)
		return nil
	}
	fmt.Printf("Image size to download: %v\n", dStream)
	return buf.Bytes()
}
func GetFileAndSave(imageName string) {
	conn := InitiateMongoClient()

	db := conn.Database(dbName)

	bucket, _ := gridfs.NewBucket(
		db,
	)
	var buf bytes.Buffer
	dStream, err := bucket.DownloadToStreamByName(imageName, &buf)
	if err != nil {
		fmt.Printf("Error getting image processed from db => %v\n", err)
	}
	fmt.Printf("Image size to download: %v\n", dStream)
	os.WriteFile(imageName, buf.Bytes(), 0600)
}
