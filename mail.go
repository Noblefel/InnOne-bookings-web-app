package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			sendMessage(msg)
		}
	}()
}

func sendMessage(m types.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		log.Println(err)
		return
	}

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		data, err := os.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))
		if err != nil {
			log.Println(err)
			return
		}

		mailTemplate := string(data)
		for key, content := range m.TemplateContent {
			mailTemplate = strings.Replace(mailTemplate, key, content, 1)
		}

		email.SetBody(mail.TextHTML, mailTemplate)
	}

	if err = email.Send(client); err != nil {
		log.Println(err)
		return
	}

	log.Println("Email sent!")
}
