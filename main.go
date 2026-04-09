package main

import (
	"fmt"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	server := http.Server{}
	server.Handler = serveMux
	server.Addr = ":8080"
	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	err := server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
	}
}
