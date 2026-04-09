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
	serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("/healthz", func(resW http.ResponseWriter, req *http.Request) {
		h := resW.Header()
		h["Content-Type"] = []string {"text/plain; charset=utf-8"}
		resW.WriteHeader(200)
		_, _ = resW.Write([]byte("OK"))
	})
	err := server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
	}
}
