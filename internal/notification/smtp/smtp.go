package smtp

import (
	"fmt"
	"net/smtp"
)

func Notify(email Email) error {
	// message string assembly
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s", email.To, email.From, email.Subject, email.Message)

	// TODO option to set nil for unauthenticated
	auth := smtp.PlainAuth("", email.User, email.Password, email.Host)

	// Send the email
	err := smtp.SendMail(email.Host+":"+email.Port, auth, email.From, []string{email.To}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}
