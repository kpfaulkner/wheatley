package messagehandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kpfaulkner/act/pkg"
	"log"
	"os"
	"regexp"
	"strings"
)

type AzureShutdownConfig struct {
	SubscriptionID   string `json:"SubscriptionID"`
	TenantID         string `json:"TenantID"`
	ClientID         string `json:"ClientID"`
	ClientSecret     string `json:"ClientSecret"`
	AllowedUsersList []string
}

// AzureCostMessageHandler gets the costs from Azure Billing API.
type AzureShutdownMessageHandler struct {
	config *AzureShutdownConfig
	azureClassic *pkg.AzureClassic
}

func NewAzureShutdownMessageHandler() *AzureShutdownMessageHandler {
	asHandler := AzureShutdownMessageHandler{}

	config, err := loadAzureShutdownConfig("azureshutdown.json")
	if err != nil {
		log.Fatalf("Cannot read azure costs config :  %s\n", err.Error())
	}

	asHandler.config = config
  asHandler.azureClassic = pkg.NewAzureClassic(asHandler.config.TenantID, asHandler.config.SubscriptionID, asHandler.config.ClientID, asHandler.config.ClientSecret)
	return &asHandler
}

func loadAzureShutdownConfig(configFileName string) (*AzureShutdownConfig, error) {
	var config AzureShutdownConfig
	configFile, err := os.Open(configFileName)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return &config, nil
}

func (as *AzureShutdownMessageHandler) shutdownEnv(env string, rg string) error {

	// INTENTIONALLY COMMENTED OUT
	// since we dont want to accidentally delete something important.
	// If you're using this.... beware!
	// as.azureClassic.DeleteCloudServiceDeployment(rg, env, "production")
	return nil
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (as *AzureShutdownMessageHandler) ParseMessage(msg string, user string) (MessageResponse, error) {

	shutdownPerf2Regex := regexp.MustCompile(`^shutdown (.*) in rg (.*)$`)
	soundOffRegex := regexp.MustCompile(`sound off`)
	helpRegex := regexp.MustCompile(`^help$`)

	msg = strings.ToLower(msg)
	switch {

	case shutdownPerf2Regex.MatchString(msg):
		res := shutdownPerf2Regex.FindStringSubmatch(msg)
		if res != nil && len(res) == 2 {
			env := strings.ToLower(res[1])
			rg := strings.ToLower(res[2])
			err := as.shutdownEnv(env, rg)
			if err != nil {
				return NewTextMessageResponse(fmt.Sprintf("env %s being shutdown", env)), nil
			}
		}

		return NewTextMessageResponse("Please check your query... if you think it's right... complain to Ken."), nil

	case soundOffRegex.MatchString(msg):
		return NewTextMessageResponse("AzureShutdownMessageHandler reporting for duty"), nil

	case helpRegex.MatchString(msg):
		return NewTextMessageResponse("shutdown <env> in rg <rg> : will shutdown the cloud services for the given env (alias) in a particular resource group"), nil

	}
	return NewTextMessageResponse(""), errors.New("No match")

}
