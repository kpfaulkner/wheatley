package helper

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type AzureSQLHelper struct {
  azureAuth *AzureAuth
	subscriptionID string
  tenantID string
  clientID string
  clientSecret string
	sqlAdminLogin string
  sqlAdminPassword string
  storageKey string
  storageURL string
	sqlRgName string
}

func NewAzureSQLHelper(subscriptionID string, tenantID string, clientID string, clientSecret string, sqlAdminLogin string, sqlAdminPassword string, storageKey string, storageURL string, sqlRgName string ) *AzureSQLHelper {
	ah := AzureSQLHelper{}
	ah.azureAuth = NewAzureAuth(subscriptionID, tenantID, clientID, clientSecret)
	ah.subscriptionID = subscriptionID
	ah.tenantID = tenantID
	ah.clientID = clientID
	ah.clientSecret = clientSecret
	ah.sqlAdminLogin = sqlAdminLogin
	ah.sqlAdminPassword = sqlAdminPassword
	ah.storageKey = storageKey
	ah.storageURL = storageURL
	ah.sqlRgName = sqlRgName
	return &ah
}

func (ah *AzureSQLHelper) currentToken() AzureAuthToken {
	return ah.azureAuth.currentToken
}

// wrapper around AzureAuth instance.
func (ah *AzureSQLHelper) refreshToken() error {
	err := ah.azureAuth.RefreshToken()
	return err
}


func generateExportURL(subscriptionID string, rgName string, serverName string, databaseName string) string {
	template := "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Sql/servers/%s/databases/%s/export?api-version=2014-04-01"
	url := fmt.Sprintf(template, subscriptionID, rgName, serverName, databaseName)
	return url
}

func generateExportBody(adminLogin string, adminLoginPassword string, storageKey string, storageKeyType string, storageUri string) string {
	template := `{administratorLogin: "%s",administratorLoginPassword: "%s",storageKey: "%s",storageKeyType: "%s",storageUri: "%s"}`
	body := fmt.Sprintf(template, adminLogin, adminLoginPassword, storageKey, storageKeyType, storageUri)
	return body
}

// GetAppServiceAppSettings get app settings... get them all dammit!!
// Just return a map of string/string. No need for anything fancy.
func (ah *AzureSQLHelper) StartDBExport(serverName string, databaseName string , backupFileName string) error {

	// refresh all the tokens!!!
	err := ah.refreshToken()
	if err != nil {
		return err
	}

	storageURI := fmt.Sprintf("%s/%s", ah.storageURL, backupFileName)
	body := generateExportBody(ah.sqlAdminLogin, ah.sqlAdminPassword, ah.storageKey, "SharedAccessKey", storageURI)
	url := generateExportURL(ah.subscriptionID, ah.sqlRgName, serverName, databaseName)
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Add("Authorization", "Bearer " + ah.currentToken().AccessToken)
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error on post %s\n", err.Error())
		panic(err)
	}

	fmt.Printf("status code is %d\n", resp.StatusCode)

	// if status begins with 4.... assume failure.
	if strings.HasPrefix( resp.Status, "4") {
		return errors.New("unable to start backup")
	}

	return nil
}
