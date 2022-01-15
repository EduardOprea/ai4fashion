package dbutils

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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
	fmt.Println("initiating database connectiion")
	username := os.Getenv("MONGODB_USERNAME")
	password := os.Getenv("MONGODB_dockPASSWORD")
	fmt.Printf("Connection username : %v ; password : %v \n", username, password)
	opts := options.Client()
	opts.SetAuth(options.Credential{
		Username: os.Getenv("MONGODB_USERNAME"),
		Password: os.Getenv("MONGODB_PASSWORD"),
	})

	var dbConnString string
	if len(os.Getenv("MONGODB_URL")) > 0 {
		dbConnString = os.Getenv("MONGODB_URL")
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
func DownloadFile(fileName string) {
	conn := InitiateMongoClient()

	db := conn.Database(dbName)

	bucket, _ := gridfs.NewBucket(
		db,
	)
	var buf bytes.Buffer
	dStream, err := bucket.DownloadToStreamByName(fileName, &buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("File size to download: %v\n", dStream)
	ioutil.WriteFile(fileName, buf.Bytes(), 0600)
}
