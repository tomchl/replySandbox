package main

import (
	"bytes"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"strings"

	"github.com/martindrlik/org/fakegcm"
	"github.com/tomchl/apns"
)

func main() {
	log.Println("Server version: 0.96") // change this when pushing to dockerhub to be aware of the current version

	iosPort := getPortFromEnv("IosPort", "8095")

	androidPort := getPortFromEnv("AndroidPort", "8096")

	currenturl := os.Getenv("CurrUrl")
	ownerEmail := os.Getenv("UserEmail")

	if currenturl != "" {
		_, err := mail.ParseAddress(ownerEmail)
		if err != nil {
			log.Fatal("Wrong email address format. Specify env variable 'UserEmail' correctly.")
			return
		}

		log.Printf("Generating certificates for url : %s", currenturl)
		generateCertificatesForURL(currenturl)
		log.Printf("Certificates generation done...")
	} else {
		log.Printf("Using default selfsigned certificates, if you want to create valid certificates, specify env variables 'CurrUrl' and 'UserEmail' before starting the server.")
	}

	ch := make(chan error)
	go func() {
		ch <- apns.ListenAndServeTLS(iosPort, "cert.pem", "key.pem", "APNS_Alfa.p12", "Abcd1234")
	}()
	log.Printf("Apns server is listening on port %s", iosPort)

	go func() {
		ch <- fakegcm.ListenAndServeTLS(fakegcm.Configuration{Addr: androidPort, CertFile: "cert.pem", KeyFile: "key.pem", ConfirmDelivery: true, MessageOnly: false})
	}()
	log.Printf("Gcm	 server is listening on port %s", androidPort)
	log.Fatal(<-ch)
}

func getPortFromEnv(key, defaultPort string) string {
	value := os.Getenv(key)
	if value == "" {
		return ":" + defaultPort
	}
	return ":" + value
}

func generateCertificatesForURL(url string) {
	cmd := exec.Command("bash", "createcert.sh")
	cmd.Stdin = strings.NewReader("")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Certificate generation output : \n %s", out.String())
}
