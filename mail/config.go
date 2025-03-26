package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

var (
	sendFlag = flag.String("send", "", "Send specific files (format: 'search-email recipient-email')")
	extFlag  = flag.String("ext", "", "What extension files to send")
	msgFlag  = flag.String("msg", "", "What message you would like to send (format: 'heading body')")
	fileFlag = flag.String("file", "", "Send emails only to those listed in a file (format: 'filepath')")
)

func getContext() (*Context, error) {
	ctx, err := getConfig()
	if err != nil {
		return nil, err
	}

	if ctx.UnsentFilePath == "" {
		return nil, fmt.Errorf("no unsent log file set")
	}
	file, err := os.OpenFile(ctx.UnsentFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("unable to open unsent log file %s: %v", ctx.UnsentFilePath, err)
	}
	ctx.UnsentFile = file

	if ctx.SendFilePath != "" {
		log.Printf("Sending only mails listed in file: %s\n", ctx.SendFilePath)

		emails, err := loadEmailsFromFile(ctx.SendFilePath)
		if err != nil {
			return nil, err
		}
		ctx.Emails = emails
	}

	dialer := gomail.NewDialer(ctx.Host, ctx.PortNo, ctx.Username, ctx.Password)

	if ctx.DevMode == "1" {
		// For production, remove InsecureSkipVerify or set it to false.
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	sender, err := dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("could not connect to SMTP server: %v", err)
	}

	ctx.MailSender = sender

	return ctx, nil
}

func getConfig() (*Context, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if err := loadEnvFile("./.env"); err != nil {
		return nil, fmt.Errorf("environment loading failed: %w", err)
	}

	portNo, err := strconv.Atoi(os.Getenv("PORTNO"))
	if err != nil {
		return nil, fmt.Errorf("enter valid number for PORTNO")
	}

	general := &GeneralConfig{
		UnsentFilePath: os.Getenv("UNSENTLOG"),
		AttachmentPath: os.Getenv("ATTACHMENTPATH"),
		Extension:      os.Getenv("EXTENSION"),
		MessagePath:    os.Getenv("MESSAGEPATH"),
		PortNo:         portNo,
	}
	sender := &SenderConfig{
		Host:     os.Getenv("HOST"),
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
	}
	flags := &FlagsConfig{
		SendFilePath: os.Getenv("SENDFILEPATH"),
		DevMode:      os.Getenv("DEVMODE"),
	}

	ctx := &Context{
		GeneralConfig: general,
		SenderConfig:  sender,
		FlagsConfig:   flags,
		Emails:        nil,
	}

	// Update static config with flag values.
	if err := parseFlags(ctx); err != nil {
		return nil, err
	}

	if err := validateConfig(ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}

func validateConfig(ctx *Context) error {
	required := []struct {
		value string
		name  string
	}{
		// portno, message, sendWhat, sendTo, unsent log file are checked elsewhere
		// extension may be empty
		{ctx.Username, "USERNAME"},
		{ctx.Password, "PASSWORD"},
		{ctx.Host, "HOST"},
		{ctx.AttachmentPath, "ATTACHMENTPATH"},
	}

	for _, field := range required {
		if field.value == "" {
			return fmt.Errorf("not all required env variables are filled")
		}
	}

	return nil
}

func loadEnvFile(envPath string) error {
	envFile, err := os.Open(envPath)
	if err != nil {
		return fmt.Errorf("could not read .env file: %v", err)
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		envVar := strings.SplitN(scanner.Text(), "=", 2)
		os.Setenv(envVar[0], envVar[1])
	}

	return nil
}

func parseFlags(ctx *Context) error {
	if *sendFlag != "" {
		sendDetails := strings.SplitN(*sendFlag, " ", 2)

		switch len(sendDetails) {
		case 1:
			ctx.SendWhat = sendDetails[0]
			ctx.SendTo = sendDetails[0]
		case 2:
			ctx.SendWhat = sendDetails[0]
			ctx.SendTo = sendDetails[1]
		default:
			return fmt.Errorf("use format: -send 'search_term recipient_email'")
		}
	}

	if *extFlag != "" {
		ctx.Extension = *extFlag
	}

	if *msgFlag != "" {
		messageDetails := strings.SplitN(*msgFlag, "", 2)
		if len(messageDetails) != 2 {
			return fmt.Errorf("use format: -msg 'message-head message-body'")
		}

		ctx.MessageHead = messageDetails[0]
		ctx.MessageBody = messageDetails[1]
	} else {
		if ctx.MessagePath == "" {
			return fmt.Errorf("no message or message file path provided")
		}
		head, body, err := parseMessageFile(ctx.MessagePath)
		if err != nil {
			return err
		}

		ctx.MessageHead = *head
		ctx.MessageBody = *body
	}

	if *fileFlag != "" {
		ctx.SendFilePath = *fileFlag
	}

	return nil
}

func parseMessageFile(path string) (*string, *string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read message file: %v", err)
	}

	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) < 2 {
		return nil, nil, fmt.Errorf("invalid message format, heading and body should be in seperate lines")
	}

	return &lines[0], &lines[1], nil
}

func loadEmailsFromFile(path string) (Set[string], error) {
	emails := make(Set[string])

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if !emails.Contains(scanner.Text()) {
			emails.Add(scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}
