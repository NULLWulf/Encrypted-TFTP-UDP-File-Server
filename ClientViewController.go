package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func RunClientMode() {
	router := httprouter.New() // Create HTTP router
	router.GET("/", homepage)  // Services index.html
	router.GET("/getImage", getImage)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		return
	}
}

func homepage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Println("Serving homepage")
	http.ServeFile(w, r, "./html/index.html")
}

func getImage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	imageUrl := r.URL.Query().Get("url")
	log.Printf("Serving image: %s\n", imageUrl)
	w.Header().Set("Content-Type", "image/jpeg")
	imgQue := NewImageQueueObj()
	err, img := imgQue.AddNewAndReturnImg(imageUrl)
	_, err = w.Write(img)
	if err != nil {
		return
	}
}
