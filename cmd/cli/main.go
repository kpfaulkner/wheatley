package main

import (
	"github.com/kpfaulkner/wheatley/messagehandlers"
	"github.com/slack-go/slack"
	"log"
	"os"
)

func getMessageHandlers() []messagehandlers.MessageHandler {
	misc := messagehandlers.NewMiscMessageHandler()
	sh := messagehandlers.NewServerStatusMessageHandler()
	ah := messagehandlers.NewAzureStatusMessageHandler()
	dbh := messagehandlers.NewDatabaseBackupMessageHandler()
	ach := messagehandlers.NewAzureCostMessageHandler()

	handlers := []messagehandlers.MessageHandler{misc, sh, ah, dbh, ach}
	return handlers
}

func main() {

	handlers := getMessageHandlers()

	slackKey := os.Getenv("SLACK_KEY")
	api := slack.New(slackKey)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.OptionLog(logger)
	slack.OptionDebug(true)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:

			for _, handler := range handlers {
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
