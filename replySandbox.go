package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

var currentHeaders = make(map[string]string)
var currentBody = ""
var currentStatusCode = 200

func handleRequests(w http.ResponseWriter, req *http.Request) {
	log.Println("Http reply with current headers and body")
	log.Println("	Headers:")
	for name, value := range currentHeaders {
		w.Header().Add(name, value)
		log.Printf("		%s, %s", name, value)
	}

	log.Printf("	StatusCode: %d", currentStatusCode)
	w.WriteHeader(currentStatusCode)

	log.Printf("	Body: %s", currentBody)
	w.Write([]byte(currentBody))
}

func setHeaders(w http.ResponseWriter, req *http.Request) {
	log.Println("SetHeaders:")
	log.Println("	Headers:")
	currentHeaders = make(map[string]string)
	for name, headers := range req.Header {
		for _, h := range headers {
			currentHeaders[name] = h
		}
	}
}

func setBody(w http.ResponseWriter, req *http.Request) {
	log.Println("SetBody:")
	log.Println("	Body:")

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	currentBody = string(bodyBytes)
	log.Printf("		%s", currentBody)
}

func setStatusCode(w http.ResponseWriter, req *http.Request) {
	log.Println("SetStatusCode:")

	parsedInt, err := strconv.Atoi(req.FormValue("statuscode"))
	if err == nil {
		currentStatusCode = parsedInt
	}
	log.Printf("		StatusCode: %d", currentStatusCode)
}

func setHeadersAndBody(w http.ResponseWriter, req *http.Request) {
	log.Println("Setting up headers and body")
	setBody(w, req)
	setHeaders(w, req)
	log.Printf("Headers and body are set")
}

func reflectRequest(w http.ResponseWriter, req *http.Request) {
	log.Println("Request for reflection received")
	for name, headers := range req.Header {
		for _, h := range headers {
			w.Header().Add(name, h)
		}
	}

	parsedInt, atoiErr := strconv.Atoi(req.FormValue("statuscode"))
	if atoiErr == nil {
		w.WriteHeader(parsedInt)
	}

	_, err := io.Copy(w, req.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func clear(w http.ResponseWriter, req *http.Request) {
	currentHeaders = make(map[string]string)
	currentBody = ""
	currentStatusCode = 200
}

func main() {
	port := ":8079"
	if len(os.Args) >= 2 {
		port = os.Args[1]
	}

	log.Println("Server version: 0.04")

	http.HandleFunc("/", handleRequests)
	http.HandleFunc("/setHeaders", setHeaders)
	http.HandleFunc("/setBody", setBody)
	http.HandleFunc("/setStatusCode", setStatusCode)
	http.HandleFunc("/setHeadersAndBody", setHeadersAndBody)
	http.HandleFunc("/reflect", reflectRequest)
	http.HandleFunc("/clear", clear)

	ch := make(chan error)
	go func() {
		ch <- http.ListenAndServe(port, nil)
	}()

	log.Printf("Customizable response server is listening on %s", port)
	log.Fatal(<-ch)
}
