/* This is to run the server only - without Raspberry PI*/
package main

import (
	"fmt"
	"log"
	"net/http"
)


func startServer() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	//http.HandleFunc("/targetdevice/attack", initAttack)

	fmt.Printf("| Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	startServer()
}
