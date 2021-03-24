package messagehandlers

import (
	"errors"
	"fmt"
	"github.com/kpfaulkner/wheatley/helper"
	"regexp"
)

// SendgridMessageHandler checks for SG messages.
type SendgridMessageHandler struct {
	Handler *helper.SendgridHandler
}

func NewSendgridMessageHandler() *SendgridMessageHandler {
	sgHandler := SendgridMessageHandler{}
	sgHandler.Handler = helper.NewSendgridHandler()
	return &sgHandler
}

// checkEmail checks the email for various problems stored in Sendgrid
func (sg *SendgridMessageHandler) checkEmail(emailAddr string) (string, error) {
	msg := ""
	result, _ := sg.Handler.CheckBlock(emailAddr)
	if len(result) > 0 {
		msg = fmt.Sprintf("blocked due to %s\n", result[0].Reason)
	}

	result, _ = sg.Handler.CheckBounce(emailAddr)
	if len(result) > 0 {
		msg = fmt.Sprintf("%sbounced due to %s\n", msg, result[0].Reason)
	}

	result, _ = sg.Handler.CheckInvalid(emailAddr)
	if len(result) > 0 {
		msg = fmt.Sprintf("%sinvalid due to %s\n", msg, result[0].Reason)
	}

	r, _ := sg.Handler.CheckSpam(emailAddr)
	if len(r) > 0 {
		msg = fmt.Sprintf("marked as spam")
	}

	if msg == "" {
		msg = "all good"
	}
	return msg, nil
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (sg *SendgridMessageHandler) ParseMessage(msg string, user string) (MessageResponse, error) {

	checkEmailRegex := regexp.MustCompile(`check <mailto:(.*)\|(.*)> email`)
	checkDebounceRegex := regexp.MustCompile(`debounce <mailto:(.*)\|(.*)> email`)
	checkDeblockRegex := regexp.MustCompile(`unblock <mailto:(.*)\|(.*)> email`)
	checkRemoveSpamRegex := regexp.MustCompile(`despam <mailto:(.*)\|(.*)> email`)
	checkRemoveInvalidRegex := regexp.MustCompile(`remove invalid <mailto:(.*)\|(.*)> email`)
	soundOffRegex := regexp.MustCompile(`sound off`)

	switch {
	case checkEmailRegex.MatchString(msg):
		res := checkEmailRegex.FindStringSubmatch(msg)
		if res != nil {
			msg, _ := sg.checkEmail(res[2])
			return NewTextMessageResponse(msg), nil
		}

	case checkDebounceRegex.MatchString(msg):
		res := checkDebounceRegex.FindStringSubmatch(msg)
		if res != nil {
			err := sg.Handler.DeleteBounce(res[2])
			if err != nil {
				return NewTextMessageResponse("unable to debounce"), nil
			}

			return NewTextMessageResponse("debounced"), nil
		}

	case checkDeblockRegex.MatchString(msg):
		res := checkDeblockRegex.FindStringSubmatch(msg)
		if res != nil {
			err := sg.Handler.DeleteBlock(res[2])
			if err != nil {
				return NewTextMessageResponse("unable to deblock"), nil
			}

			return NewTextMessageResponse("deblocked"), nil
		}

	case checkRemoveSpamRegex.MatchString(msg):
		res := checkRemoveSpamRegex.FindStringSubmatch(msg)
		if res != nil {
			err := sg.Handler.DeleteSpam(res[2])
			if err != nil {
				return NewTextMessageResponse("unable to de spam..."), nil
			}

			return NewTextMessageResponse("despammed"), nil
		}

	case checkRemoveInvalidRegex.MatchString(msg):
		res := checkRemoveInvalidRegex.FindStringSubmatch(msg)
		if res != nil {
			err := sg.Handler.DeleteSpam(res[2])
			if err != nil {
				return NewTextMessageResponse("unable to remove invalid.."), nil
			}

			return NewTextMessageResponse("de-invalidateded ;)"), nil
		}

	case soundOffRegex.MatchString(msg):
		return NewTextMessageResponse("SendgridMessageHandler reporting for duty"), nil

	}
	return NewTextMessageResponse(""), errors.New("No match")

}
