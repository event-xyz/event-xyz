package main

import (
	"os"

	gomail "gopkg.in/mail.v2"
)

type Set[T comparable] map[T]struct{}

func (set Set[T]) Add(ele T) {
	set[ele] = struct{}{}
}

func (set Set[T]) Contains(ele T) bool {
	_, exists := set[ele]
	return exists
}

type GeneralConfig struct {
	UnsentFilePath string // file to log failed email addresses
	AttachmentPath string
	Extension      string
	MessagePath    string
	PortNo         int
}

type SenderConfig struct {
	Host        string
	Username    string
	Password    string
	MessageHead string
	MessageBody string
}

type FlagsConfig struct {
	SendWhat     string
	SendTo       string
	SendFilePath string
	DevMode      string
}

type Context struct {
	*GeneralConfig
	*SenderConfig
	*FlagsConfig

	MailSender gomail.SendCloser
	UnsentFile *os.File
	Emails     Set[string]
}

type Receiver struct {
	email string
	name  string
}
