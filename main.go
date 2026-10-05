package main

import (
	"log"
	"net/http"
)

func main() {
	c, err := NewInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/payments/followup", followupHandler(c))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
