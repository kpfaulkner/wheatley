package helper

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type AzureVMHelper struct {
	azureAuth     *AzureAuth
	clientID      string
	clientSecret  string
	tenantID      string
	resourceGroup string
	subscriptionID string
}

func NewAzureVMHelper(subscriptionID string,tenantID string, clientID string, clientSecret string, resourceGroup string) *AzureVMHelper {
	ah := AzureVMHelper{}
	ah.azureAuth = NewAzureAuth(tenantID, clientID, clientSecret)
	ah.resourceGroup = resourceGroup
	ah.tenantID = tenantID
	ah.clientID = clientID
	ah.subscriptionID = subscriptionID
	ah.clientSecret = clientSecret
	return &ah
}

func (ah *AzureVMHelper) currentToken() AzureAuthToken {
	return ah.azureAuth.currentToken
}

// wrapper around AzureAuth instance.
func (ah *AzureVMHelper) refreshToken() error {
	err := ah.azureAuth.RefreshToken()
	return err
}

func generateVMStartupURL(subscriptionID string, rgName string, vmName string) string {
	template := "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s/start?api-version=2020-12-01"
	url := fmt.Sprintf(template, subscriptionID, rgName, vmName)
	return url
}

// StartVM
// See https://docs.microsoft.com/en-us/rest/api/compute/virtualmachines/start for details
func (ah *AzureVMHelper) StartVM(vmName string, rgName string) error {

	// refresh all the tokens!!!
	err := ah.refreshToken()
	if err != nil {
		return err
	}

	url := generateVMStartupURL(ah.subscriptionID,rgName,vmName)
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, nil)
	req.Header.Add("Authorization", "Bearer "+ah.currentToken().AccessToken)
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error on post %s\n", err.Error())
		panic(err)
	}

	fmt.Printf("status code is %d\n", resp.StatusCode)

	// if status begins with 4.... assume failure.
	if strings.HasPrefix(resp.Status, "4") {
		return errors.New("unable to start backup")
	}

	return nil
}
