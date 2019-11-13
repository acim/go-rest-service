package mail

import "context"

// Mail contains all data needed to send an e-mail.
type Mail struct {
	From    string
	Subject string
	Text    string
	To      []string
}

// Response contains data returned by mail service.
type Response struct {
	Message string
	ID      string
}

// Sender is an interface to send e-mails.
type Sender interface {
	Send(context.Context, *Mail) (*Response, error)
}
