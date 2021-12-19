package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"models"
	"net/http"
	"rabbitmqutils"
	"strconv"

	"github.com/gorilla/mux"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint hit")
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	tempFile, err := ioutil.TempFile("imgs-process", "upload-*.png")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

	processImageTran := models.ProcessImageTran{DesiredAttributes: desiredAttributes, ImageName: tempFile.Name()}
	if err := rabbitmqutils.PublishImageToProcessTransaction(processImageTran); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File as %s\n", tempFile.Name())
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

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/upload", uploadImage).Methods("POST")
	myRouter.HandleFunc("/download/{imageName}", downloadImage)
	log.Fatal(http.ListenAndServe(":8081", myRouter))
}
func main() {
	fmt.Println("Server started")
	handleRequests()
}
