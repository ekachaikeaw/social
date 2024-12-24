package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"time"

	gomail "gopkg.in/mail.v2"
)

type mailTrapClient struct {
	fromEmail string
	apiKey    string
}

func NewMailTrapClient(fromEmail, apiKey string) (mailTrapClient, error) {
	if apiKey == "" {
		return mailTrapClient{}, errors.New("api key is required")
	}

	return mailTrapClient{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}, nil
}

func (m mailTrapClient) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	// Template parsing and building
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return -1, err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())

	message.AddAlternative("text/html", body.String())
	
	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)

	var retryErr error
	for i := 0; i < maxRetries; i++ {
		retryErr := dialer.DialAndSend(message)
		if retryErr != nil {
			// exponential back off
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		return 200, nil
	}

	return -1, fmt.Errorf("fail to send email after %d attempt, error: %v", maxRetries, retryErr)
}
