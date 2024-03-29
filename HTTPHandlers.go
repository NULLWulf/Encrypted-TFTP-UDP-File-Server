package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Mutex to lock thread

// RunClientMode starts the client mode
func RunClientMode() {
	router := httprouter.New() // Create HTTP router
	router.GET("/", homepage)  // Services index.html
	router.GET("/getImage", getImage2)

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
}

func getImage2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Lock thread
	imageUrl := r.URL.Query().Get("url")
	log.Printf("Serving image: %s\n", imageUrl)

	client, err := NewTFTPClient() // instantiate a new TFTP client
	if err != nil {
		log.Printf("Error Creating TFTP Client: %s\n", err)
		return
	}
	defer client.Close()

	var img []byte
	func() {
		img, _, err = client.RequestFile(imageUrl) // request the file via url
		if err != nil {
			log.Printf("Error Requesting File over TFTP: %s\n", err)
		}
	}()

	w.Header().Set("Content-Type", "image/jpeg") // set the content type
	w.Write(img)

}
