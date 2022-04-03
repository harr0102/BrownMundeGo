/* This is to run the server only - without Raspberry PI*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type TargetDevice struct {
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
}

func initAttack(w http.ResponseWriter, r *http.Request) {
	// Declare a new Person struct.
	var td TargetDevice
	fmt.Printf("lol")
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&td)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Do something with the TargetDevice struct...
	fmt.Fprintf(w, "TargetDevice: %+v", td)
}

func startServer() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/targetdevice/attack", initAttack)

	fmt.Printf("| Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	startServer()
}
