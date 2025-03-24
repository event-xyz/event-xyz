package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gomail "gopkg.in/mail.v2"
)

type Config struct {
	username       string
	password       string
	host           string
	logfile        string
	attachmentPath string
	extension      string
	portno         int
}

type Receiver struct {
	email string
	name  string
}

func extractEmailFromFileName(fileName string) (string, string) {
	parts := strings.Split(fileName, "-")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return "", ""
}

func sendMailsAll(sender gomail.SendCloser, config Config) error {
	var currentEmail string
	var currentAttachments []string
	var currentName string

	// Walk through the directory and process each PNG file
	err := filepath.Walk(fmt.Sprintf("./%s", config.attachmentPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), config.extension) {
			// Extract email from the file name
			email, name := extractEmailFromFileName(info.Name())
			if email == "" {
				return fmt.Errorf("could not extract email from file name: %s", info.Name())
			}

			if email != currentEmail {
				if currentEmail != "" {
					err := sendMail(
						sender,
						Receiver{
							email: currentEmail,
							name:  currentName,
						},
						config,
						currentAttachments,
					)

					if err != nil {
						return err
					}

					currentAttachments = nil
				}
				currentEmail = email
				currentName = name
			}

			currentAttachments = append(currentAttachments, fmt.Sprintf("%s/%s", config.attachmentPath, info.Name()))
		}

		return nil
	})

	if len(currentAttachments) >= 1 {
		if err := sendMail(sender, Receiver{email: currentEmail, name: currentName}, config, currentAttachments); err != nil {
			return err
		}
	}

	if err != nil {
		return fmt.Errorf("error walking the directory: %v", err)
	}

	return nil
}

func sendMail(sender gomail.SendCloser, receiver Receiver, config Config, attachmentPaths []string) error {
	file, err := os.OpenFile(config.logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("unable to open logfile: %s: %v", config.logfile, err)
	}
	defer file.Close()

	msg := gomail.NewMessage()

	// Set email headers
	msg.SetHeader("From", config.username)
	msg.SetHeader("To", receiver.email)
	msg.SetHeader("Subject", "Important Email with Attachments")

	msg.SetBody("text/plain", fmt.Sprintf("Hello %s,\n\nPlease find your attachments.\n\nBest regards", receiver.name))

	for _, attachment := range attachmentPaths {
		msg.Attach(attachment)
	}

	if err := gomail.Send(sender, msg); err != nil {
		return fmt.Errorf("could not send email to %s: %v", receiver.email, err)
	}

	byteInfo := []byte(receiver.email + "\n")
	if _, err := file.Write(byteInfo); err != nil {
		return fmt.Errorf("could not write to logfile, email may have been sent to %s: %v", receiver.email, err)
	}

	return nil
}

func getConfig() (Config, error) {
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
		logfile:        os.Getenv("LOGFILE"),
		attachmentPath: os.Getenv("ATTACHMENTPATH"),
		extension:      os.Getenv("EXTENSION"),
		portno:         portno,
	}

	if config.username == "" || config.password == "" || config.logfile == "" ||
		config.attachmentPath == "" || config.extension == "" || portnoString == "" {
		return Config{}, fmt.Errorf("not all env variables are filled")
	}

	fmt.Printf("%v", config)

	return config, nil
}

func main() {
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

	if err := sendMailsAll(sender, config); err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println("Sent mail!")

}
