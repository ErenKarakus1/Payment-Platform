package mail

import (
	"fmt"
	"net/smtp"
)

type Sender struct {
	host     string
	port     string
	username string
	password string
}

func NewSender(
	host string,
	port string,
	username string,
	password string,
) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (s *Sender) Send(to string, subject string, body string) error {
	auth := smtp.PlainAuth(
		"",
		s.username,
		s.password,
		s.host,
	)

	message := []byte(
		"From: Notification Service <" + s.username + ">\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body,
	)

	address := fmt.Sprintf("%s:%s", s.host, s.port)

	return smtp.SendMail(
		address,
		auth,
		s.username,
		[]string{to},
		message,
	)
}
