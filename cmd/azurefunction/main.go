package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kpfaulkner/wheatley/messagehandlers"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"log"
	"net/http"
	"os"
)

type SlackServer struct {
	slackApi *slack.Client
	token string
	verificationToken string
	handlers []interface{}
}

func NewSlackServer(token string, verificationToken string) SlackServer {
  ss := SlackServer{}
  ss.token = token
  ss.verificationToken = verificationToken
  ss.slackApi = slack.New(token)
	ss.handlers = azureFunctionGetMessageHandlers()

  return ss
}

func (s *SlackServer) slackHttp( w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: s.verificationToken}))
	if e != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// used for event API verification that we're a legit bot for
	// this slack cct.
	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	 	w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	}

	// dealing with actual messages. At the moment we're just after
	// basic messaging.... no fancy events subscribed to already.
	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent


		//w.WriteHeader(http.StatusInternalServerError)
		switch ev := innerEvent.Data.(type) {

		case *slackevents.MessageEvent:

			// return 200 immediately... according to https://api.slack.com/events-api#prepare
			// otherwise if we dont return in 3seconds the delivery is considered to have failed and we'll get another
			// message. So can return 200 immediately but then the code that processes the messages can
			// return their results later on
			w.WriteHeader(http.StatusOK)

			// loop through all handlers and reply.
			for _, handlerInterface := range s.handlers {
				handler, _ := handlerInterface.(messagehandlers.MessageHandler)

				// not a GO channel... but just using slack-go terminology
				go func(text string, user string, channelID string, h messagehandlers.MessageHandler) {
					u , err := s.slackApi.GetUserInfo(user)
					if err != nil {
						// error.....  die a mysterious death..
						return
					}
					msg, _ := h.ParseMessage(text, u.Name)
					s.slackApi.PostMessage(channelID, slack.MsgOptionText( msg, false))
				}(ev.Text, ev.User, ev.Channel, handler)
			}
		}
	}
}


func (s *SlackServer) routes() {
  http.HandleFunc("/events-endpoint", s.slackHttp)
}

func (s *SlackServer) run() {
	port, exists := os.LookupEnv("FUNCTIONS_HTTPWORKER_PORT")
	if !exists {
		port = "3000"
	}
	fmt.Printf("port used is %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port,nil))
}

func azureFunctionGetMessageHandlers() []interface{} {
	misc := messagehandlers.NewMiscMessageHandler()
	//sh := messagehandlers.NewServerStatusMessageHandler()
	//ah := messagehandlers.NewAzureStatusMessageHandler()
	//dbh := messagehandlers.NewDatabaseBackupMessageHandler()

	handlerArray := []messagehandlers.MessageHandler{misc}
	handlers := make([]interface{}, len(handlerArray))
	for i, h := range handlerArray {
		handlers[i] = h
	}
	return handlers
}

// Due to running as an Azure Function (no websockets for us)
// this version will be using the events and web APIs.
func main() {

	fmt.Printf("hello world :)\n")
	slackKey := os.Getenv("SLACK_KEY")
	verificationToken := os.Getenv("SLACK_VERIFICATION_TOKEN")
	s := NewSlackServer(slackKey, verificationToken)
	s.routes()
	s.run()
}
