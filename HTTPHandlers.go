package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"sync"
)

// Mutex to lock thread
var mutex = &sync.Mutex{}

// RunClientMode starts the client mode
func RunClientMode() {
	router := httprouter.New() // Create HTTP router
	router.GET("/", homepage)  // Services index.html
	router.GET("/getImage", getImage)

	log.Printf("Listening on port 40500\n")
	err := http.ListenAndServe(":8080", router) // ports for serving web page
	if err != nil {
		log.Fatal("Failed to Listen and Serve: ", err)
		return
	}
}

// homepage serves the index.html file
func homepage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Println("Serving homepage")
	http.ServeFile(w, r, "./html/index.html")
	return
}

// getImage serves the image
func getImage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Lock thread
	mutex.Lock()
	defer mutex.Unlock()
	imageUrl := r.URL.Query().Get("url")
	log.Printf("Serving image: %s\n", imageUrl)
	log.Printf("Serving image: %s\n", imageUrl)

	client, err := NewTFTPClient() // instantiate a new TFTP client
	if err != nil {
		log.Printf("Error Creating TFTP Client: %s\n", err)
		return
	}
	defer client.Close()
	err, img, _ := client.RequestFile(imageUrl) // request the file via url
	log.Printf("Image Size: %d\n", len(img))
	if err != nil {
		log.Printf("Error Requesting File over TFTP: %s\n", err)
		return
	}

	// save the image
	err = os.WriteFile("image.mp4", img, 0777) // save the image

	w.Header().Set("Content-Type", "video/mp4") // set the content type
	n, err := w.Write(img)
	log.Printf("Serving image of size: %d\n", n)
	return
}
