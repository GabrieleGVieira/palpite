package email

import "context"

type EmailSender interface {
	Send(ctx context.Context, message EmailMessage) error
}

type EmailMessage struct {
	To      []string
	Subject string
	HTML    string
	Text    string
}

type Sender = EmailSender
type Message = EmailMessage
