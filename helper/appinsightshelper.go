package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kpfaulkner/wheatley/models/Metrics"
	"io/ioutil"
	"net/http"
	"strings"
)

type AppInsightsHelper struct {
  //configMap map[string]AppInsightConfig
  config AzureMonitoringConfigMap
}

func NewAppInsightsHelper( config AzureMonitoringConfigMap ) *AppInsightsHelper {
	ah := AppInsightsHelper{}
	ah.config = config
	return &ah
}

// Checks app insights, which it tied to specific types of infra anyway.
func getPerfCounter(appID string, apiKey string, perfCounter Metrics.AzureMetricName, timeSpanInMinutes int) (string, error) {

	ts := fmt.Sprintf("PT%dM", timeSpanInMinutes)
	client := http.Client{}
	template := "https://api.applicationinsights.io/v1/apps/%s/metrics/%s?timespan=%s&interval=%s&segment=cloud/roleName"
	url := fmt.Sprintf(template, appID, perfCounter, ts,ts)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("x-api-key", apiKey)
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (aih AppInsightsHelper) getAppInsightsCreds(env string, appInsightName string) (string, string, error) {
	config, ok := aih.config.AppInsightsMap[ env]
	if ok {

		// manual loop through, ugly but is short.
		for _, r := range config.Resources {
			if strings.ToLower(r.Name) == strings.ToLower(appInsightName) {
				return r.AppID, r.APIKey, nil
			}
		}
	}

	return "","",errors.New("Unable to find config for env")
}

// GetCPUAverage gets the average CPU usage over a given time period.
// Will try and make this more generic it expands.
// Return the data via a channel
func (aih AppInsightsHelper) GetCPUAverage( env string, appInsightsName string, spanInMinutes int, ch chan string ) {

	appID, apiKey, err := aih.getAppInsightsCreds(env, appInsightsName)
	if err != nil {
		return // nothing we can do. Maybe log it?
	}


	res, err := getPerfCounter( appID, apiKey, Metrics.ProcessorCpuPercentageMetric, spanInMinutes)
	if err != nil {
		ch <- ""
	}

	var perfStats Metrics.CPUProcessorAveragePercentage
	json.Unmarshal([]byte(res), &perfStats)

	// There are 2 lots of segments (go figure). Time Segments (time range split over multiple segments).
	// second segment is via role!
	// Only care about first segment for now...
	for _, seg := range perfStats.Value.TimeSegments[0].Segments {
		ch <- fmt.Sprintf("Average CPU percent for role %s over %d minutes is %.2f%%", seg.CloudRoleName, spanInMinutes, seg.CPUPercentage.Avg)
	}
}

// GetMemoryAverage gets the average memory available over given span.
func (aih AppInsightsHelper) GetMemoryAverage( env string,appInsightsName string, spanInMinutes int, ch chan string  ) {

	appID, apiKey, err := aih.getAppInsightsCreds(env, appInsightsName)
	if err != nil {
		return // nothing we can do. Maybe log it?
	}

	res, err := getPerfCounter( appID, apiKey, Metrics.MemoryAvailableBytesMetric, spanInMinutes)
	if err != nil {
		ch <- ""
	}

	var perfStats Metrics.MemoryAvailableBytes
	json.Unmarshal([]byte(res), &perfStats)

	// There are 2 lots of segments (go figure). Time Segments (time range split over multiple segments).
	// second segment is via role!
	// Only care about first segment for now...
	for _, seg := range perfStats.Value.TimeSegments[0].Segments {
		mem := int(seg.MemoryAvailable.Avg / (1024*1024))
		ch <- fmt.Sprintf("Average Memory available for role %s over %d minutes is %dMB", seg.CloudRoleName, spanInMinutes, mem)
	}
}

