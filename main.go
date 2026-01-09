package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", helloHandler)

	log.Println("Listening on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, htmx + Alpine.js tutorial!")
}
