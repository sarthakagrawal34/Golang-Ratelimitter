package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func perClientRateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	type client struct{
		limiter *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients{
				if time.Since(client.lastSeen) > 3 * time.Minute{
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil{
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println("client ip: ", ip)
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2,4)}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow(){
			mu.Unlock()
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, please try again later",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(message)
			return
		}
		mu.Unlock()
		// execute the next function
		next(w, r)
	})
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
	http.Handle("/ping", perClientRateLimiter(endPointHandler))
	fmt.Println("Server running on port: 9080")
	err := http.ListenAndServe(":9080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
