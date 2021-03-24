package helper

import (
	"encoding/json"
	"fmt"
	"github.com/google/martian/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TimeSeries struct {
	Metadatavalues []interface{} `json:"metadatavalues"`
	Data           []struct {
		TimeStamp time.Time `json:"timeStamp"` // we will get one of these per minute.
		Average   float64   `json:"average"`
	} `json:"data"`
}

type ResourceMetric struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name struct {
		Value          string `json:"value"`
		LocalizedValue string `json:"localizedValue"`
	} `json:"name"`
	Unit       string       `json:"unit"`
	Timeseries []TimeSeries `json:"timeseries"`
}

type MetricResponse struct {
	Cost           int              `json:"cost"`
	Timespan       string           `json:"timespan"`
	Interval       string           `json:"interval"`
	Value          []ResourceMetric `json:"value"`
	Namespace      string           `json:"namespace"`
	Resourceregion string           `json:"resourceregion"`
}

// used to get metrics via AzureMonitor (as opposed to app insights)
type AzureMonitorHelper struct {
	config AzureMonitoringConfigMap

	azureAuthMap map[string]*AzureAuth
	//azureAuth *AzureAuth
}

// NewAzureMonitorHelper does the Azure specifics....
func NewAzureMonitorHelper(config AzureMonitoringConfigMap) *AzureMonitorHelper {
	ah := AzureMonitorHelper{}
	ah.config = config

	// which env?
	//ah.azureAuth = NewAzureAuth(config)
	aam, err := generateAzureAuthMap(config)
	if err != nil {
		fmt.Printf("KABOOM with AzureMonitorHelper!! %s\n", err.Error())
		panic(err)
	}
	ah.azureAuthMap = aam
	return &ah
}

func generateAzureAuthMap(config AzureMonitoringConfigMap) (map[string]*AzureAuth, error) {

	aam := make(map[string]*AzureAuth)
	for k, v := range config.AzureMonitorMap {
		auth := NewAzureAuth(v.TenantID, v.ClientID, v.ClientSecret)
		aam[k] = auth
	}

	return aam, nil
}

func (ah *AzureMonitorHelper) currentToken(env string) AzureAuthToken {
	aa := ah.azureAuthMap[env]
	return aa.currentToken
}

// wrapper around AzureAuth instance.
func (ah *AzureMonitorHelper) refreshToken(env string) error {
	aa := ah.azureAuthMap[env]
	err := aa.RefreshToken()
	return err
}

func (ah *AzureMonitorHelper) GetMetrics(env string, subscriptionID string, resourceGroup string, metricDefinition string, resourceName string, startTime time.Time, endTime time.Time, metricNamesSlice []string) (*MetricResponse, error) {

	err := ah.refreshToken(env)
	if err != nil {
		return nil, err
	}

	// should be something like: "serverLoad,usedmemorypercentage,usedmemory"
	metricNames := url.PathEscape(strings.Join(metricNamesSlice, ","))
	template := "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/%s/%s/providers/microsoft.insights/metrics?metricnames=%s&timespan=%s/%s&aggregation=Average&api-version=2018-01-01"
	url := fmt.Sprintf(template, subscriptionID, resourceGroup, metricDefinition, resourceName, metricNames, startTime.Format("2006-01-02T15:04:05Z"), endTime.Format("2006-01-02T15:04:05Z"))

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	request.Header.Set("Authorization", "Bearer "+ah.currentToken(env).AccessToken)
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := string(body)
	fmt.Printf("body is %s\n", s)

	mr := MetricResponse{}
	err = json.Unmarshal(body, &mr)
	if err != nil {
		return nil, err
	}
	return &mr, nil
}

// takes a time series and generates average over time period.
//
func generateAverageForTimeSpan(resourceName string, metricName string, unit string, timeSeries []TimeSeries) (string, error) {

	total := 0.0
	count := 0
	for _, ts := range timeSeries {
		for _, d := range ts.Data {
			total += d.Average
			count++
		}
	}

	average := total / float64(count)

	if unit == "Percent" {
		return fmt.Sprintf("%s has average %s is %0.2f%%", resourceName, metricName, average), nil
	}

	if unit == "Bytes" {
		return fmt.Sprintf("%s has average %s is %0.0f bytes ", resourceName, metricName, average), nil
	}

	return "", nil
}

func (ah *AzureMonitorHelper) GetResourceMetrics(env string, resourceFriendlyName string, resourceGroup string, resourceName string, metricDefinition string, metrics []string, spanInMinutes int, ch chan string) {

	start := time.Now().UTC().Add(time.Duration(spanInMinutes) * time.Minute * -1)
	end := time.Now().UTC()
	resp, err := ah.GetMetrics(env, ah.config.AzureMonitorMap[env].SubscriptionID, resourceGroup, metricDefinition, resourceName, start, end, metrics)
	if err != nil {
		// ignore for moment...
		log.Errorf("blew up when getting redis?!? %s\n", err.Error())
		return
	}

	fmt.Printf("resp is %v\n", resp)

	for _, metric := range resp.Value {
		av, err := generateAverageForTimeSpan(resourceFriendlyName, metric.Name.LocalizedValue, metric.Unit, metric.Timeseries)
		if err == nil {
			ch <- av
		}
	}

}
