package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/EduardOprea/ai4fashion/processer/dbutils"
	"github.com/EduardOprea/ai4fashion/processer/models"
	"github.com/streadway/amqp"
)

const torchServeApiUrlDefault = "http://localhost:8080"
const webApiDefaultUrl = "http://localhost:3031"
const amqpDefaultUrl = "amqp://guest:guest@localhost:5672/"

func processImageTransactionReceived(received []byte) {
	var tran models.ProcessImageTran
	if err := json.Unmarshal(received, &tran); err != nil {
		fmt.Println("Error deserializing bytes received to process image transaction")
		return
	}

	fmt.Println("Deserialised bytes received succesfully")
	fmt.Printf("Result => %v", tran)
	data, err := getImageToProcess(tran.ImageName)
	if err != nil {
		fmt.Printf("Downloading image to process from web api failed => %v", err)
		return
	}
	fmt.Printf("Downloaded the image succesfully from web-api size %v \n", len(data))
	// ioutil.WriteFile("originalimagetest.jpg", data, 0600)
	dataProcesssed, err := editImageTest(data, tran.DesiredAttributes)
	if err != nil {
		fmt.Printf("Error when editing image with torchserve => %v\n", err)
		return
	}
	fmt.Println("Succesfully get fashion-serve result")
	fmt.Printf("The size of the data received is %v\n", len(dataProcesssed))

	// ioutil.WriteFile("response_torchserve.txt", dataProcesssed, 0600)

	var rawImage models.RawImage
	json.Unmarshal(dataProcesssed, &rawImage)
	// saveRawImageAsJpeg(rawImage)
	if len(rawImage.Data) < 1 {
		fmt.Println("editing the image failed")
		return
	}
	var jpegImageData bytes.Buffer

	var opts jpeg.Options
	opts.Quality = 100
	imageAsArray := flattenMatrix(rawImage.Data)

	imageTest := image.NewRGBA(image.Rect(0, 0, 128, 128))
	imageTest.Pix = imageAsArray
	err = jpeg.Encode(&jpegImageData, imageTest, &opts)

	//ioutil.WriteFile("finalprocessedimgtest.jpeg", jpegImageData.Bytes(), 0600)

	dbutils.UploadFile(jpegImageData.Bytes(), tran.ImageName)
	return
}

func saveRawImageAsJpeg(rawImage models.RawImage) {
	imageAsArray := flattenMatrix(rawImage.Data)

	imageTest := image.NewRGBA(image.Rect(0, 0, 128, 128))
	imageTest.Pix = imageAsArray

	out, _ := os.Create("./imgTest.jpeg")
	defer out.Close()

	var opts jpeg.Options
	opts.Quality = 100

	err := jpeg.Encode(out, imageTest, &opts)
	if err != nil {
		fmt.Println("error saving raw bytes processed as jpeg image")
	}
}
func flattenMatrix(matrix [][][]byte) []byte {
	height := len(matrix)
	width := len(matrix[0])
	// nChannels := len(matrix[0][0])
	nChannels := 4
	size := width * height * nChannels
	count := 0
	array := make([]byte, size)
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			for k := 0; k < nChannels-1; k++ {
				// array[i*width+j*nChannels+k] = matrix[i][j][k]
				array[count] = matrix[i][j][k]
				count++

			}
			array[count] = 1
			count++

		}
	}
	return array
}
func saveBytesAsImageTest(imgByte []byte) {

	img, _, err := image.Decode(bytes.NewReader(imgByte))
	if err != nil {
		log.Fatalln(err)
	}

	out, _ := os.Create("./img.jpeg")
	defer out.Close()

	var opts jpeg.Options
	opts.Quality = 1

	err = jpeg.Encode(out, img, &opts)
	//jpeg.Encode(out, img, nil)
	if err != nil {
		log.Println(err)
	}

}
func editImageTest(image []byte, desiredAttr string) ([]byte, error) {
	httpClient := http.Client{Timeout: time.Duration(60) * time.Second}
	r := bytes.NewReader(image)
	var torchServeApiUrl string
	if len(os.Getenv("TORCHSERVE_URL")) > 0 {
		torchServeApiUrl = os.Getenv("TORCHSERVE_URL")
	} else {
		torchServeApiUrl = torchServeApiUrlDefault
	}
	fmt.Printf("Using the following url for torch serve => %s \n ", torchServeApiUrl)

	var modelEndpointUrl string

	switch desiredAttr {
	case "add-floral":
		modelEndpointUrl = fmt.Sprintf("%s/predictions/%s", torchServeApiUrl, "cycleganfloraladd")
		break
	case "remove-floral":
		modelEndpointUrl = fmt.Sprintf("%s/predictions/%s", torchServeApiUrl, "cycleganfloralremove")
		break
	case "add-stripes":
		modelEndpointUrl = fmt.Sprintf("%s/predictions/%s", torchServeApiUrl, "cycleganstripesadd")
		break
	case "remove-stripes":
		modelEndpointUrl = fmt.Sprintf("%s/predictions/%s", torchServeApiUrl, "cycleganstripesremove")
		break
	}

	fmt.Printf("Calling the following endpoint => %v\n", modelEndpointUrl)
	resp, err := httpClient.Post(modelEndpointUrl, "binary", r)

	// resp, err := httpClient.Post(modelEndpointUrl, "binary", r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	fmt.Println("Succes sending req to torch serve")
	fmt.Printf("Response status code => %v \n", resp.StatusCode)
	imageProcessed, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return imageProcessed, nil
}
func getImageToProcess(fileName string) ([]byte, error) {
	c := http.Client{Timeout: time.Duration(60) * time.Second}
	var webApiUrl string
	if len(os.Getenv("API_URL")) > 0 {
		webApiUrl = os.Getenv("API_URL")
	} else {
		webApiUrl = webApiDefaultUrl
	}
	fmt.Printf("Using the following url for web-api => %s\n", webApiUrl)
	resp, err := c.Get(fmt.Sprintf("%s/localImage/%s", webApiUrl, fileName))
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
	var amqpUrl string
	if len(os.Getenv("AMQP_URL")) > 0 {
		amqpUrl = os.Getenv("AMQP_URL")
	} else {
		amqpUrl = amqpDefaultUrl
	}
	fmt.Printf("Using the following url for rabbitmq -> %s \n", amqpUrl)
	fmt.Println("Rabbit MQ consumer start")
	conn, err := amqp.Dial(amqpUrl)
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
