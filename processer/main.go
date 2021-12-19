package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/EduardOprea/ai4fashion/processer/models"

	"github.com/streadway/amqp"
)

const connString = "mlongodb://localhost:27017"

func processImageTransactionReceived(received []byte) {
	var tran models.ProcessImageTran
	if err := json.Unmarshal(received, &tran); err != nil {
		fmt.Println("Error deserializing bytes received to process image transaction")
		return
	}

	fmt.Println("Deserialised bytes received succesfully")
	fmt.Printf("Result => %v", tran)
	downloadImage(tran.ImageName)

}

func downloadImage(fileName string) ([]byte, error) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	resp, err := c.Get(fmt.Sprintf("http://localhost:8081/download/%s", fileName))
	if err != nil {
		fmt.Printf("Error %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("Body : %s", body)
	return nil, nil
}
func main() {
	fmt.Println("Rabbit MQ consumer start")
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Channel ok")
	msgs, err := ch.Consume(
		"ImagesToProcess",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			fmt.Println("Message received")
			go processImageTransactionReceived(d.Body)
		}
	}()

	fmt.Println("Successfully Connected to our RabbitMQ Instance")
	fmt.Println(" [*] - Waiting for messages")
	<-forever
}
