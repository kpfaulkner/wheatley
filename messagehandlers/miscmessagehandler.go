package messagehandlers

import (
	"errors"
	"regexp"
)

// MiscMessageHandler misc rubbish
type MiscMessageHandler struct {
}

func NewMiscMessageHandler() *MiscMessageHandler {
	handler := MiscMessageHandler{}
	return &handler
}

// ParseMessage takes a message, determines what to do
// return the text that should go to the user.
func (sg *MiscMessageHandler) ParseMessage(msg string, user string) (MessageResponse, error) {

	haskellRegex := regexp.MustCompile(`.*haskell.*`)
	signOfLifeRegex := regexp.MustCompile(`hello`)
	soundOffRegex := regexp.MustCompile(`sound off`)

	switch {
	case signOfLifeRegex.MatchString(msg):
		return NewTextMessageResponse("alive and well"), nil

	case haskellRegex.MatchString(msg):
		return NewTextMessageResponse(".... haskell... don't get me started!!!"), nil

	case soundOffRegex.MatchString(msg):
		return NewTextMessageResponse("MiscMessageHandler reporting for duty "), nil
	}

	return NewTextMessageResponse(""), errors.New("No match")
}
