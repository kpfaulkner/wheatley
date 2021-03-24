package messagehandlers

import (
	"errors"
	"github.com/kpfaulkner/wheatley/helper"
	"github.com/kpfaulkner/wheatley/models"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GET /v1/apps/0444a374-65a4-429c-83d7-f7130658b6b8/metrics/performanceCounters/processorCpuPercentage?timespan=PT4H
// GET /v1/apps/<app insights APP ID on API access page>/metrics/performanceCounters/processorCpuPercentage?timespan=PT4H
// header x-api-key <API key generated on same page above!>
//

// ServerStatusMessageHandler stores status about server (test, stage, prod etc) status
type AzureStatusMessageHandler struct {
	state    models.State
	AIHelper *helper.AppInsightsHelper
	AMHelper *helper.AzureMonitorHelper
	config   helper.AzureMonitoringConfigMap
}

func NewAzureStatusMessageHandler() *AzureStatusMessageHandler {
	asHandler := AzureStatusMessageHandler{}

	config, err := helper.LoadAzureMonitoringConfig("azuremonitoring.json")
	if err != nil {
		log.Fatalf("Unable to monitoring Azure... something is broken %s\n", err.Error())
	}

	asHandler.config = *config
	asHandler.AIHelper = helper.NewAppInsightsHelper(*config)
	asHandler.AMHelper = helper.NewAzureMonitorHelper(*config)

	return &asHandler
}

func (ss *AzureStatusMessageHandler) GetStat(f func(string, int, chan string), env string, minutes int, ch chan string) {
	go func(env string, min int, ch chan string) {
		f(env, min, ch)
	}(env, minutes, ch)
}

// checkProd checks all the various parts of the prod environment
// and combines the results here.
// Won't try and do anything overly intelligent here, just basic calling of Azure monitor/app-insights code.
// Will be VERY hard coded here about which services to call (AM/AI) and what the resource
// names are. Will make it more intelligent later.
func (ss *AzureStatusMessageHandler) checkEnv(env string, minsToCheck int) (string, error) {
	ch := make(chan string, 20)

	// App insights... get details for env.
	aiEnvConfig := ss.config.AppInsightsMap[env]
	for _, aiRes := range aiEnvConfig.Resources {
		go ss.AIHelper.GetCPUAverage(env, aiRes.Name, minsToCheck, ch)
	}

	// Azure Montitor config.
	amEnvConfig := ss.config.AzureMonitorMap[env]
	for _, amRes := range amEnvConfig.ResourceToMonitor {
		go ss.AMHelper.GetResourceMetrics(env, amRes.Name, amRes.ResourceGroup, amRes.ResourceName, amRes.MetricDefinition, amRes.Metrics, minsToCheck, ch)
	}

	responseList := []string{}
	// loop forever (for 2 goroutines this is overkill, but not much overhead in preparation for other
	// metrics to soon be returned.
	quit := false
	for !quit {
		select {
		case msg := <-ch:
			responseList = append(responseList, msg)
			break
		case <-time.After(5 * time.Second):
			// quit...  quit it all. You've had your time, now bugger off :)
			quit = true
		}
	}

	sort.Strings(responseList)
	return strings.Join(responseList, "\n"), nil
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (ss *AzureStatusMessageHandler) ParseMessage(msg string, user string) (MessageResponse, error) {

	checkAzureStatusRegex := regexp.MustCompile(`^check (.*)`)
	checkAzureStatusLastRegex := regexp.MustCompile(`^check (.*?) last (.*?) min`)
	soundOffRegex := regexp.MustCompile(`sound off`)
	helpRegex := regexp.MustCompile(`^help$`)

	msg = strings.ToLower(msg)
	switch {

	case checkAzureStatusLastRegex.MatchString(msg):
		res := checkAzureStatusLastRegex.FindStringSubmatch(msg)
		if res != nil && len(res) == 3 {
			env := strings.ToLower(res[1])
			if env == "test" || env == "prod" {

				minsCheck, err := strconv.Atoi(res[2])
				if err != nil || minsCheck < 5 || minsCheck > 60 {
					return NewTextMessageResponse("not a valid number... please pick something between 5 and 60"), nil
				}
				answer, err := ss.checkEnv(res[1], minsCheck)
				if err != nil {
					return NewTextMessageResponse("unable to get answer"), nil
				}
				return NewTextMessageResponse(answer), nil
			}

			return NewTextMessageResponse("Please check your query... if you think it's right... complain to Ken."), nil
		}

	case checkAzureStatusRegex.MatchString(msg):
		res := checkAzureStatusRegex.FindStringSubmatch(msg)
		if res != nil {
			env := strings.ToLower(res[1])
			if env == "test" || env == "prod" {
				answer, err := ss.checkEnv(res[1], 5)
				if err != nil {
					return NewTextMessageResponse("unable to get answer"), nil
				}
				return NewTextMessageResponse(answer), nil
			}

			return NewTextMessageResponse("Please check your query... if you think it's right... complain to Ken."), nil
		}

	case soundOffRegex.MatchString(msg):
		return NewTextMessageResponse("ServerStatusMessageHandler reporting for duty"), nil

	case helpRegex.MatchString(msg):
		return NewTextMessageResponse("check <test|prod>  : gives details about test/prod env.\ncheck <test/prod> last <n> mins : will check env for last n mins, where 5<= n <= 60"), nil

	}
	return NewTextMessageResponse(""), errors.New("No match")

}
