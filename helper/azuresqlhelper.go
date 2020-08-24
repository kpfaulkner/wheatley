package helper

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type AzureSQLHelper struct {
	azureAuth              *AzureAuth
	exportSubscriptionID   string
	importSubscriptionID   string
	tenantID               string
	clientID               string
	clientSecret           string
	sqlExportAdminLogin    string
	sqlExportAdminPassword string
	sqlImportAdminLogin    string
	sqlImportAdminPassword string
	storageKey             string
	storageURL             string
	exportSqlRgName        string
	importSqlRgName        string
	importStorageKey       string
}

func NewAzureSQLHelper(importSubscriptionID string, exportSubscriptionID string, tenantID string, clientID string, clientSecret string, sqlExportAdminLogin string, sqlExportAdminPassword string, sqlImportAdminLogin string, sqlImportAdminPassword string, storageKey string, storageURL string, exportSqlRgName string, importSqlRgName string, importStorageKey string) *AzureSQLHelper {
	ah := AzureSQLHelper{}
	ah.azureAuth = NewAzureAuth(tenantID, clientID, clientSecret)
	ah.exportSubscriptionID = exportSubscriptionID
	ah.importSubscriptionID = importSubscriptionID
	ah.tenantID = tenantID
	ah.clientID = clientID
	ah.clientSecret = clientSecret
	ah.sqlExportAdminLogin = sqlExportAdminLogin
	ah.sqlExportAdminPassword = sqlExportAdminPassword
	ah.sqlImportAdminLogin = sqlImportAdminLogin
	ah.sqlImportAdminPassword = sqlImportAdminPassword
	ah.storageKey = storageKey
	ah.storageURL = storageURL
	ah.exportSqlRgName = exportSqlRgName
	ah.importSqlRgName = importSqlRgName
	ah.importStorageKey = importStorageKey
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

func generateImportURL(subscriptionID string, rgName string, serverName string, databaseName string) string {
	template := "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Sql/servers/%s/databases/%s/extensions/import?api-version=2014-04-01"
	url := fmt.Sprintf(template, subscriptionID, rgName, serverName,databaseName)
	return url
}

func generateImportBody(adminLogin string, adminLoginPassword string, storageKey string,storageUri string) string {
	/*template2 := `{administratorLogin: "%s",administratorLoginPassword: "%s",storageKey: "%s",storageKeyType: "%s",storageUri: "%s",
								databasename:"%s", edition:"%s",tier:"Standard", skuName:"S2", location:"southcentralus", serviceObjectiveName:"%s",maxSizeBytes:"%d"}` */
	template := `{"properties": {"storageKeyType": "StorageAccessKey", "storageKey": "%s", "storageUri": "%s", "administratorLogin": "%s", "administratorLoginPassword": "%s", "operationMode": "Import"}}`
	body := fmt.Sprintf(template, storageKey,storageUri,adminLogin, adminLoginPassword)
	return body
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

// StartDBExport starts an export of an Azure DB to blob storage.
// https://docs.microsoft.com/en-us/rest/api/sql/databases%20-%20import%20export/export
func (ah *AzureSQLHelper) StartDBExport(serverName string, databaseName string, backupFileName string) error {

	// refresh all the tokens!!!
	err := ah.refreshToken()
	if err != nil {
		return err
	}

	storageURI := fmt.Sprintf("%s/%s", ah.storageURL, backupFileName)
	body := generateExportBody(ah.sqlExportAdminLogin, ah.sqlExportAdminPassword, ah.storageKey, "SharedAccessKey", storageURI)
	url := generateExportURL(ah.exportSubscriptionID, ah.exportSqlRgName, serverName, databaseName)
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
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

// StartDBImport starts to import from a blob backup file to a specific DB server and dbname
// keep to a default size for now.
// https://docs.microsoft.com/en-us/rest/api/sql/databases%20-%20import%20export/import
func (ah *AzureSQLHelper) StartDBImport(importServerName string, databaseName string, backupBlobName string) error {

	// refresh all the tokens!!!
	err := ah.refreshToken()
	if err != nil {
		return err
	}

	storageURI := fmt.Sprintf("%s/%s", ah.storageURL, backupBlobName)
	body := generateImportBody(ah.sqlImportAdminLogin, ah.sqlImportAdminPassword, ah.importStorageKey, storageURI)
	url := generateImportURL(ah.importSubscriptionID, ah.importSqlRgName, importServerName, databaseName)
	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, strings.NewReader(body))
	req.Header.Add("Authorization", "Bearer "+ah.currentToken().AccessToken)
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error on put %s\n", err.Error())
		panic(err)
	}

	fmt.Printf("status code is %d\n", resp.StatusCode)
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("body is %s\n", string(b))

	// if status begins with 4.... assume failure.
	if strings.HasPrefix(resp.Status, "4") {
		return errors.New("unable to start backup")
	}

	return nil
}

// CreateDB Creates DB
// https://docs.microsoft.com/en-us/rest/api/sql/databases/createorupdate#code-try-0
func (ah *AzureSQLHelper) CreateDB(importServerName string, databaseName string) error {

	// refresh all the tokens!!!
	err := ah.refreshToken()
	if err != nil {
		return err
	}

	body := generateCreateDBBody("ah.sqlImportAdminLogin, ah.sqlImportAdminPassword, ah.importStorageKey, storageURI")
	url := generateCreateDBURL(ah.importSubscriptionID, ah.importSqlRgName, importServerName, databaseName)
	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, strings.NewReader(body))
	req.Header.Add("Authorization", "Bearer "+ah.currentToken().AccessToken)
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error on put %s\n", err.Error())
		panic(err)
	}

	fmt.Printf("status code is %d\n", resp.StatusCode)
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("body is %s\n", string(b))

	// if status begins with 4.... assume failure.
	if strings.HasPrefix(resp.Status, "4") {
		return errors.New("unable to start backup")
	}

	return nil
}


func generateCreateDBURL(subscriptionID string, rgName string, serverName string, databaseName string) string {
	template := "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Sql/servers/%s/databases/%s?api-version=2017-10-01-preview"
	url := fmt.Sprintf(template, subscriptionID, rgName, serverName,databaseName)
	return url
}

func generateCreateDBBody(dbSku string) string {
	template := ` {"location": "southcentralus", "sku": {"name": "%s"}}`
	body := fmt.Sprintf(template, dbSku)
	return body
}
