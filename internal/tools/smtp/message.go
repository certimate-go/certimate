package smtp

import (
	"github.com/wneessen/go-mail"
)

type Message = mail.Msg

func NewMessage() *Message {
	return mail.NewMsg()
}

type MIMEType = mail.ContentType

const (
	MIMETypeTextHTML  MIMEType = mail.TypeTextHTML
	MIMETypeTextPlain MIMEType = mail.TypeTextPlain
)
