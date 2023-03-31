package main

import (
	"CSC445_Assignment2/tftp"
	"log"
	"os"
)

func TestRetrievingFile() {
	ImgQue := NewImageQueueObj()
	err, img := ImgQue.AddNewAndReturnImg("https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png")
	if err != nil {
		log.Fatal(err)
	}
	DQS, err := tftp.PrepareData(img, 512)
	if err != nil {
		log.Fatal(err)
	}
	RD := tftp.RebuildData(DQS)
	// save image to disc
	err = os.WriteFile("test.png", RD, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
