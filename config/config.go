package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {

	// database to work against.
	Database struct {
		Host     string `json:"host"`
		Password string `json:"password"`
	} `json:"database"`

	// Azure storage connection string
	AzureStorage string `json:"azurestorage"`

	AzureQueries map[string]string `json:"azurequeries"`

	/*
	   // limit ourselves to 3 queries until I come up with a generic (map?) version of things
	   AzureQuery1 string `json:"azurestoragequery1"`
	   AzureQuery2 string `json:"azurestoragequery2"`
	   AzureQuery3 string `json:"azurestoragequery3"`
	*/
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
