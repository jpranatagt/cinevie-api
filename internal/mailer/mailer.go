package mailer

import (
  "bytes"
  "embed"
  "html/template"
  "time"

  "github.com/go-mail/mail/v2"
)

//go:embed templates/*
var templateFS embed.FS

// mail.Dialer used to connect to a SMTP server
// and the sender information for the emails
type Mailer struct {
  dialer *mail.Dialer
  sender string
}

func New(host string, port int, username, password, sender string) Mailer {
  // initialize a new mail.Dialer instance with the given SMTP server settings
  // also configure 5 seconds timeout whenever email being sent
  dialer := mail.NewDialer(host, port, username, password)
  dialer.Timeout = 5 * time.Second

  // return Mailer instance containing the dialer and sender information
  return Mailer {
    dialer: dialer,
    sender: sender,
  }
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
  // parse the required template file
  tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
  if err != nil {
    return err
  }

  subject := new(bytes.Buffer)
  err = tmpl.ExecuteTemplate(subject, "subject", data)
  if err != nil {
    return err
  }

  plainBody := new(bytes.Buffer)
  err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
  if err != nil {
    return err
  }

  htmlBody := new(bytes.Buffer)
  err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
  if err != nil {
    return err
  }

  msg := mail.NewMessage()
  msg.SetHeader("To", recipient)
  msg.SetHeader("From", m.sender)
  msg.SetHeader("Subject", subject.String())
  msg.SetBody("text/plain", plainBody.String())
  msg.AddAlternative("text/html", htmlBody.String())

  // opens a connection to the SMTP server
	// retrying email send attempts
  for i:= 0; i < 3; i++ {
    err = m.dialer.DialAndSend(msg)
    // return nil if email send is succeed or err is nil
    if nil == err {
      return nil
    }

	// if didn't work sleep for 1s before retrying
    time.Sleep(1 * time.Second)
	}

  return err
}
