package utils

import (
	"bytes"
	"crypto/tls"
	"d2api/pkg/requests"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

func SendNotification(tournamentId int64, playerIds []string, content, notifType string, notifySubtype string, metadata map[string]string) {
	data := requests.Notification{
		Content:  content,
		Metadata: requests.Metadata{Ids: metadata},
		UserIds:  playerIds,
		Type:     notifType,
		Subtype:  notifySubtype,
		RefId:    strconv.FormatInt(tournamentId, 10),
		Service:  os.Getenv("TOURNAMENT_SERVICE_HOST"),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling data: ", err)
	}

	serviceURL := os.Getenv("NOTIFICATION_URL")
	endpoint := "/v1/notification"
	fullURL := serviceURL + endpoint

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	_, err = http.Post(fullURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error calling %s: %s", endpoint, err.Error())
	}
	log.Printf("Sent event on... %s", endpoint)
}
