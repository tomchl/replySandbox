package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

var headersPerCompany = make(map[string]map[string]map[string]string) // [company : [path : [header : value]]
var bodyPerCompany = make(map[string]map[string]string)               // [company : [path : body]]
var codePerCompany = make(map[string]map[string]int)                  // [company : [path : intCode]]

func handleRequests(w http.ResponseWriter, req *http.Request) {
	companyIDFromURLPath := req.Header.Get("companyId")
	path := req.RequestURI
	log.Printf("Http reply with current headers and body for company %s", companyIDFromURLPath)
	log.Printf("	Used endpoint: %s", path)
	log.Println("	Returning following: ")

	headers, ok := headersPerCompany[companyIDFromURLPath][path]
	if ok {
		log.Printf("	Customized headers for company %s:", companyIDFromURLPath)
	} else {
		log.Println("	Default Headers:")
		headers = headersPerCompany["0"]["/"]
	}

	for name, value := range headers {
		w.Header().Add(name, value)
		log.Printf("		%s, %s", name, value)
	}

	if code, ok := codePerCompany[companyIDFromURLPath][path]; ok {
		log.Printf("	Customized StatusCode for company %s: %d", companyIDFromURLPath, code)
		w.WriteHeader(code)
	} else {
		log.Printf("	Default StatusCode returned: %d", codePerCompany["0"]["/"])
		w.WriteHeader(codePerCompany["0"]["/"])
	}

	if body, ok := bodyPerCompany[companyIDFromURLPath][path]; ok {
		log.Printf("	Customized body for company %s: %s", companyIDFromURLPath, body)
		w.Write([]byte(body))
	} else {
		log.Printf("	Default body returned: %s", bodyPerCompany["0"]["/"])
		w.Write([]byte(bodyPerCompany["0"]["/"]))
	}
}

func setStatusCode(w http.ResponseWriter, req *http.Request) {
	companyID, jsonResult, e := getCompanyIdAndBody(w, req)
	if e != nil {
		log.Fatal(e)
		return
	}

	value, ok := jsonResult["statusCode"]
	if res, okk := value.(map[string]interface{}); ok && okk {
		for key, val := range res {

			if r, err := val.(float64); err {
				if m, _ := codePerCompany[companyID]; m == nil {
					codePerCompany[companyID] = make(map[string]int)
				}
				codePerCompany[companyID][key] = int(r)
			}
		}

	}
}

func setHeaders(w http.ResponseWriter, req *http.Request) {
	companyID, jsonResult, e := getCompanyIdAndBody(w, req)
	if e != nil {
		log.Fatal(e)
		return
	}

	value, ok := jsonResult["headers"]
	if res, okk := value.(map[string]interface{}); ok && okk {
		for key, val := range res {

			if r, err := val.(map[string]interface{}); err {
				if m, _ := headersPerCompany[companyID]; m == nil {
					headersPerCompany[companyID] = make(map[string]map[string]string)
				}
				if m, _ := headersPerCompany[companyID][key]; m == nil {
					headersPerCompany[companyID][key] = make(map[string]string)
				}
				for headerKey, headerVal := range r {
					if endR, err := headerVal.(string); err {
						headersPerCompany[companyID][key][headerKey] = endR
					}
				}
			}
		}

	}
}

func setBody(w http.ResponseWriter, req *http.Request) {
	companyID, jsonResult, e := getCompanyIdAndBody(w, req)
	if e != nil {
		log.Fatal(e)
		return
	}

	value, ok := jsonResult["body"]
	if res, okk := value.(map[string]interface{}); ok && okk {
		for key, val := range res {
			if r, err := json.Marshal(val); err == nil {
				if m, _ := bodyPerCompany[companyID]; m == nil {
					bodyPerCompany[companyID] = make(map[string]string)
				}
				bodyPerCompany[companyID][key] = string(r)
			}
		}
	}
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

	setHeaders(w, req)
	setBody(w, req)
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
		headersPerCompany[companyID] = make(map[string]map[string]string)
		bodyPerCompany[companyID]["/"] = ""
		codePerCompany[companyID]["/"] = 200
	}
}

func reinitGlobalVariables() {
	headersPerCompany = make(map[string]map[string]map[string]string) // [company : [header : value]]
	headersPerCompany["0"] = make(map[string]map[string]string)

	bodyPerCompany = make(map[string]map[string]string) // [company : body]
	bodyPerCompany["0"] = make(map[string]string)
	bodyPerCompany["0"]["/"] = "default body, use companyId in url as sandboxurl:port/<yourcompanyid>/anything or configure body with setBody, setHeadersAndBody or setEverything endpoint"

	codePerCompany = make(map[string]map[string]int) // [company : intCode]
	codePerCompany["0"] = make(map[string]int)
	codePerCompany["0"]["/"] = 200
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

func getCompanyIdAndBody(w http.ResponseWriter, req *http.Request) (string, map[string]interface{}, error) {
	var companyID = getCompanyIDFromHeader("setEverything", w, req)
	if companyID == "" {
		return "", nil, errors.New("Couldnt parse companyId")
	}

	var jsonResult map[string]interface{}
	body, e := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	if e != nil {
		return "", nil, e
	}
	json.Unmarshal(body, &jsonResult)

	return companyID, jsonResult, nil
}