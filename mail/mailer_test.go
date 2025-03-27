// TODO: Write better tests, but for now fine ig

package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	gomail "gopkg.in/mail.v2"
)

var (
	// None of these mails exist but we can use them for testing purposes
	// the app will successfully send out all mails
	// you should see them bounce back to your inbox

	// add your own email to the list if you want to test
	testEmails = []string{
		"foo@bar.com",
		"test@test.com",
		"sudhirrsbrigtestmind@gmail.com",
		"sprite@sprite.com",
		"pepsi@pepsi.com",
		"dictatoresh@gmail.com",
	}

	attachmentsPath = "test_attachments"
)

// generate empty files to send
func setupAttachments() error {
	if err := os.MkdirAll(attachmentsPath, 0755); err != nil {
		return fmt.Errorf("failed to create attachments directory: %v", err)
	}

	// Generate between 1 to 5 attachments per email.
	for _, email := range testEmails {
		nFiles := rand.Intn(5) + 1
		for i := 0; i < nFiles; i++ {
			randomNumber := rand.Intn(1000000) // Random number for filename uniqueness
			fileName := fmt.Sprintf("%s-test-%d.txt", email, randomNumber)
			filePath := filepath.Join(attachmentsPath, fileName)

			f, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create attachment file %s: %v", filePath, err)
			}
			f.Close()
		}
	}

	return nil
}

func removeAttachments() error {
	if attachmentsPath != "" {
		return os.RemoveAll(attachmentsPath)
	}
	return nil
}

func TestMain(m *testing.M) {
	rand.New(rand.NewSource(1))

	if err := setupAttachments(); err != nil {
		log.Fatalf("Setup failed: %v", err)
	}

	code := m.Run()

	if err := removeAttachments(); err != nil {
	 	log.Printf("Removal failed: %v", err)
	}

	os.Exit(code)
}

func getTestContext(config *Config) (*Context, error) {
	if err := loadEnvFile("./.env"); err != nil {
		return nil, fmt.Errorf("environment loading failed: %w", err)
	}

	senderConfig := &Sender{
		Host:     os.Getenv("HOST"),
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
	}

	ctx := &Context{
		Config: config,
		Sender: senderConfig,
	}

	if ctx.Sender.Host == "" || ctx.Sender.Username == "" || ctx.Sender.Password == "" {
		return nil, fmt.Errorf("env variables not set")
	}

	dialer := gomail.NewDialer(ctx.Sender.Host, ctx.Config.PortNo, ctx.Sender.Username, ctx.Sender.Password)

	if ctx.Config.DevMode == "1" {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	sender, err := dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("could not connect to SMTP server: %v", err)
	}

	ctx.MailSender = sender

	return ctx, nil
}

func TestSendMailAll(t *testing.T) {
	config := &Config{
		AttachmentPath: attachmentsPath,
		PortNo:         587,
		MessageHead:    "Test",
		MessageBody:    "Hello,\nThis is a test email",
		DevMode:        "true",
	}

	ctx, err := getTestContext(config)
	if err != nil {
		t.Fatalf("Failed to get test context: %v", err)
	}
	defer ctx.MailSender.Close()

	if err := sendAttachmentMailsAll(ctx); err != nil {
		t.Errorf("sendAttachmentMailsAll() error: %v", err)
	}
}

func TestSendOtherMailToSpecificRecipient(t *testing.T) {
	config := &Config{
		AttachmentPath: attachmentsPath,
		PortNo:         587,
		MessageHead:    "Test",
		MessageBody:    "Hello,\nThis is a test email.",
		DevMode:        "true",
	}

	sendTo := testEmails[rand.Intn(len(testEmails))]
	sendWhat := testEmails[rand.Intn(len(testEmails))]

	config.SendTo = sendTo
	config.SendWhat = sendWhat

	ctx, err := getTestContext(config)
	if err != nil {
		t.Fatalf("Failed to get test context: %v", err)
	}
	defer ctx.MailSender.Close()

	if err := sendAttachmentMailTo(ctx); err != nil {
		t.Errorf("sendAttachmentMailTo() error: %v", err)
	}
}

func TestSendMailToSpecificRecipient(t *testing.T) {
	config := &Config{
		AttachmentPath: attachmentsPath,
		PortNo:         587,
		MessageHead:    "Test",
		MessageBody:    "Hello,\nThis is a test email.",
		DevMode:        "true",
	}

	sendTo := testEmails[rand.Intn(len(testEmails))]
	config.SendTo = sendTo
	config.SendWhat = sendTo

	ctx, err := getTestContext(config)
	if err != nil {
		t.Fatalf("Failed to get test context: %v", err)
	}
	defer ctx.MailSender.Close()

	if err := sendAttachmentMailTo(ctx); err != nil {
		t.Errorf("sendAttachmentMailTo() error: %v", err)
	}
}
