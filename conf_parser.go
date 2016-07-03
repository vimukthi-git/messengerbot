package messengerbot

import (
	"encoding/json"
	"os"
	"log"
)

type Configuration struct {
	ValidationToken string `json:"validation_token"`
	PageAccessToken string `json:"page_access_token"`
}

func getTestConfig() Configuration {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal("error:", err)
	}
	return configuration
}
