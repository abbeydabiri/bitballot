package utils

import (
	"fmt"
	"log"
	
	// "bytes"
	// "html/template"

	// "bitballot/config"
	// "bitballot/database"
)

//Email structure of mail
type Email struct {
	From, FromName, Replyto,
	To, Subject, Message string
	BCC, CC []string
}

func SendEmail(email Email, mySMTP SMTP) string {

	if email.To == "" || email.Subject == "" || email.Message == "" {
		return "either To, Subject or Message fields is blank/empty"
	}

	if mySMTP.Port == 0 || mySMTP.Server == "" || mySMTP.Username == "" || mySMTP.Password == "" {
		return "either Port, Server or Username or Password fields is/are blank/empty"
	}

	if email.From == "" {
		email.From = mySMTP.Username
	}

	if email.FromName == "" {
		email.FromName = mySMTP.Username
	}

	emailSender := fmt.Sprintf("%s <%s>", email.FromName, email.From)

	var messageList []EMailMessage
	messageList = append(messageList,
		EMailMessage{
			Attachment: "",
			To:         email.To,
			From:       emailSender,
			Cc:         email.CC, Bcc: email.BCC, Replyto: email.Replyto,
			Subject: email.Subject,
			Content: email.Message,
		})
	mailer := Mailer{mySMTP, messageList}

	//log.Printf(" - - -- - - - -- - -- - --- - \n Mail:  %v \n\n", mailer)
	// return ""

	sMessage := mailer.CheckMail()
	if len(sMessage) > 0 {
		log.Printf(sMessage)
		return sMessage
	}

	sMessage = mailer.SendMail()
	if len(sMessage) > 0 {
		sMessage = fmt.Sprintf("To: %v\nFrom: %v\nReply: %v\nSubject: %v\n %v",
			email.To, emailSender, email.Replyto, email.Subject, sMessage)
		log.Printf(sMessage)
		log.Printf(email.Message)
		return sMessage
	}

	return sMessage
}

