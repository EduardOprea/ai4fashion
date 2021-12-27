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

const connString = "mongodb://localhost:27017"
const dbName = "ai4fashionDB"

//TODO check where the mongo client is initiated and make sure to close it
func InitiateMongoClient() *mongo.Client {
	var err error
	var client *mongo.Client
	opts := options.Client()
	opts.ApplyURI(connString)
	opts.SetMaxPoolSize(5)
	if client, err = mongo.Connect(context.Background(), opts); err != nil {
		fmt.Println(err.Error())
	}
	return client
}

func UploadFile(data []byte, filename string) {
	conn := InitiateMongoClient()
	//defer conn.
	bucket, err := gridfs.NewBucket(
		conn.Database(dbName),
	)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	uploadStream, err := bucket.OpenUploadStream(
		filename,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
func InsertInDBTest() error {
	fmt.Println("Test")
	return nil
}

// func Connect(uri string) (*mongo.Client, context.Context,
// 	context.CancelFunc, error) {

// 	// ctx will be used to set deadline for process, here
// 	// deadline will of 30 seconds.
// 	ctx, cancel := context.WithTimeout(context.Background(),
// 		30*time.Second)

// 	// mongo.Connect return mongo.Client method
// 	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
// 	return client, ctx, cancel, err
// }

// // This is a user defined method to close resources.
// // This method closes mongoDB connection and cancel context.
// func Close(client *mongo.Client, ctx context.Context,
// 	cancel context.CancelFunc) {

// 	// CancelFunc to cancel to context
// 	defer cancel()

// 	// client provides a method to close
// 	// a mongoDB connection.
// 	defer func() {

// 		// client.Disconnect method also has deadline.
// 		// returns error if any,
// 		if err := client.Disconnect(ctx); err != nil {
// 			panic(err)
// 		}
// 	}()
// }
// func Ping(client *mongo.Client, ctx context.Context) error {

// 	// mongo.Client has Ping to ping mongoDB, deadline of
// 	// the Ping method will be determined by cxt
// 	// Ping method return error if any occurred, then
// 	// the error can be handled.
// 	if err := client.Ping(ctx, readpref.Primary()); err != nil {
// 		return err
// 	}
// 	fmt.Println("connected successfully")
// 	return nil
// }
