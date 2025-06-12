package utils

import (
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "yammanuru.tejaswini@vegrow.in")
	fmt.Println("Sending email to:", to)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "yammanuru.tejaswini@vegrow.in", "ldqkbfsaipsielav")

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Println("Email sent successfully to:", to)
	return nil
}
