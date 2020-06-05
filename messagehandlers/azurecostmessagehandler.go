package messagehandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kpfaulkner/wheatley/helper"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type AzureCostsConfig struct {
	SubscriptionID string   `json:"SubscriptionID"`
	TenantID       string   `json:"TenantID"`
	ClientID       string   `json:"ClientID"`
	ClientSecret   string   `json:"ClientSecret"`
	Subscriptions  []string
	AllowedUsersList []string

}


// AzureCostMessageHandler gets the costs from Azure Billing API.
type AzureCostsMessageHandler struct {
  config *AzureCostsConfig
}

func NewAzureCostMessageHandler() *AzureCostsMessageHandler {
	asHandler := AzureCostsMessageHandler{}

	config,err := loadAzureCostsConfig( "azurecosts.json")
	if err != nil {
		log.Fatalf("Cannot read azure costs config :  %s\n", err.Error())
	}

	asHandler.config = config

	return &asHandler
}


func loadAzureCostsConfig(configFileName string) (*AzureCostsConfig, error) {
	var config AzureCostsConfig
	configFile, err := os.Open(configFileName)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return &config, nil
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (ss *AzureCostsMessageHandler) ParseMessage(msg string, user string) (string, error) {

	reportAzureCostsRegex := regexp.MustCompile(`^report azurecosts from (.*) to (.*)$`)
	reportAzureCostsForRGRegex := regexp.MustCompile(`^report azurecosts from (.*) to (.*) with prefix (.*)$`)
	soundOffRegex := regexp.MustCompile(`sound off`)
	helpRegex := regexp.MustCompile(`^help$`)

	msg = strings.ToLower(msg)
	switch {

	case reportAzureCostsRegex.MatchString(msg):
		res := reportAzureCostsRegex.FindStringSubmatch(msg)
		if res != nil && len(res) == 3 {
			startDateStr := strings.ToLower(res[1])
			endDateStr := strings.ToLower(res[2])

			layout := "2006-01-02"
			startDate, err := time.Parse(layout, startDateStr)
			if err != nil {
				return "Start date should be in YYYY-MM-DD format", nil
			}
			endDate, err := time.Parse(layout, endDateStr)
			if err != nil {
				return "End date should be in YYYY-MM-DD format", nil
			}

			ac := helper.NewAzureCost(ss.config.TenantID, ss.config.ClientID, ss.config.ClientSecret)
			subCosts,err  := ac.GenerateSubscriptionCostDetails(ss.config.Subscriptions, startDate, endDate)
			if err != nil {
				fmt.Printf("Error generating sub costs %s\n", err.Error())
				return "Unable to generate subscription costs.", nil
			}

			subTotals := []string{}
			total := 0.0
			for _,sc := range subCosts {
				subTotals = append(subTotals, fmt.Sprintf("Total of sub %s is %0.2f",sc.SubscriptionID,sc.Total))
				total += sc.Total
			}
			subTotals = append(subTotals, fmt.Sprintf("TOTAL is %0.2f", total))
			return strings.Join(subTotals,"\n" ), nil
		}

		return "Please check your query... if you think it's right... complain to Ken.", nil

	case reportAzureCostsForRGRegex.MatchString(msg):
		res := reportAzureCostsForRGRegex.FindStringSubmatch(msg)
		if res != nil && len(res) == 4 {
			startDateStr := strings.ToLower(res[1])
			endDateStr := strings.ToLower(res[2])
			prefix := strings.ToLower(res[3])

			layout := "2006-01-02"
			startDate, err := time.Parse(layout, startDateStr)
			if err != nil {
				return "Start date should be in YYYY-MM-DD format", nil
			}
			endDate, err := time.Parse(layout, endDateStr)
			if err != nil {
				return "End date should be in YYYY-MM-DD format", nil
			}

			ac := helper.NewAzureCost(ss.config.TenantID, ss.config.ClientID, ss.config.ClientSecret)
			subCosts,err  := ac.GenerateSubscriptionCostDetails(ss.config.Subscriptions, startDate, endDate)
			if err != nil {
				fmt.Printf("Error generating sub costs %s\n", err.Error())
				return "Unable to generate subscription costs.", nil
			}

			subTotals := []string{}
			total := 0.0
			for _,sc := range subCosts {
				subTotals = append(subTotals, fmt.Sprintf("Total of sub %s is %0.2f",sc.SubscriptionID,sc.Total))
				total += sc.Total
			}

			// just prefix data
			prefixCosts, err  := helper.GetCostsPerRGPrefix([]string{prefix}, subCosts)
			if err != nil {
				return "Unable to get costs for prefix", nil
			}

			for prefix,cost := range prefixCosts {
				subTotals = append(subTotals, fmt.Sprintf("Prefix %s cost %0.2f", prefix, cost))
			}

			subTotals = append(subTotals, fmt.Sprintf("TOTAL is %0.2f", total))
			return strings.Join(subTotals,"\n" ), nil
		}
	case soundOffRegex.MatchString(msg):
		return "AzureCostsMessageHandler reporting for duty", nil

	case helpRegex.MatchString(msg):
		return "report azurecosts from <YYYY-MM-DD> to <YYYY-MM-DD> : Gives costings between the 2 dates. Splits into pre-defined groups.", nil

	}
	return "", errors.New("No match")

}
