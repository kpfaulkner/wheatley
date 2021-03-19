package messagehandlers

import (
	"fmt"
	"github.com/slack-go/slack"
	"strings"
	"time"
)

const (
	TextMessageType int = 1 // just returning a text message.
	FileMessageType int = 2 // contents of message should be u
)

type MessageResponse interface {
	GetMessageResponseType() int
}

type BaseMessageResponse struct {
	messageType int
}

func (bm BaseMessageResponse) GetMessageResponseType() int {
  return bm.messageType
}

type TextMessageResponse struct {
	BaseMessageResponse
	Message string
}

func NewTextMessageResponse( msg string) TextMessageResponse {
	tm := TextMessageResponse{}
	tm.messageType = TextMessageType
	tm.Message = msg
	return tm
}

type FileMessageResponse struct {
	BaseMessageResponse

	Title string
	FileName string
	Contents []byte
	FileType string // default to "txt" ??
}

func NewFileMessageResponse( fileName string, title string, contents []byte) FileMessageResponse {
	fm := FileMessageResponse{}
	fm.messageType = FileMessageType
	fm.Title = title
	fm.FileName = fileName
	fm.FileType = "txt"  // just default for now.
	fm.Contents = contents
	return fm
}



// MessageHandler takes a message, parses it and determines what it should do.
// Returns the MessageResponse (which could be text, or a file?) that should be returned to the user.
type MessageHandler interface {
	
	// ParseMessage takes a message, determines what to do
	// return the text that should go to the user.
	ParseMessage( msg string, user string) (MessageResponse, error)
}

func ProcessMessageResponse(msg MessageResponse, channel string, api *slack.Client, rtm *slack.RTM) error {

	// could just use type assertions, but will stick with this for now.
	switch msg.GetMessageResponseType() {
	case TextMessageType:
		textMessage := msg.(TextMessageResponse)
		sp := strings.Split(textMessage.Message, "\n")

		// slows things down.. but stops being blocked by Slack.
		// If we send responses > 5000 bytes, then Slack will also block it.
		for _, line := range sp {
			rtm.SendMessage(rtm.NewOutgoingMessage(line,channel))
			<-time.After(2 * time.Second)
		}

	case FileMessageType:
		fileMessage := msg.(FileMessageResponse)

		params := slack.FileUploadParameters{
			Title: fileMessage.Title,
			Filetype: fileMessage.FileType,
			File: fileMessage.FileName,
			Content:  string(fileMessage.Contents),   // should fileMesage.Contents just be string to begin with?
		}

		api.UploadFile()
		file, err := api.UploadFile(params)
		if err != nil {
			fmt.Printf("%s\n", err)
			return err
		}
		fmt.Printf("Name: %s, URL: %s\n", file.Name, file.URLPrivateDownload)
		rtm.SendMessage(rtm.NewOutgoingMessage( fmt.Sprintf("file %s is at %s", file.Name, file.URLPrivateDownload), channel))
		rtm.SendMessage(rtm.NewOutgoingMessage( fmt.Sprintf("file %s is at %s", file.Name, file.Permalink), channel))
		rtm.SendMessage(rtm.NewOutgoingMessage( fmt.Sprintf("file %s is at %s", file.Name, file.PermalinkPublic), channel))

	}
	return nil
}

