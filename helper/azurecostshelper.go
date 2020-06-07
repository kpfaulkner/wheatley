package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SubscriptionCosts struct {
	SubscriptionID string
	Total float64 // total for subscription, not really concerted about using floats here.
	ResourceGroupCosts map[string]float64
}

func NewSubscriptionCosts(subscriptionID string) SubscriptionCosts {
	s := SubscriptionCosts{}
	s.SubscriptionID = subscriptionID
	s.ResourceGroupCosts = make(map[string]float64)
	return s
}

type DailyBillingDetails struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Tags       interface{} `json:"tags"`
	Properties struct {
		BillingPeriodID         string      `json:"billingPeriodId"`
		UsageStart              time.Time   `json:"usageStart"`
		UsageEnd                time.Time   `json:"usageEnd"`
		InstanceID              string      `json:"instanceId"`
		InstanceName            string      `json:"instanceName"`
		InstanceLocation        string      `json:"instanceLocation"`
		MeterID                 string      `json:"meterId"`
		UsageQuantity           float64         `json:"usageQuantity"`
		PretaxCost              float64     `json:"pretaxCost"`
		Currency                string      `json:"currency"`
		IsEstimated             bool        `json:"isEstimated"`
		SubscriptionGUID        string      `json:"subscriptionGuid"`
		SubscriptionName        string      `json:"subscriptionName"`
		Product                 string      `json:"product"`
		ConsumedService         string      `json:"consumedService"`
		PartNumber              string      `json:"partNumber"`
		ResourceGUID            string      `json:"resourceGuid"`
		OfferID                 string      `json:"offerId"`
		ChargesBilledSeparately bool        `json:"chargesBilledSeparately"`
		MeterDetails            interface{} `json:"meterDetails"`
	} `json:"properties"`
}

type BillingResponse struct {
	NextLink string `json:"nextLink"`
	Value []DailyBillingDetails `json:"value"`
}

type AzureCost struct {
	azureAuth *AzureAuth
	subscriptionID string
	tenantID string
	clientID string
	clientSecret string
}

func NewAzureCost(tenantID string, clientID string, clientSecret string) AzureCost {
	a := AzureCost{}
	a.tenantID = tenantID
	a.clientID = clientID
	a.clientSecret = clientSecret
	a.azureAuth = NewAzureAuth(tenantID, clientID, clientSecret)

	return a
}

// just testing out ideas....   naming rocks.
func (ac *AzureCost) GetAllBillingForSubscriptionID( subscriptionID string, startDate time.Time, endDate time.Time) ([]DailyBillingDetails, error) {
	err := ac.azureAuth.RefreshToken()
	if err != nil {
		fmt.Printf("unable to refresh token: %s\n", err.Error())
		return nil, err
	}


	// taken from https://docs.microsoft.com/en-us/azure/cost-management-billing/costs/quick-acm-cost-analysis   unsure if works yet
	// https://management.azure.com/{scope}/providers/Microsoft.Consumption/usageDetails?metric=AmortizedCost&$filter=properties/usageStart+ge+'2019-04-01'+AND+properties/usageEnd+le+'2019-04-30'&api-version=2019-04-01-preview
	// THiS WORKSSSS template := "https://management.azure.com/subscriptions/%s/providers/Microsoft.Consumption/usageDetails?metric=ActualCost&$filter=properties/usageStart+ge+'2020-04-01'+AND+properties/usageEnd+le+'2020-04-30'&api-version=2019-04-01-preview"
	template := "https://management.azure.com/subscriptions/%s/providers/Microsoft.Consumption/usageDetails?metric=ActualCost&$filter=properties/usageStart+ge+'%s'+AND+properties/usageEnd+le+'%s'&api-version=2019-04-01-preview"

	url := fmt.Sprintf(template, subscriptionID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	done := false
	billingDetails := []DailyBillingDetails{}
	for !done {
		// replace spaces with +.... should call proper encode...
		url = strings.Replace(url, " ", "+",-1)
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("couldn't generate HTTP request %s\n", err.Error())
			return nil, err
		}

		client := http.Client{}
		request.Header.Set("Authorization", "Bearer "+ac.azureAuth.CurrentToken().AccessToken)
		resp, err := client.Do(request)
		if err != nil {
			fmt.Printf("couldn't execute HTTP request %s\n", err.Error())
			return nil, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		br := BillingResponse{}
		err = json.Unmarshal(body, &br)
		if err != nil {
			return nil, err
		}

		billingDetails = append(billingDetails, br.Value...)
		if br.NextLink == "" {
			done = true
		} else {
			url = br.NextLink
		}
	}

	return billingDetails, nil
}

func CalculateCostsPerResourceGroup( data []DailyBillingDetails) (map[string]float64, float64,  error) {
	resourceGroupCosting := make(map[string]float64)

	total := 0.0
	for _, costDetails := range data {

		sp := strings.Split(costDetails.Properties.InstanceID, "/")
		rg := strings.ToLower(sp[4])  // just deal with lowercase.

		// default value is 0.0 :)
		cost, _ := resourceGroupCosting[rg]
		cost += costDetails.Properties.PretaxCost
		resourceGroupCosting[rg] = cost
		total += costDetails.Properties.PretaxCost

	}

	return resourceGroupCosting, total, nil
}

func (ac *AzureCost) GenerateSubscriptionCostDetails( subscriptionIDs []string, startDate time.Time, endDate time.Time) ([]SubscriptionCosts, error) {

	subscriptionCosts := []SubscriptionCosts{}

	// for this simple case, just lock it.
  var lock sync.Mutex
  var wg sync.WaitGroup

	for _, subscriptionID := range subscriptionIDs {
		sId := subscriptionID
		start := startDate
		end := endDate
		wg.Add(1)
		go func( subID string, startDate time.Time, endDate time.Time) {
			data, err := ac.GetAllBillingForSubscriptionID(subscriptionID, startDate, endDate)
			if err != nil {
				return
			}

			rgData, total, err := CalculateCostsPerResourceGroup(data)
			if err != nil {
				return
			}

			// merge results into subscriptionCosts.
			sc := NewSubscriptionCosts(subscriptionID)
			sc.ResourceGroupCosts = rgData
			sc.Total = total

			lock.Lock()
			subscriptionCosts = append(subscriptionCosts, sc)
			lock.Unlock()
			wg.Done()
		}( sId, start, end)
	}

	wg.Wait()

	return subscriptionCosts, nil
}

// getCostsPerPrefix takes costs that have already been retrieved and give summaries where
// particular prefixes are met.
// ie, will search for RG prefixes of "test-" for the testenv etc.
// Simply return a map of prefix and total for RGs matching that prefix.
func GetCostsPerRGPrefix( wantedPrefixes []string, allData []SubscriptionCosts ) (map[string]float64, error) {

	prefixData := make(map[string]float64)

	// stupid level of iteration, but will probably be fine...  (<--- famous last words!)
	for _, sc := range allData {
		for rg,v := range sc.ResourceGroupCosts {
			// check if RG name has a prefix we're interested in.
			for _, prefix := range wantedPrefixes {
				if strings.HasPrefix(rg, prefix) {
					cost := prefixData[prefix]
					cost += v
					prefixData[prefix] = cost
				}
			}
		}
	}

	return prefixData, nil
}

func contains( key string, l []string) bool {
	for _,s := range l {
		if s == key {
			return true
		}
	}
	return false
}

func FilterDataBasedOnSubscription( allData []SubscriptionCosts, subIDs []string)  []SubscriptionCosts {

	data := []SubscriptionCosts{}

	for _, subData := range allData {
		if contains( subData.SubscriptionID, subIDs) {
			data = append(data, subData)
		}
	}
	return data
}
