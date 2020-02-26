package main

import (
	"github.com/kpfaulkner/wheatley/messagehandlers"
	"github.com/slack-go/slack"
	"log"
	"os"
)

// when we want to add a new handler. Add it to the slice here... then all is fine!
func getMessageHandlers() []interface{} {
  misc := messagehandlers.NewMiscMessageHandler()
  sh := messagehandlers.NewServerStatusMessageHandler()
  ah := messagehandlers.NewAzureStatusMessageHandler()
  dbh := messagehandlers.NewDatabaseBackupMessageHandler()

	handlerArray := []messagehandlers.MessageHandler{misc, sh, ah, dbh}
	handlers := make([]interface{}, len(handlerArray))
	for i, h := range handlerArray {
		handlers[i] = h
	}
	return handlers
}

func main() {

	handlers := getMessageHandlers()

	slackKey := os.Getenv("SLACK_KEY")
	slackKey = "xoxb-6737872145-dN2E88nyrAWoot486wQmqBZD"
	api := slack.New(slackKey)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.OptionLog(logger)
	slack.OptionDebug(true)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:

			for _, handlerInterface := range handlers {
				handler, _ := handlerInterface.(messagehandlers.MessageHandler)
				go func(text string, user string, grRtm *slack.RTM, h messagehandlers.MessageHandler) {
					u, err := api.GetUserInfo(user)
					if err != nil {
						// error.....  die a mysterious death..
						return
					}

					msg, _ := h.ParseMessage(text, u.Name)
					grRtm.SendMessage(grRtm.NewOutgoingMessage(msg, ev.Channel))
				}(ev.Text, ev.User, rtm, handler)

			}

		default:
		}
	}
}
