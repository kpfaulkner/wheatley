package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type AzureAppSettings struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Location   string `json:"location"`
	Properties map[string]string  `json:"properties"`
}


type AzureAppServiceHelper struct {
  azureAuth *AzureAuth
	subscriptionID string
  tenantID string
  clientID string
  clientSecret string
}

func NewAzureAppServiceHelper(subscriptionID string, tenantID string, clientID string, clientSecret string ) *AzureAppServiceHelper {
	ah := AzureAppServiceHelper{}
	ah.azureAuth = NewAzureAuth(tenantID, clientID, clientSecret)
	ah.subscriptionID = subscriptionID
	ah.tenantID = tenantID
	ah.clientID = clientID
	ah.clientSecret = clientSecret
	return &ah
}

func (ah *AzureAppServiceHelper) currentToken() AzureAuthToken {
	return ah.azureAuth.currentToken
}

// wrapper around AzureAuth instance.
func (ah *AzureAppServiceHelper) refreshToken() error {
	err := ah.azureAuth.RefreshToken()
	return err
}


// GetAppServiceAppSettings get app settings... get them all dammit!!
// Just return a map of string/string. No need for anything fancy.
func (ah *AzureAppServiceHelper) GetAppServiceAppSettings(subscriptionID string, resourceGroup string,  appServerName string) (*AzureAppSettings,error) {

	// refresh all the tokens!!!
	err := ah.refreshToken()
	if err != nil {
		return nil, err
	}

  template := "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Web/sites/%s/config/appsettings/list?api-version=2019-08-01"
	url := fmt.Sprintf(template, subscriptionID, resourceGroup, appServerName)

  // POST to get it... REALLY?  naughty Azure :)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	req.Header.Set("Authorization", "Bearer "+ah.currentToken().AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil,err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil,err
	}

	appSettings := AzureAppSettings{}

	err = json.Unmarshal(body, &appSettings)
	if err != nil {
		return nil,err
	}

	// just return a basic map[string]string to keep things simple.
	return &appSettings, nil
}

	// SetAppServiceAppSettings making bold assumption that key/value can always be strings.
func (ah *AzureAppServiceHelper) SetAppServiceAppSettings(subscriptionID string, resourceGroup string,  appServerName string, appSettings AzureAppSettings) error {

	// refresh all the tokens!!!
	err := ah.refreshToken()
	if err != nil {
		return err
	}

	// https://management.azure.com/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/sites/{name}/config/web?api-version=2019-08-01
	template := "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Web/sites/%s/config/appsettings?api-version=2019-08-01"
	url := fmt.Sprintf(template, subscriptionID, resourceGroup, appServerName)

	jsonBytes,err :=json.Marshal(appSettings)
	if err != nil {
		return err
	}

	fmt.Printf("body %s\n", string(jsonBytes))

	req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")


	client := http.Client{}
	req.Header.Set("Authorization", "Bearer "+ah.currentToken().AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("resp %s\n", string(body))


	return  nil
}
