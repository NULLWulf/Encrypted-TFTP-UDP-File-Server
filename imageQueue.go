package main

import (
	"container/list"
	"errors"
	"io"
	"log"
	"net/http"
)

// ImageQueueObj is a queue of images
type ImageQueueObj struct {
	Queue *list.List
}

// NewImageQueueObj creates a new ImageQueueObj
func NewImageQueueObj() *ImageQueueObj {
	return &ImageQueueObj{
		Queue: list.New(),
	}
}

// AddNewAndReturnImg AddImage adds a new image to the queue, returns the error if one is found during the request
func (iq *ImageQueueObj) AddNewAndReturnImg(url string) (err error, img []byte) {
	img, err = iq.requestImage(url)
	// Immediately return error if found, if error is found
	// image is not added to queue
	if err != nil {
		log.Printf("Error Making Request for Image: %s\n", err)
		err = errors.New("problem requesting image")
		return
	}
	// If image is nil, image assumed to be not found error flag is thrown
	if img == nil {
		log.Printf("Image body is empty.\n")
		err = errors.New("image body is empty")
		return
	}

	// Stores no more than 10 images at a time
	if iq.Queue.Len() < 10 {
		iq.Queue.PushBack(img)
	} else {
		iq.Queue.Remove(iq.Queue.Front())
		iq.Queue.PushBack(img)
	}

	// returns image to be handled by caller
	return
}

// RequestImage makes a GET request to the specified URL and returns the image data
func (iq *ImageQueueObj) requestImage(url string) (img []byte, err error) {
	// Make a GET request to the image URL
	var resp *http.Response
	resp, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) error {
		err := Body.Close()
		if err != nil {
			return err
		}
		return err
	}(resp.Body)

	// Read the image data into a byte slice
	img, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func ProxyRequest(url string) (file []byte, err error) {
	// Make a GET request to the image URL
	var resp *http.Response
	resp, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) error {
		err := Body.Close()
		if err != nil {
			return err
		}
		return err
	}(resp.Body)

	// Read the image data into a byte slice
	file, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}
