package messagehandlers

import (
	"errors"
	"fmt"
	"github.com/kpfaulkner/wheatley/config"
	"regexp"
)

// AzureStorageMessageHandler checks for SG messages.
type AzureStorageMessageHandler struct {
	Configuration config.Config

	// regexp generated from config
	regexpMap map[*regexp.Regexp]string
}

func NewAzureStorageMessageHandler(configFilePath string) *AzureStorageMessageHandler {
	asHandler := AzureStorageMessageHandler{}

	if configFilePath != "" {
		asHandler.Configuration = config.LoadConfiguration(configFilePath)

		// generate regexs for azure storage.
		asHandler.regexpMap = generateRegexps(asHandler.Configuration.AzureQueries)
	}

	return &asHandler
}

// generateRegexps generates a map between compiled regex's and the query string that should be associated when that regex is matched.
func generateRegexps(azureQueries map[string]string) map[*regexp.Regexp]string {

	compiledQueries := make(map[*regexp.Regexp]string)

	for k, v := range azureQueries {
		fmt.Printf("key %s, val %s\n", k, v)
		compiledQuery := regexp.MustCompile(k)
		compiledQueries[compiledQuery] = v
	}

	return compiledQueries
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (sg *AzureStorageMessageHandler) ParseMessage(msg string, user string) (MessageResponse, error) {

	fmt.Printf("msg is %s", msg)

	return NewTextMessageResponse(""), errors.New("No match")

}
