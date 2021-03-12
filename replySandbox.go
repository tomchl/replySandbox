package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"
)

var headersPerCompany = make(map[string]map[string]map[string]string) // [company : [path : [header : value]]
var bodyPerCompany = make(map[string]map[string]string)               // [company : [path : body]]
var codePerCompany = make(map[string]map[string]int)                  // [company : [path : intCode]]

func main() {
	port := ":80"
	if len(os.Args) >= 2 {
		port = os.Args[1]
	}
	reinitGlobalVariables()
	log.Println("Server version: 0.08")

	http.HandleFunc("/", handleRequests)
	http.HandleFunc("/setEverything", setEverything)
	http.HandleFunc("/clear", clear)

	// static responses - same for each company
	http.HandleFunc("/bob/api/v1/service/instance/list", instanceList)
	http.HandleFunc("/bob/api/v2/service/instance/list", instanceList)

	http.HandleFunc("/bob/api/v1/service/definition/listVisible", listVisible)
	http.HandleFunc("/bob/api/v2/service/definition/listVisible", listVisible)
	
	http.HandleFunc("/bob/api/v2/service/instance/listAllPerCompany", listAllPerCompany)

	ch := make(chan error)
	go func() {
		ch <- http.ListenAndServe(port, nil)
	}()

	log.Printf("Customizable response server is listening on %s", port)
	log.Fatal(<-ch)
}

func handleRequests(w http.ResponseWriter, req *http.Request) {
	companyIDFromURLPath := getCompanyID(req)
	path := req.RequestURI

	log.Printf("CompanyId : %s  Endpoint: %s - Returning following response", companyIDFromURLPath, path)

	headers, ok := headersPerCompany[companyIDFromURLPath][path]
	if ok {
		log.Printf("	Company specific headers")
	} else {
		log.Println("	Default Headers")
		headers = headersPerCompany["0"]["/"]
	}

	for name, value := range headers {
		w.Header().Add(name, value)
	}

	if code, ok := codePerCompany[companyIDFromURLPath][path]; ok {
		log.Printf("	Company specific StatusCode: %d", code)
		w.WriteHeader(code)
	} else {
		log.Printf("	Default StatusCode: %d", codePerCompany["0"]["/"])
		w.WriteHeader(codePerCompany["0"]["/"])
	}

	if body, ok := bodyPerCompany[companyIDFromURLPath][path]; ok {
		log.Printf("	Company specific body")
		w.Write([]byte(body))
	} else {
		log.Printf("	Default body returned")
		w.Write([]byte(bodyPerCompany["0"]["/"]))
	}
}

func setStatusCode(companyID string, jsonBody map[string]interface{}) {
	value, ok := jsonBody["statusCode"]
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

func setHeaders(companyID string, jsonBody map[string]interface{}) {
	value, ok := jsonBody["headers"]
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

func setBody(companyID string, jsonBody map[string]interface{}) {
	value, ok := jsonBody["body"]
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

func setEverything(w http.ResponseWriter, req *http.Request) {
	var companyID = getCompanyIDFromHeader(req)
	jsonBody, e := getJsonBody(req)
	if e != nil {
		log.Fatal(e)
		return
	}

	setHeaders(companyID, jsonBody)
	setBody(companyID, jsonBody)
	setStatusCode(companyID, jsonBody)

	log.Printf("SetEverything done for CompanyId %s", companyID)
}

func clear(w http.ResponseWriter, req *http.Request) {
	var companyID = req.Header.Get("companyId")
	if len(companyID) < 1 {
		log.Printf("Clearing all setting")
		reinitGlobalVariables()
	} else {
		log.Printf("Clearing setting for company %s", companyID)
		headersPerCompany[companyID] = make(map[string]map[string]string)
		bodyPerCompany[companyID] = make(map[string]string)
		codePerCompany[companyID] = make(map[string]int)
	}
}

func reinitGlobalVariables() {
	headersPerCompany = make(map[string]map[string]map[string]string) // [company : [header : value]]
	headersPerCompany["0"] = make(map[string]map[string]string)

	bodyPerCompany = make(map[string]map[string]string) // [company : body]
	bodyPerCompany["0"] = make(map[string]string)
	bodyPerCompany["0"]["/"] = defaultJsonBodyForSetEverythingRequest

	codePerCompany = make(map[string]map[string]int) // [company : intCode]
	codePerCompany["0"] = make(map[string]int)
	codePerCompany["0"]["/"] = 200
}

func getCompanyIDFromHeader(req *http.Request) string {
	var companyID = req.Header.Get("companyid")
	if len(companyID) < 1 {
		companyID = "0"
	}

	return companyID
}

func getCompanyID(req *http.Request) string {
	companyID := req.Header.Get("companyId")
	if len(companyID) < 1 {
		companyID = strings.Split(req.URL.Path, "/")[1]
		if _, err := strconv.ParseInt(companyID,10,64); err != nil {
			companyID = "0"
		}
	}
	
	return companyID
}

func getJsonBody(req *http.Request) (map[string]interface{}, error) {
	var jsonResult map[string]interface{}
	body, e := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	json.Unmarshal(body, &jsonResult)

	return jsonResult, nil
}

// static responses - same for each company
func instanceList(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(instancesListsStaticBody))
	log.Println("Received request on /api/v1/service/instance/list - static services list returned")
}

func listVisible(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(listVisibleStaticBody))
	log.Println("Received request on /api/v1/service/definition/listVisible - static services listVisible returned")
}

func listAllPerCompany(w http.ResponseWriter, req *http.Request) {
	body := `[]`
	w.Write([]byte(body))
	log.Println("Received request on /api/v2/service/instance/listAllPerCompany - empty list returned")
}

var instancesListsStaticBody = `{"Services": [{
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

var listVisibleStaticBody = `{"ServicesDefinitions": [{
		"Name": "DummyName",
		"Version": "1",
		"Description": "BlahBlah",
		"CurrentState": "Running",
		"DesiredState": "Running"
	}]}`

var defaultJsonBodyForSetEverythingRequest = `[Default]
To receive proper response, configure mockServer with "/setEverything" endpoint and 
use companyId header with desired companyId as value and request body in following valid json format:

{
    "body":{
        "/endpoint/to/mock/1":{"exampleNode":[{
            "exampleNode":"exampleNode",
            "exampleNode":100
        }]},
        "/endpoint/to/mock/2" :{"exampleNode": [{
            "exampleNode" : "exampleNode",
            "exampleNode" : "2017-09-08T19:01:55.714942+03:00"
        }]}
    },
    "headers":{
        "/endpoint/to/mock/1":{
            "ExampleHeader1" : "ExampleHeaderValue1",
            "ExampleHeader2" : "ExampleHeaderValue2"
        },
        "/endpoint/to/mock/2":{
            "ExampleHeader1" : "ExampleHeaderValue1",
            "ExampleHeader2" : "ExampleHeaderValue2"
        }
    },
    "statusCode":{
        "/endpoint/to/mock/1":200,
        "/endpoint/to/mock/2":403
   }
}`