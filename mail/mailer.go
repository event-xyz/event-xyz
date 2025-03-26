package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	gomail "gopkg.in/mail.v2"
)

func sendAttachmentMailTo(ctx *Context) error {
	var currentAttachments []string
	var currentName string

	walkerFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (ctx.Extension != "" && !strings.HasSuffix(info.Name(), ctx.Extension)) {
			return nil
		}

		email, name := extractEmailFromFileName(info.Name())
		if email == "" {
			return fmt.Errorf("could not extract email from file name: %s", info.Name())
		}

		// send emails specifically
		if email != ctx.SendWhat {
			return nil
		}

		currentAttachments = append(currentAttachments, fmt.Sprintf("%s/%s", ctx.AttachmentPath, info.Name()))
		currentName = name

		return nil
	}

	err := filepath.Walk(fmt.Sprintf("./%s", ctx.AttachmentPath), walkerFunc)
	if err != nil {
		return err
	}

	if len(currentAttachments) > 0 {
		if err := sendMail(
			ctx,
			Receiver{
				email: ctx.SendTo,
				name:  currentName,
			},
			currentAttachments,
		); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no files to send")
	}

	return nil
}

func sendAttachmentMailsAll(ctx *Context) error {
	var currentEmail string
	var currentAttachments []string
	var currentName string // could refer to name or team name whatever is in qr filename

	walkerFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (ctx.Extension != "" && !strings.HasSuffix(info.Name(), ctx.Extension)) {
			return nil
		}

		// Extract email from the file name
		email, name := extractEmailFromFileName(info.Name())
		if email == "" {
			return fmt.Errorf("could not extract email from file name: %s", info.Name())
		}

		// send emails listed in file
		if ctx.Emails != nil && !ctx.Emails.Contains(email) {
			return nil
		}

		if email != currentEmail {
			if currentEmail != "" {
				if err := sendMail(
					ctx,
					Receiver{
						email: currentEmail,
						name:  currentName,
					},
					currentAttachments,
				); err != nil {
					return err
				}

				currentAttachments = nil
			}
			currentEmail = email
			currentName = name
		}

		currentAttachments = append(currentAttachments, fmt.Sprintf("%s/%s", ctx.AttachmentPath, info.Name()))

		return nil
	}

	err := filepath.Walk(fmt.Sprintf("./%s", ctx.AttachmentPath), walkerFunc)

	if len(currentAttachments) >= 1 {
		if err := sendMail(ctx, Receiver{email: currentEmail, name: currentName}, currentAttachments); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func extractEmailFromFileName(fileName string) (string, string) {
	parts := strings.Split(fileName, "-")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return "", ""
}

func sendMail(ctx *Context, receiver Receiver, attachmentPaths []string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", ctx.Username)
	msg.SetHeader("To", receiver.email)
	msg.SetHeader("Subject", ctx.MessageHead)
	msg.SetBody("text/plain", ctx.MessageBody)

	for _, attachment := range attachmentPaths {
		msg.Attach(attachment)
	}

	if err := gomail.Send(ctx.MailSender, msg); err != nil {
		byteInfo := []byte(receiver.email + "\n")
		if _, err := ctx.UnsentFile.Write(byteInfo); err != nil {
			return fmt.Errorf("failed to log unsent email for: %s, email may have been sent. (original error): %v", receiver.email, err)
		}

		log.Printf("Failed to send email to %s. (original error): %v", receiver.email, err) // just log that we couldnt send email and move on
		return nil
	}

	log.Printf("Email sent to: %s\n", receiver.email)

	return nil
}
