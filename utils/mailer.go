package utils

import (
	"crypto/tls"
	"fmt"
	"html"
	"net/smtp"
	"strings"

	"gopkg.in/gomail.v2"
)

type SMTP struct {
	Port                       int
	Server, Username, Password string
}

type EMailMessage struct {
	To, From, Replyto, Subject,
	Content, Attachment string
	Cc, Bcc []string
}

type Mailer struct {
	SMTP
	EMailMessageList []EMailMessage
}

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

func (this *Mailer) SendMail() (sMessage string) {

	sMessage = ""
	goMailMsg := gomail.NewMessage()
	goDialer := gomail.NewDialer(this.SMTP.Server, this.SMTP.Port, this.SMTP.Username, this.SMTP.Password)
	goDialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	goDialer.Auth = unencryptedAuth{
		smtp.PlainAuth(
			"",
			this.SMTP.Username,
			this.SMTP.Password,
			this.SMTP.Server,
		),
	}

	goSender, err := goDialer.Dial()
	if err != nil {
		sMessage = strings.Replace(err.Error(), "\n", "", -1)
		sMessage = html.EscapeString(sMessage)
		return
	}

	for _, MailMsg := range this.EMailMessageList {

		goMailMsg.SetHeader("To", MailMsg.To)
		goMailMsg.SetHeader("From", MailMsg.From)

		for _, cc := range MailMsg.Cc {
			goMailMsg.SetHeader("Cc", cc)
		}

		for _, bcc := range MailMsg.Bcc {
			goMailMsg.SetHeader("Bcc", bcc)
		}

		if MailMsg.Replyto != "" {
			goMailMsg.SetHeader("Reply-to", MailMsg.Replyto)
		}

		goMailMsg.SetHeader("Subject", MailMsg.Subject)
		goMailMsg.SetBody("text/html", MailMsg.Content)
		if MailMsg.Attachment != "" {
			goMailMsg.Attach(MailMsg.Attachment)
		}

		if err := gomail.Send(goSender, goMailMsg); err != nil {
			sMessage = strings.Replace(err.Error(), "\n", "", -1)
			sMessage = html.EscapeString(sMessage)
		}
		goMailMsg.Reset()
	}

	return
}

func (this *Mailer) CheckMail() (sMessage string) {

	sMessage = ""

	if this.SMTP.Port == 0 {
		sMessage += "SMTP.Port is blank <br>"
	}

	if this.SMTP.Server == "" {
		sMessage += "SMTP.Server is blank <br>"
	}

	if this.SMTP.Username == "" {
		sMessage += "SMTP.Username is blank <br>"
	}

	if this.SMTP.Password == "" {
		sMessage += "SMTP.Password is blank <br>"
	}

	for Key, MailMsg := range this.EMailMessageList {

		if MailMsg.To == "" {
			sMessage += fmt.Sprintf("MailMsg %d To is blank <br>", Key)
		}

		if MailMsg.From == "" {
			sMessage += fmt.Sprintf("MailMsg %d From is blank <br>", Key)
		}

		if MailMsg.Subject == "" {
			sMessage += fmt.Sprintf("MailMsg %d Subject is blank <br>", Key)
		}

		if MailMsg.Content == "" {
			sMessage += fmt.Sprintf("MailMsg %d Content is blank <br>", Key)
		}
	}

	return
}
