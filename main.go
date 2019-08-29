package main

import (
	"context"
	"fmt"
	"os"

	"github.com/antihax/optional"
	mg "github.com/mailgun/mailgun-go/v3"
	mailslurp "github.com/mailslurp/mailslurp-client-go"
)

func main() {
	fmt.Println("creating clients")
	cfg := mailslurp.NewConfiguration()
	mailslurpClient := mailslurp.NewAPIClient(cfg)
	auth := context.WithValue(context.Background(), mailslurp.ContextAPIKey, mailslurp.APIKey{
		Key: getEnvOrPanic("MAILSLURP_API_KEY"),
	})

	mailgunClient := mg.NewMailgun(getEnvOrPanic("MAILGUN_DOMAIN"), getEnvOrPanic("MAILGUN_API_KEY"))

	fmt.Println("creating inbox")
	inbox, _, err := mailslurpClient.ExtraOperationsApi.CreateInbox(auth)
	mustNot(err)

	// Send email via Mailgun
	sender := fmt.Sprintf("Mailslurp Test <%s>", getEnvOrPanic("SENDER_EMAIL"))
	subject := "Mailslurp Test Email"
	body := "Test Body"

	message := mailgunClient.NewMessage(sender, subject, body, inbox.EmailAddress)
	fmt.Println("sending email")
	_, _, err = mailgunClient.Send(context.Background(), message)
	mustNot(err)

	fmt.Println("waiting for email")
	email, _, err := mailslurpClient.CommonOperationsApi.WaitForLatestEmail(auth, &mailslurp.WaitForLatestEmailOpts{
		InboxEmailAddress: optional.NewString(inbox.EmailAddress),
	})
	mustNot(err)

	fmt.Println("received email")
	fmt.Println("From: " + email.From)
	fmt.Println("Subject: " + email.Subject)
	fmt.Println("Body: " + email.Body)

	fmt.Println("deleting inbox")
	_, err = mailslurpClient.ExtraOperationsApi.DeleteInbox(auth, inbox.Id)
	mustNot(err)
}

func mustNot(err error) {
	if err != nil {
		panic(err)
	}
}

func getEnvOrPanic(env string) string {
	value := os.Getenv(env)
	if value == "" {
		panic(fmt.Sprintf("Environment variable '%s' must be set", env))
	}

	return value
}
