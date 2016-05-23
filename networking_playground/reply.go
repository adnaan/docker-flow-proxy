package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var port string
var name string
var branch string

func reply(w http.ResponseWriter, r *http.Request) {

	io.WriteString(w, fmt.Sprintf("I am %v:%v listening on port %v \n", name, branch, port))
}

func main() {

	http.HandleFunc("/", reply)
	if port == "" {
		log.Println("Port empty")
		return
	}
	log.Println("Listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
