# Mail

Tool to mass send emails with attachments

```
go run .
```

`.env` example:
```
USERNAME=<email>
PASSWORD=<password>
HOST=smtp.mail.com
```

Create your app password here: https://myaccount.google.com/u/3/apppasswords

### Flags
Mails meant for one email can be sent to another email using `-send` flag:
```
./mail -send '<original-email> <receiver-email>'
```

The tool takes files with this name format:
`<email>-<name>-[other-fields].ext`

You can also send messages through cli flag:
`./mail -msg '<message-head> <message-body>'`

### Notes
1. All paths to be specified are relative
2. Tool assumes that all emails in files names are correct
