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
	companyIDFromURLPath := strings.Split(req.URL.Path, "/")[1]
	log.Printf("Http reply with current headers and body for company %s", companyIDFromURLPath)
	log.Printf("	Used endpoint: %s", req.URL.Path)
	log.Println("	Returning following: ")

	headers, ok := headersPerCompany[companyIDFromURLPath]
	if ok {
		log.Printf("	Customized headers for company %s:", companyIDFromURLPath)
	} else {
		log.Println("	Default Headers:")
		headers = headersPerCompany["0"]
	}

	for name, value := range headers {
		w.Header().Add(name, value)
		log.Printf("		%s, %s", name, value)
	}

	if code, ok := codePerCompany[companyIDFromURLPath]; ok {
		log.Printf("	Customized StatusCode for company %s: %d", companyIDFromURLPath, code)
		w.WriteHeader(code)
	} else {
		log.Printf("	Default StatusCode returned: %d", codePerCompany["0"])
		w.WriteHeader(codePerCompany["0"])
	}

	if body, ok := bodyPerCompany[companyIDFromURLPath]; ok {
		log.Printf("	Customized body for company %s: %s", companyIDFromURLPath, body)
		w.Write([]byte(body))
	} else {
		log.Printf("	Default body returned: %s", bodyPerCompany["0"])
		w.Write([]byte(bodyPerCompany["0"]))
	}
}

func setHeaders(w http.ResponseWriter, req *http.Request) {
	log.Println("SetHeaders:")
	var newHeadersForCompany = make(map[string]string)
	var companyID = getCompanyIDFromHeader("setHeaders", w, req)
	if companyID == "" {
		return
	}
	log.Printf("	New headers set for companyId: %s", companyID)
	for name, headers := range req.Header {
		for _, h := range headers {
			if !strings.EqualFold("companyid", name) {
				newHeadersForCompany[name] = h
				log.Printf("		%s : %s", name, h)
			}
		}
	}
	headersPerCompany[companyID] = newHeadersForCompany
}

func setBody(w http.ResponseWriter, req *http.Request) {
	log.Println("SetBody:")
	log.Println("	Body:")
	var companyID = getCompanyIDFromHeader("setBody", w, req)
	if companyID == "" {
		return
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

	var companyID = getCompanyIDFromHeader("setStatusCode", w, req)
	if companyID == "" {
		return
	}

	parsedInt, err := strconv.Atoi(req.FormValue("statuscode"))
	if err == nil && (parsedInt >= 100 && parsedInt <= 599) {
		codePerCompany[companyID] = parsedInt
		log.Printf("		StatusCode for company %s : %d", companyID, codePerCompany[companyID])
		return
	}
	codePerCompany[companyID] = 200
	log.Printf("		StatusCode for company %s not set - must be in range 200-500", companyID)
}

func setHeadersAndBody(w http.ResponseWriter, req *http.Request) {
	log.Println("Setting up headers and body")
	var companyID = getCompanyIDFromHeader("setHeadersAndBody", w, req)
	if companyID == "" {
		return
	}
	setBody(w, req)
	setHeaders(w, req)
	log.Printf("Headers and body are set")
}

func setEverything(w http.ResponseWriter, req *http.Request) {
	log.Println("Setting up headers, body and code")
	var companyID = getCompanyIDFromHeader("setEverything", w, req)
	if companyID == "" {
		return
	}
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
	bodyPerCompany["0"] = "default body, use companyId in url as sandboxurl:port/<yourcompanyid>/anything or configure body with setBody, setHeadersAndBody or setEverything endpoint"

	codePerCompany = make(map[string]int) // [company : intCode]
	codePerCompany["0"] = 200
}

func main() {
	port := ":80"
	if len(os.Args) >= 2 {
		port = os.Args[1]
	}
	reinitGlobalVariables()
	log.Println("Server version: 0.07")

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

func getCompanyIDFromHeader(callerEndpoint string, w http.ResponseWriter, req *http.Request) string {
	var companyID = req.Header.Get("companyid")
	if companyID == "" {
		log.Printf("	%s was called without company ID, returning error...", callerEndpoint)
		w.WriteHeader(403)
		w.Write([]byte(`Request was called without companyid header specified. Specify companyid header.`))
	}
	return companyID
}
