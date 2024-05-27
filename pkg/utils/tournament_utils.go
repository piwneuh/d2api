package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
)

func SendMatchResult(endpoint string, match interface{}) bool {
	jsonData, err := json.Marshal(match)
	if err != nil {
		log.Println("Error marshalling data: ", err)
		return false
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error calling %s: %s", endpoint, err.Error())
		return false
	}

	if resp.StatusCode != 200 {
		log.Printf("Error calling %s: %s", endpoint, resp.Status)
		return false
	}

	return true
}
