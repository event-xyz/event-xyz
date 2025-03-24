package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gomail "gopkg.in/mail.v2"
)

var (
	sendFlag = flag.String("send", "", "Send specific files (format: 'search-email recipient-email')")
	extFlag  = flag.String("ext", "", "What extension files to send")
	msgFlag  = flag.String("msg", "", "What message you would like to send (format: 'heading body')")
)

func main() {
	flag.Parse()

	config, err := getConfig()
	if err != nil {
		log.Fatalf("%s", err)
	}

	dialer := gomail.NewDialer(config.host, config.portno, config.username, config.password)
	// For production, remove InsecureSkipVerify or set it to false.
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	sender, err := dialer.Dial()
	if err != nil {
		log.Fatalf("Could not connect to SMTP server: %v", err)
	}
	defer sender.Close()

	if config.sendSpecific {
		if err := sendMailTo(sender, &config); err != nil {
			log.Fatalf("%v", err)
		}
	} else {
		if err := sendMailsAll(sender, &config); err != nil {
			log.Fatalf("%v", err)
		}
	}

	fmt.Println("Operation Complete!")
}

func getConfig() (Config, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	envFile, err := os.Open("./.env")
	if err != nil {
		return Config{}, fmt.Errorf("could not read .env file: %v", err)
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		envVar := strings.Split(scanner.Text(), "=")
		os.Setenv(envVar[0], envVar[1])
	}

	portnoString := os.Getenv("PORTNO")
	portno, err := strconv.Atoi(portnoString)
	if err != nil {
		return Config{}, fmt.Errorf("enter valid number for PORTNO")
	}

	config := Config{
		username:       os.Getenv("USERNAME"),
		password:       os.Getenv("PASSWORD"),
		host:           os.Getenv("HOST"),
		unsentFile:     os.Getenv("UNSENTLOG"),
		attachmentPath: os.Getenv("ATTACHMENTPATH"),
		extension:      os.Getenv("EXTENSION"),
		messagePath:    os.Getenv("MESSAGEPATH"),
		portno:         portno,
		sendSpecific:   false,
		checkExtension: false,
	}

	if *sendFlag != "" {
		sendDetails := strings.Split(*sendFlag, " ")
		if len(sendDetails) != 2 {
			return Config{}, fmt.Errorf("use format: -send 'search_term recipient_email'")
		}
		config.sendWhat = sendDetails[0]
		config.sendTo = sendDetails[1]
		config.sendSpecific = true
	}

	if *extFlag != "" {
		config.checkExtension = true
		config.extension = *extFlag
	}

	if *msgFlag != "" {
		messageDetails := strings.Split(*msgFlag, "")
		if len(messageDetails) != 2 {
			return Config{}, fmt.Errorf("use format: -msg 'message-head message-body'")
		}

		config.messageHead = messageDetails[0]
		config.messageBody = messageDetails[1]
	} else {
		if config.messagePath == "" {
			return Config{}, fmt.Errorf("no message or message file path provided")
		}
		messageHead, messageBody, err := parseMessage(config.messagePath)
		if err != nil {
			return Config{}, err
		}

		config.messageHead = *messageHead
		config.messageBody = *messageBody
	}

	if config.username == "" || config.password == "" || config.unsentFile == "" ||
		config.attachmentPath == "" || portnoString == "" {
		return Config{}, fmt.Errorf("not all env variables are filled")
	}

	return config, nil
}

func parseMessage(messagePath string) (*string, *string, error) {
	splitMessageDetails := func(messageDetails []string) (*string, *string) {
		messageHead := messageDetails[0]
		messageBody := strings.Join(messageDetails, "\n")

		return &messageHead, &messageBody
	}

	if *msgFlag != "" {
		messageDetails := strings.Split(*msgFlag, "")
		if len(messageDetails) < 2 {
			return nil, nil, fmt.Errorf("use format: -msg 'message-head message-body'")
		}

		messageHead, messageBody := splitMessageDetails(messageDetails)
		return messageHead, messageBody, nil
	}

	bytes, err := os.ReadFile(messagePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open message file: %v", err)
	}

	message := string(bytes[:])
	messageDetails := strings.Split(message, "\n")

	if len(messageDetails) < 2 {
		return nil, nil, fmt.Errorf("wrong format for message file, check README")
	}

	messageHead, messageBody := splitMessageDetails(messageDetails)
	return messageHead, messageBody, nil
}

func extractEmailFromFileName(fileName string) (string, string) {
	parts := strings.Split(fileName, "-")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return "", ""
}

func sendMailTo(sender gomail.SendCloser, config *Config) error {
	var currentAttachments []string
	var currentName string

	walkerFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (config.checkExtension && !strings.HasSuffix(info.Name(), config.extension)) {
			return nil
		}

		email, name := extractEmailFromFileName(info.Name())
		if email == "" {
			return fmt.Errorf("could not extract email from file name: %s", info.Name())
		}

		if email != config.sendWhat {
			return nil
		}

		currentAttachments = append(currentAttachments, fmt.Sprintf("%s/%s", config.attachmentPath, info.Name()))
		currentName = name

		return nil
	}

	err := filepath.Walk(fmt.Sprintf("./%s", config.attachmentPath), walkerFunc)
	if err != nil {
		return err
	}

	if len(currentAttachments) > 0 {
		if err := sendMail(
			sender,
			Receiver{
				email: config.sendTo,
				name:  currentName,
			},
			config,
			currentAttachments,
		); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no files to send")
	}

	return nil
}

// TODO: Change name to team name once changes pushed on eventloop
func sendMailsAll(sender gomail.SendCloser, config *Config) error {
	var currentEmail string
	var currentAttachments []string
	var currentName string

	walkerFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (config.checkExtension && !strings.HasSuffix(info.Name(), config.extension)) {
			return nil
		}

		// Extract email from the file name
		email, name := extractEmailFromFileName(info.Name())
		if email == "" {
			return fmt.Errorf("could not extract email from file name: %s", info.Name())
		}

		if email != currentEmail {
			if currentEmail != "" {
				if err := sendMail(
					sender,
					Receiver{
						email: currentEmail,
						name:  currentName,
					},
					config,
					currentAttachments,
				); err != nil {
					return err
				}

				currentAttachments = nil
			}
			currentEmail = email
			currentName = name
		}

		currentAttachments = append(currentAttachments, fmt.Sprintf("%s/%s", config.attachmentPath, info.Name()))

		return nil
	}

	err := filepath.Walk(fmt.Sprintf("./%s", config.attachmentPath), walkerFunc)

	if len(currentAttachments) >= 1 {
		if err := sendMail(sender, Receiver{email: currentEmail, name: currentName}, config, currentAttachments); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func sendMail(sender gomail.SendCloser, receiver Receiver, config *Config, attachmentPaths []string) error {
	file, err := os.OpenFile(config.unsentFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("unable to open unsent file: %s: %v", config.unsentFile, err)
	}
	defer file.Close()

	msg := gomail.NewMessage()

	// Set email headers
	msg.SetHeader("From", config.username)
	msg.SetHeader("To", receiver.email)
	msg.SetHeader("Subject", config.messageHead)

	msg.SetBody("text/plain", config.messageBody)

	for _, attachment := range attachmentPaths {
		msg.Attach(attachment)
	}

	if err := gomail.Send(sender, msg); err != nil {
		byteInfo := []byte(receiver.email + "\n")
		if _, err := file.Write(byteInfo); err != nil {
			return fmt.Errorf("could not write to unsent file, email may have been sent to %s: %v", receiver.email, err)
		}

		return fmt.Errorf("could not send email to %s: %v", receiver.email, err)
	}

	log.Printf("Email sent to: %s\n", receiver.email)

	return nil
}
