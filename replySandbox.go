package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func handleRequests(w http.ResponseWriter, req *http.Request) {
	log.Println("Received:")
	log.Println("	Headers:")
	for name, headers := range req.Header {
		for _, h := range headers {
			w.Header().Add(name, h)
			log.Printf("		%s, %s", name, h)
		}
	}

	_, err := io.Copy(w, req.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	port := ":8079"
	if len(os.Args) >= 2 {
		port = os.Args[1]
	}

	log.Println("Server version: 0.03")

	http.HandleFunc("/", handleRequests)

	ch := make(chan error)
	go func() {
		ch <- http.ListenAndServe(port, nil)
	}()

	log.Printf("SaaS server is listening on %s", port)
	log.Fatal(<-ch)
}
