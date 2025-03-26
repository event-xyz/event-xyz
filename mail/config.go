package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	gomail "gopkg.in/mail.v2"
	"gopkg.in/yaml.v2"
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
		log.Println("Warning: No unsent emails log set")
		ctx.UnsentFile = nil
	} else {
		file, err := os.OpenFile(ctx.UnsentFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("unable to open unsent log file %s: %v", ctx.UnsentFilePath, err)
		}
		ctx.UnsentFile = file
	}

	if ctx.SendFilePath != "" {
		log.Printf("Sending only mails listed in file: %s\n", ctx.SendFilePath)

		emails, err := loadEmailsFromFile(ctx.SendFilePath)
		if err != nil {
			return nil, err
		}
		ctx.Emails = emails
	}

	dialer := gomail.NewDialer(ctx.Host, ctx.PortNo, ctx.Username, ctx.Password)

	if ctx.DevMode == "true" {
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

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func getConfig() (*Context, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if err := loadEnvFile("./.env"); err != nil {
		return nil, fmt.Errorf("environment loading failed: %w", err)
	}

	configPath := os.Getenv("CONFIGPATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	sender := &Sender{
		Host:     os.Getenv("HOST"),
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
	}

	ctx := &Context{
		Config: config,
		Sender: sender,
		Emails: nil,
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
		{ctx.Username, "USERNAME"},
		{ctx.Password, "PASSWORD"},
		{ctx.Host, "HOST"},
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
	}

	if *fileFlag != "" {
		ctx.SendFilePath = *fileFlag
	}

	return nil
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
