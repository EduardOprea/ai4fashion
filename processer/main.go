package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/EduardOprea/ai4fashion/processer/dbutils"
	"github.com/EduardOprea/ai4fashion/processer/models"
	"github.com/streadway/amqp"
)

func processImageTransactionReceived(received []byte) {
	var tran models.ProcessImageTran
	if err := json.Unmarshal(received, &tran); err != nil {
		fmt.Println("Error deserializing bytes received to process image transaction")
		return
	}

	fmt.Println("Deserialised bytes received succesfully")
	fmt.Printf("Result => %v", tran)
	// TODO => do not save the image to filesystem, instead I need to process it and
	// and then store it to the db
	// after storing it to the db, we need to notify the web api that is ready in order to further notify
	// the client, or the client will make continuous polling based on a transaction id that was provided to him
	// and the name under which the image is saved in the db should be correlated with the transaction id
	data, err := getImageToProcess(tran.ImageName)
	if err != nil {
		fmt.Printf("Downloading image to process from web api failed => %v", err)
		return
	}
	fmt.Println("Downloaded the image succesfully, uploading to DB")
	dbutils.UploadFile(data, tran.ImageName)
}

func getImageToProcess(fileName string) ([]byte, error) {
	c := http.Client{Timeout: time.Duration(60) * time.Second}
	resp, err := c.Get(fmt.Sprintf("http://localhost:8081/download/%s", fileName))
	if err != nil {
		// fmt.Printf("Error %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.ContentLength <= 0 {
		log.Println("[*] Destination server does not support breakpoint download.")
	}
	raw := resp.Body
	reader := bufio.NewReaderSize(raw, 1024*32)

	// file, err := os.Create("to-process/" + fileName)
	// defer file.Close()
	// if err != nil {
	// 	panic(err)
	// }
	// writer := bufio.NewWriter(file)
	buff := make([]byte, 0)
	buffTemp := make([]byte, 32*1024)
	written := 0
	go func() {
		for {
			nr, er := reader.Read(buffTemp)
			if nr > 0 {
				buff = append(buff, buffTemp...)
				written += nr
				// nw, ew := writer.Write(buffTemp[0:nr])
				// if nw > 0 {
				// 	written += nw
				// }
				// if ew != nil {
				// 	err = ew
				// 	break
				// }
				// if nr != nw {
				// 	err = io.ErrShortWrite
				// 	break
				// }
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break
			}
		}
		if err != nil {
			panic(err)
		}
	}()

	spaceTime := time.Second * 1
	ticker := time.NewTicker(spaceTime)
	lastWtn := 0
	stop := false

	for {
		select {
		case <-ticker.C:
			speed := written - lastWtn
			fmt.Printf("[*] Speed %s / %s \n", bytesToSize(speed), spaceTime.String())
			if written-lastWtn == 0 {
				ticker.Stop()
				stop = true
				break
			}
			lastWtn = written
		}
		if stop {
			break
		}
	}

	return buff, nil
}

func bytesToSize(length int) string {
	var k = 1024 // or 1024
	var sizes = []string{"Bytes", "KB", "MB", "GB", "TB"}
	if length == 0 {
		return "0 Bytes"
	}
	i := math.Floor(math.Log(float64(length)) / math.Log(float64(k)))
	r := float64(length) / math.Pow(float64(k), i)
	return strconv.FormatFloat(r, 'f', 3, 64) + " " + sizes[int(i)]
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
