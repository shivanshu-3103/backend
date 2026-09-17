package services

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailService interface {
	SendEmail(to []string, subject string, body string) error
}

type emailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService() EmailService {
	return &emailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (s *emailService) SendEmail(to []string, subject string, body string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.from, to[0], subject, body))

	err := smtp.SendMail(addr, auth, s.from, to, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
