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

type Config struct {
	UnsentFilePath string `yaml:"unsent_log"`
	AttachmentPath string `yaml:"attachment_path"`
	Extension      string `yaml:"extension"`
	PortNo         int    `yaml:"port_no"`
	SendFilePath   string `yaml:"sendfile_path"`
	MessageHead    string `yaml:"message_head"`
	MessageBody    string `yaml:"message_body"`
	DevMode        string `yaml:"devmode"`
	SendWhat       string `yaml:"-"`
	SendTo         string `yaml:"-"`
}

type Sender struct {
	Host     string
	Username string
	Password string
}

type Context struct {
	*Config
	*Sender

	MailSender gomail.SendCloser
	UnsentFile *os.File
	Emails     Set[string]
}

type Receiver struct {
	email string
	name  string
}
