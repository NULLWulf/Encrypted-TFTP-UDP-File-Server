package main

import (
	"CSC445_Assignment2/tftp"
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
	log.Printf("Image Size: %d\n", len(img))

	dps, _ := tftp.PrepareTFTPDataPackets(img, 512)
	newImgr := make([]byte, 0)
	for _, dp := range dps {
		newImgr = append(newImgr, dp.Data...)
	}
	log.Printf("New Image Size: %d\n", len(newImgr))
	_, err = w.Write(newImgr)
	if err != nil {
		return
	}
}
