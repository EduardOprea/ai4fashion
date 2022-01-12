package rabbitmqutils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/EduardOprea/ai4fashion/web-api/models"
	"github.com/streadway/amqp"
)

func PublishImageToProcessTransaction(tran models.ProcessImageTran) error {
	ch, err := GetAMQPChannel()
	defer ch.Close()
	if err != nil {
		fmt.Printf("Error opening amqp channel => %+v", err)
		return err
	}

	q, err := ch.QueueDeclare("ImagesToProcess", false, false, false, false, nil)
	if err != nil {
		fmt.Printf("Error when declaring queue => %+v", err)
		return err
	}

	tranJson, err := json.Marshal(&tran)
	if err != nil {
		fmt.Printf("Error JSON convert of process image transaction => %v", err)
	}
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        tranJson,
		},
	)

	if err != nil {
		fmt.Printf("Error publishing to queue => %+v", err)
		return err
	}

	fmt.Println("Succesfully published query image to the Queue")
	return nil
}
func GetAMQPChannel() (*amqp.Channel, error) {
	fmt.Println("Rabbit MQ connect")
	// amqpConn := "amqp://guest:guest@localhost:5672/"
	amqpConn := os.Getenv("AMQP_URL")
	fmt.Printf("Connection string to rabbit mq -> %s \n", amqpConn)
	conn, err := amqp.Dial(amqpConn)
	if err != nil {
		return nil, err
	}
	return conn.Channel()
}
