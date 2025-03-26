package main

import (
	"flag"
	"log"
)

func main() {
	flag.Parse()

	ctx, err := getContext()
	if err != nil {
		log.Fatalf("%s", err)
	}

	if ctx.SendTo != "" {
		if err := sendAttachmentMailTo(ctx); err != nil {
			log.Fatalf("%v", err)
		}
	} else {
		if err := sendAttachmentMailsAll(ctx); err != nil {
			log.Fatalf("%v", err)
		}
	}

	ctx.MailSender.Close()
}
