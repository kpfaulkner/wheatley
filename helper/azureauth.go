package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AzureAuthToken struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`

	ExpiresOnTime  time.Time // converted from string above. Don't want to make custom unmarshaller.
}

type AzureAuth struct {
	tenantID string
	clientID string
	clientSecret string

	// current token. Check expiry time before trying to use it!
	currentToken AzureAuthToken
}

func NewAzureAuth(tenantID string, clientID string, clientSecret string) *AzureAuth {
	aa := AzureAuth{}
	aa.tenantID = tenantID
	aa.clientSecret = clientSecret
	aa.clientID = clientID

	return &aa
}

func (aa *AzureAuth) CurrentToken()  AzureAuthToken {
  return aa.currentToken
}

// refreshToken checks the token, if it's going to expire in the next 30 seconds then it will refresh it.
func (aa *AzureAuth) RefreshToken() error {
	n := time.Now().UTC().Add( 5 * time.Minute)
	fmt.Printf("time check is %s\n", n)
  if aa.currentToken.AccessToken != "" {
	  fmt.Printf("existing token time %s\n", aa.currentToken.ExpiresOnTime.UTC())
  }

  if aa.currentToken.AccessToken == "" || aa.currentToken.ExpiresOnTime.UTC().Before(time.Now().UTC()) {
	  token, err := generateAuthHeader(aa.tenantID, aa.clientID, aa.clientSecret)
  	if err != nil {
  		fmt.Printf("error while generating auth token! %s\n", err.Error())
  		return err
	  }
  	fmt.Printf("generated token with expiry %s\n", token.ExpiresOnTime)
	  aa.currentToken = *token
  }
  return nil
}

// see http://devchat.live/en/2017/02/27/access-metrics-using-azure-monitor-rest-api/
// URL is https://login.microsoftonline.com/<tenantID>/oauth2/token
func generateAuthHeader(tenantID string, clientID string, clientSecret string) (*AzureAuthToken, error) {
	urlTemplate := "https://login.microsoftonline.com/%s/oauth2/token"
	bodyTemplate := "grant_type=client_credentials&resource=https://management.core.windows.net/&client_id=%s&client_secret=%s"
	url := fmt.Sprintf(urlTemplate, tenantID)
	body := fmt.Sprintf(bodyTemplate, clientID, clientSecret)
	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := string(respBody)
	fmt.Printf("resp is %s\n", s)

	var auth AzureAuthToken
	err = json.Unmarshal( respBody, &auth)
	if err != nil {
		return nil, err
	}

	i64,_ := strconv.Atoi(auth.ExpiresOn)
	auth.ExpiresOnTime = time.Unix(int64(i64), 0)
	return &auth, nil
}

