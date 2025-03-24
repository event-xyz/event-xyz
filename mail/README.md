# Mail

Tool to mass send emails with attachments

```
go build 
```

`.env` example:
```
USERNAME=<email>
PASSWORD=<password>
HOST=smtp.mail.com
LOGFILE=unsent.log
EXTENSION=<.ext>
ATTACHMENTPATH=<path>
PORTNO=587
```

Mails meant for one email can be sent to another email using `-send` flag:
```
./mail -send '<original-email> <receiver-email>'
```

The tool takes files with this name format:
`<email>-<name>-[other-fields].ext`