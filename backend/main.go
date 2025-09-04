package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from Go backend!"))
	})

	log.Println("Starting server on :6080")
	log.Fatal(http.ListenAndServe(":6080", nil))
}
