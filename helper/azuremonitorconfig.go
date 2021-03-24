package helper

import (
	"encoding/json"
	"os"
)

type AppInsightsConfig struct {
	Env       string `json:"env"`
	Resources []struct {
		Name   string `json:"Name"`
		AppID  string `json:"AppID"`
		APIKey string `json:"APIKey"`
	} `json:"resources"`
}

type AzureMonitorResource struct {
	Name             string   `json:"Name"`
	ResourceGroup    string   `json:"ResourceGroup"`
	ResourceName     string   `json:"ResourceName"`
	MetricDefinition string   `json:"MetricDefinition"`
	Metrics          []string `json:"Metrics"`
}

type AzureMonitor struct {
	Name              string                 `json:"Name"`
	SubscriptionID    string                 `json:"SubscriptionID"`
	TenantID          string                 `json:"TenantID"`
	ClientID          string                 `json:"ClientID"`
	ClientSecret      string                 `json:"ClientSecret"`
	ResourceToMonitor []AzureMonitorResource `json:"ResourceToMonitor"`
}

type AzureMonitoringConfig struct {
	AzureMonitor []AzureMonitor `json:"AzureMonitor"`
	AppInsights  struct {
		Configs []AppInsightsConfig `json:"Configs"`
	} `json:"AppInsights"`
}

type AzureMonitoringConfigMap struct {
	AzureMonitoringConfig

	// maps for app insights and azure monitor.... just for quick and easy lookup!
	AppInsightsMap  map[string]AppInsightsConfig
	AzureMonitorMap map[string]AzureMonitor
}

func LoadAzureMonitoringConfig(configFileName string) (*AzureMonitoringConfigMap, error) {
	var config AzureMonitoringConfig
	configFile, err := os.Open(configFileName)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	// generate keys!!!
	amcm := AzureMonitoringConfigMap{}
	amcm.AppInsightsMap = make(map[string]AppInsightsConfig)
	amcm.AzureMonitorMap = make(map[string]AzureMonitor)

	// azure monitor!
	amcm.AzureMonitoringConfig = config
	for _, r := range config.AzureMonitor {
		amcm.AzureMonitorMap[r.Name] = r
	}

	// app insights!!
	for _, r := range config.AppInsights.Configs {
		amcm.AppInsightsMap[r.Env] = r
	}

	return &amcm, nil
}
