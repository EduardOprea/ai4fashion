package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EduardOprea/ai4fashion/web-api/dbutils"
	"github.com/EduardOprea/ai4fashion/web-api/models"
	"github.com/EduardOprea/ai4fashion/web-api/rabbitmqutils"

	"github.com/gorilla/mux"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint hit")
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
func uploadImage(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	fmt.Println("File Upload Endpoint Hit")
	//r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("image")
	desiredAttributes := r.FormValue("desiredAttributes")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fmt.Printf("Desired attirbutes: %s \n", desiredAttributes)
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	imgExtension := strings.Split(handler.Filename, ".")[1]
	fmt.Printf("File extension: %+v\n", imgExtension)

	tempFile, err := ioutil.TempFile("imgs-process", fmt.Sprintf("upload-*.%s", imgExtension))
	fmt.Printf("Saving uploaded file as %s \n", tempFile.Name())
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	fmt.Printf("The size of the image in bytes is => %v\n", len(fileBytes))
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

	processImageTran := models.ProcessImageTran{DesiredAttributes: desiredAttributes, ImageName: filepath.Base(tempFile.Name())}
	if err := rabbitmqutils.PublishImageToProcessTransaction(processImageTran); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bodyResponse := models.ProcessImageResponse{
		Message:   "Succes => processing",
		ImageName: filepath.Base(tempFile.Name()),
	}
	bodyJson, _ := json.Marshal(bodyResponse)
	fmt.Fprint(w, string(bodyJson))
}
func downloadImage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Download image endpoint hit")
	vars := mux.Vars(r)
	imageName := vars["imageName"]
	fmt.Printf("Returning image %s \n", imageName)
	// TODO check if image exists first
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(imageName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, "imgs-process/"+imageName)
	// after serving file perhaps delete it or make a separate service to do that
	// in case it can not be deleted imediately after being served
}
func getProcesedImage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get processed image endpoint hit")
	vars := mux.Vars(r)
	imageName := vars["imageName"]
	fmt.Printf("Returning image from db %s \n", imageName)
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(imageName))
	w.Header().Set("Content-Type", "application/octet-stream")

	imageData := dbutils.GetImageProcessed(imageName)
	if imageData == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
func handleRequests(listenPort string) {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/upload", uploadImage).Methods("POST")
	myRouter.HandleFunc("/localImage/{imageName}", downloadImage)
	myRouter.HandleFunc("/processedImage/{imageName}", getProcesedImage)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", listenPort), myRouter))
}
func main() {
	listenPort := "8081"
	if len(os.Getenv("API_LISTEN_PORT")) > 0 {
		listenPort = os.Getenv("API_LISTEN_PORT")
	}
	fmt.Printf("Server started, listening on port %s\n", listenPort)
	handleRequests(listenPort)
}
