package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Set the flags for the logging package to give us the filename in the logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("starting server...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `Hello, visitor!`)
	})
	log.Fatal(http.ListenAndServe(":8000", nil))
}
