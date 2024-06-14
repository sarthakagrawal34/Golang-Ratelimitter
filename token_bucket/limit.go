package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/time/rate"
)

func rateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := rate.NewLimiter(2, 6)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, please try again later",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			err := json.NewEncoder(w).Encode(message)
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			next(w, r)
		}
	})
}
