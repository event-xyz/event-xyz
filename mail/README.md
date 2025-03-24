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
Create your app password here: https://myaccount.google.com/u/3/apppasswords

Mails meant for one email can be sent to another email using `-send` flag:
```
./mail -send '<original-email> <receiver-email>'
```

The tool takes files with this name format:
`<email>-<name>-[other-fields].ext`

Format of message file:
```
<Message-Head>
<Message Body>
```
You can also send messages through cli flag:
`./mail -msg '<message-head> <message-body>'`

PSA: All paths to be specified are relative