package main

type Config struct {
	// general config
	unsentFile     string // file to log all emails we failed to send emails to
	attachmentPath string
	extension      string
	messagePath    string
	portno         int

	// sender info
	host        string
	username    string
	password    string
	messageHead string
	messageBody string

	// flags
	checkExtension bool
	// parallel       bool

	sendSpecific bool
	sendWhat     string
	sendTo       string
}

type Receiver struct {
	email string
	name  string
}
