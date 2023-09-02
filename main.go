package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

const PORT = "8080"

type Images struct {
	Count int      `json:"counts"`
	Data  []string `json:"data"`
}

func main() {
	fmt.Println("Image Management RestAPI ")
	handleRequests()

}

func handleRequests() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ping", Ping).Methods("GET")
	router.HandleFunc("/showall", ShowAllImages).Methods("GET")
	router.HandleFunc("/upload", UploadImages).Methods("POST")
	log.Printf("Server is running on http://localhost:%s", PORT)
	log.Println(http.ListenAndServe(":"+PORT, router))

}

func Ping(w http.ResponseWriter, r *http.Request) {

	answer := map[string]interface{}{
		"messageType": "S",
		"message":     "",
		"data":        "ResponsePING",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(answer)
}

func ShowAllImages(w http.ResponseWriter, r *http.Request) {
	imagedir := "./uploads"
	images, err := os.Open(imagedir) //open the directory to read all images in the directory
	if err != nil {
		fmt.Println("Error opening directory:", err)
		return
	}

	defer images.Close() //Close the opened directory

	imageinfo, err := images.Readdir(-1) //read image files form directory
	if err != nil {
		fmt.Println("Error reading image file in directory:", err)
		return
	}

	for _, imageinfo := range imageinfo {
		fmt.Println(imageinfo.Name()) //print the files from directory
	}

}

func UploadImages(w http.ResponseWriter, r *http.Request) {

	if errorMSg := r.ParseMultipartForm(32 << 20); errorMSg != nil {
		http.Error(w, errorMSg.Error(), http.StatusInternalServerError)
		return
	}

	//Fetch reference to the fileHeaders post ParseMultiPartForm
	files := r.MultipartForm.File["file"]

	var errNewMSg string
	var httpStatus int

	for _, fileHeader := range files {
		//Opening the file
		file, errorMsg := fileHeader.Open()
		if errorMsg != nil {
			errNewMSg = errorMsg.Error()
			httpStatus = http.StatusInternalServerError
			break
		}

		defer file.Close()

		buffer := make([]byte, 512)
		_, errorMsg = file.Read(buffer)
		if errorMsg != nil {
			errNewMSg = errorMsg.Error()
			httpStatus = http.StatusInternalServerError
			break
		}

		fileType := http.DetectContentType(buffer)
		if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/jpg" {
			errNewMSg = "File type not supported."
			httpStatus = http.StatusBadRequest
			break
		}

		_, errorMsg = file.Seek(0, io.SeekStart)
		if errorMsg != nil {
			errNewMSg = errorMsg.Error()
			httpStatus = http.StatusInternalServerError
			break
		}

		//Directory Creation
		errorMsg = os.MkdirAll("./uploads", os.ModePerm)
		if errorMsg != nil {
			errNewMSg = errorMsg.Error()
			httpStatus = http.StatusInternalServerError
			break
		}

		f, errorMsg := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		if errorMsg != nil {
			errNewMSg = errorMsg.Error()
			httpStatus = http.StatusBadRequest
			break
		}

		defer f.Close()

		_, errorMsg = io.Copy(f, file)
		if errorMsg != nil {
			errNewMSg = errorMsg.Error()
			httpStatus = http.StatusBadRequest
			break
		}
	}

	message := "Image file Uploaded Sucessfully!"
	messageType := "S"

	if errNewMSg != "" {
		message = errNewMSg
		messageType = "E"
	}

	if httpStatus == 0 {
		httpStatus = http.StatusOK
	}

	response := map[string]interface{}{
		"messageType": messageType,
		"message":     message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)

}
