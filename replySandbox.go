package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var headersPerCompany = make(map[string]map[string]string) // [company : [header : value]]
var bodyPerCompany = make(map[string]string)               // [company : body]
var codePerCompany = make(map[string]int)                  // [company : intCode]

func handleRequests(w http.ResponseWriter, req *http.Request) {
	log.Println("Http reply with current headers and body")

	for company, headers := range headersPerCompany {
		if strings.HasPrefix(req.URL.Path, "/"+company) {
			log.Printf("Customized response for company %s:", company)
			log.Println("	Headers:")
			for name, value := range headers {
				//w.Header().Add(name, value)
				log.Printf("		%s, %s", name, value)
			}

			log.Printf("	StatusCode: %d", codePerCompany[company])
			w.WriteHeader(codePerCompany[company])

			log.Printf("	Body: %s", bodyPerCompany[company])
			w.Write([]byte(bodyPerCompany[company]))
			return
		}
	}

	// if company setting is not present, return default values present for company "0"
	for name, value := range headersPerCompany["0"] {
		//w.Header().Add(name, value)
		log.Printf("		%s, %s", name, value)
	}

	log.Printf("	StatusCode: %d", codePerCompany["0"])
	w.WriteHeader(codePerCompany["0"])

	log.Printf("	Body: %s", bodyPerCompany["0"])
	w.Write([]byte(bodyPerCompany["0"]))
}

func setHeaders(w http.ResponseWriter, req *http.Request) {
	log.Println("SetHeaders:")
	log.Println("	Headers:")
	var newHeadersForCompany = make(map[string]string)
	var companyID = "0"
	for name, headers := range req.Header {
		for _, h := range headers {
			if strings.EqualFold("companyid", name) {
				companyID = h
			} else {
				newHeadersForCompany[name] = h
				log.Printf("	%s : %s", name, h)
			}
		}
	}
	log.Printf("	headers set for companyId: %s", companyID)
	headersPerCompany[companyID] = newHeadersForCompany
}

func setBody(w http.ResponseWriter, req *http.Request) {
	log.Println("SetBody:")
	log.Println("	Body:")
	var companyID = req.Header.Get("companyid")
	if companyID == "" {
		companyID = "0"
	}
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	bodyPerCompany[companyID] = string(bodyBytes)
	log.Printf("	body set for company %s : %s", companyID, bodyPerCompany[companyID])
}

func setStatusCode(w http.ResponseWriter, req *http.Request) {
	log.Println("SetStatusCode:")

	var companyID = req.Header.Get("companyid")
	if companyID == "" {
		companyID = "0"
	}

	parsedInt, err := strconv.Atoi(req.FormValue("statuscode"))
	if err == nil && (parsedInt > 200 && parsedInt < 500) {
		codePerCompany[companyID] = parsedInt
		log.Printf("		StatusCode for company %s : %d", companyID, codePerCompany[companyID])
		return
	}
	log.Printf("		StatusCode for company %s not set - must be in range 200-500", companyID)
}

func setHeadersAndBody(w http.ResponseWriter, req *http.Request) {
	log.Println("Setting up headers and body")
	setBody(w, req)
	setHeaders(w, req)
	log.Printf("Headers and body are set")
}

func setEverything(w http.ResponseWriter, req *http.Request) {
	log.Println("Setting up headers, body and code")
	setBody(w, req)
	setHeaders(w, req)
	setStatusCode(w, req)
	log.Printf("Headers, body and code are set")
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
	var companyID = req.Header.Get("companyid")
	if companyID == "" {
		reinitGlobalVariables()
	} else {
		headersPerCompany[companyID] = make(map[string]string)
		bodyPerCompany[companyID] = ""
		codePerCompany[companyID] = 200
	}
}

func reinitGlobalVariables() {
	headersPerCompany = make(map[string]map[string]string) // [company : [header : value]]
	headersPerCompany["0"] = make(map[string]string)

	bodyPerCompany = make(map[string]string) // [company : body]
	bodyPerCompany["0"] = ""

	codePerCompany = make(map[string]int) // [company : intCode]
	codePerCompany["0"] = 200
}

func main() {
	port := ":80"
	if len(os.Args) >= 2 {
		port = os.Args[1]
	}
	reinitGlobalVariables()
	log.Println("Server version: 0.04")

	http.HandleFunc("/", handleRequests)
	http.HandleFunc("/setHeaders", setHeaders)
	http.HandleFunc("/setBody", setBody)
	http.HandleFunc("/setStatusCode", setStatusCode)
	http.HandleFunc("/setHeadersAndBody", setHeadersAndBody)
	http.HandleFunc("/setEverything", setEverything)
	http.HandleFunc("/reflect", reflectRequest)
	http.HandleFunc("/clear", clear)

	http.HandleFunc("/bob/api/v1/service/instance/list", instanceList)
	http.HandleFunc("/bob/api/v1/service/definition/listVisible", listVisible)

	ch := make(chan error)
	go func() {
		ch <- http.ListenAndServe(port, nil)
	}()

	log.Printf("Customizable response server is listening on %s", port)
	log.Fatal(<-ch)
}

// static responses - same for each company
func instanceList(w http.ResponseWriter, req *http.Request) {
	log.Println("Received request on /api/v1/service/instance/list")

	body := `{"Services": [{
				"Id": "1",
				"CompanyId": "1",
				"Name": "DummyName",
				"Definition": {
					"Name": "DummyName",
					"Version": "1",
					"Description": "BlahBlah",
					"CurrentState": "Running",
					"DesiredState": "Running"
				},
				"User": "Dummy",
				"Details": {
					"Endpoints": []
					}
				}]}`
	log.Printf("	Services returned: %s", body)
	w.Write([]byte(body))
}

func listVisible(w http.ResponseWriter, req *http.Request) {
	log.Println("Received request on /api/v1/service/definition/listVisible")

	body := `{"ServicesDefinitions": [{
				"Name": "DummyName",
				"Version": "1",
				"Description": "BlahBlah",
				"CurrentState": "Running",
				"DesiredState": "Running"
			}]}`
	log.Printf("	Services returned: %s", body)
	w.Write([]byte(body))
}
