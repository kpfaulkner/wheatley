package messagehandlers

// MessageHandler takes a message, parses it and determines what it should do.
// Returns the text (if any) that should be returned to the user.
type MessageHandler interface {
	
	// ParseMessage takes a message, determines what to do
	// return the text that should go to the user.
	ParseMessage( msg string, user string) (string, error)
	
}
