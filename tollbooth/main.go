package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	tollbooth "github.com/didip/tollbooth/v7"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func endPointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)

	message := Message{
		Status: "Successful",
		Body:   "Hi! You've reached the API. How may I help you?",
	}

	err := json.NewEncoder(writer).Encode(message)
	if err != nil {
		log.Println(err)
		return
	}
}

func main() {
	message := Message{
		Status: "Request Failed",
		Body: "The API is at capacity, please try again later",
	}
	jsonMessage, _ := json.Marshal(message)
	tlbthLimiter := tollbooth.NewLimiter(4, nil)
	tlbthLimiter.SetMessageContentType("application/json")
	tlbthLimiter.SetMessage(string(jsonMessage))
	http.Handle("/ping", tollbooth.LimitFuncHandler(tlbthLimiter, endPointHandler))
	fmt.Println("Server running on port: 9080")
	err := http.ListenAndServe(":9080", nil)
	if err != nil {
		log.Fatal(err)
	}
}