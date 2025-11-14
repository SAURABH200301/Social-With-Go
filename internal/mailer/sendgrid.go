package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	sendGridGo "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendGridGo.Client
}

func NewSendGridMailer(fromEmail, apiKey string) *SendGridMailer {
	client := sendGridGo.NewSendClient(apiKey)
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (sg *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, sg.fromEmail)
	to := mail.NewEmail(username, email)

	//template parsing and building

	tmpl, err := template.ParseFiles("templates/" + templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %v", templateFile, err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("failed to execute subject template for %s: %v", email, err)
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return fmt.Errorf("failed to execute body template for %s: %v", email, err)
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	//sandbox mode == dev
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})
	var retryErr error
	for i := 0; i < maxRetries; i++ {
		response, retryErr := sg.client.Send(message)
		if retryErr != nil {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Printf("email sent to %s with status code %d", email, response.StatusCode)
		return nil
	}
	return fmt.Errorf("failed to send email to %s after %d attempts: %v", email, maxRetries, retryErr)
}
