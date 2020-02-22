package messagehandlers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kpfaulkner/wheatley/helper"
	"regexp"
	"sort"
	"strings"
)

// ServiceBusMessageHandler snooping
type ServiceBusMessageHandler struct {
	sbHelper *helper.ServiceBusHelper
}

func NewServiceBusMessageHandler() *ServiceBusMessageHandler {
	handler := ServiceBusMessageHandler{}
	handler.sbHelper = helper.NewServiceBusHelper()
	return &handler
}

func parseServiceBusResults(res *helper.ServiceBusResponse, onlyActive bool) string {

	var buffer bytes.Buffer
	if res != nil {
		fmt.Println(buffer.String())

		allKeys := []string{}
		for k := range res.AllCountDetails {
			allKeys = append(allKeys, k)
		}
		sort.Strings(allKeys)
		for _, k := range allKeys {
			v := res.AllCountDetails[k]

			if !onlyActive || (onlyActive && *v.ActiveMessageCount > 0) {
				buffer.WriteString(fmt.Sprintf("%s Active : %d\n", k, *v.ActiveMessageCount))
				buffer.WriteString(fmt.Sprintf("%s Dead : %d\n", k, *v.DeadLetterMessageCount))
			}
		}
	}
	if buffer.Len() == 0 {
		return "no results"
	}

	return buffer.String()
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (sb *ServiceBusMessageHandler) ParseMessage(msg string, user string) (string, error) {

	soundOffRegex := regexp.MustCompile(`sound off`)
	checkQueueMessageCountRegex := regexp.MustCompile(`sb check queue(.*)`)
	checkTopicMessageCountRegex := regexp.MustCompile(`sb check topic(.*)`)

	checkActiveQueueMessageCountRegex := regexp.MustCompile(`sb check active queue(.*)`)
	checkActiveTopicMessageCountRegex := regexp.MustCompile(`sb check active topic(.*)`)

	switch {
	case soundOffRegex.MatchString(msg):
		return "ServiceBusMessageHandler reporting for duty", nil

	case checkQueueMessageCountRegex.MatchString(msg):
		fmt.Printf("\n\nQUEUE CHECK\n\n")
		returnString := "no results"
		res := checkQueueMessageCountRegex.FindStringSubmatch(msg)
		if res != nil && len(res) > 0 {
			results, _ := sb.sbHelper.CheckQueue(strings.TrimSpace(res[1]))
			returnString = parseServiceBusResults(&results, false)
		}

		return returnString, nil

	case checkTopicMessageCountRegex.MatchString(msg):
		fmt.Printf("\n\nTOPIC CHECK\n\n")

		returnString := "no results"
		res := checkTopicMessageCountRegex.FindStringSubmatch(msg)
		if res != nil && len(res) > 0 {
			results, _ := sb.sbHelper.CheckTopic(strings.TrimSpace(res[1]))
			returnString = parseServiceBusResults(&results, false)
		}

		return returnString, nil

	case checkActiveQueueMessageCountRegex.MatchString(msg):
		fmt.Printf("\n\nACTIVE QUEUE CHECK\n\n")
		returnString := "no results"
		res := checkActiveQueueMessageCountRegex.FindStringSubmatch(msg)
		if res != nil && len(res) > 0 {
			results, _ := sb.sbHelper.CheckQueue(strings.TrimSpace(res[1]))
			returnString = parseServiceBusResults(&results, true)
		}

		return returnString, nil

	case checkActiveTopicMessageCountRegex.MatchString(msg):
		fmt.Printf("\n\nnACTIVE TOPIC CHECK\n\n")

		returnString := "no results"
		res := checkActiveTopicMessageCountRegex.FindStringSubmatch(msg)
		if res != nil && len(res) > 0 {
			results, _ := sb.sbHelper.CheckTopic(strings.TrimSpace(res[1]))
			returnString = parseServiceBusResults(&results, true)
		}

		return returnString, nil

	}

	return "", errors.New("No match")
}
