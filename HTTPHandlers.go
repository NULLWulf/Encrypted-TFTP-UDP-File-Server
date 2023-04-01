package main

import (
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var mutex = &sync.Mutex{}

func RunClientMode() {
	router := httprouter.New() // Create HTTP router
	router.GET("/", homepage)  // Services index.html
	router.GET("/getImage", getImage)

	log.Printf("Listening on port 8080\n")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Failed to Listen and Serve: ", err)
		return
	}
}

// t
func homepage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Println("Serving homepage")
	http.ServeFile(w, r, "./html/index.html")
	return
}

func getImage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Lock thread
	mutex.Lock()
	defer mutex.Unlock()
	imageUrl := r.URL.Query().Get("url")
	log.Printf("Serving image: %s\n", imageUrl)

	client, err := NewTFTPClient()
	if err != nil {
		log.Printf("Error Creating TFTP Client: %s\n", err)
		return
	}
	defer client.Close()
	err, img, _ := client.RequestFile(imageUrl)
	log.Printf("Image Size: %d\n", len(img))
	if err != nil {
		log.Printf("Error Requesting File over TFTP: %s\n", err)
		return
	}
	// save the image
	err = ioutil.WriteFile("test.jpg", img, 7777)

	w.Header().Set("Content-Type", "image/jpeg")
	_, err = w.Write(img)
	return
}
