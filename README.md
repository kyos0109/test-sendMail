# test-sendMail
test send spam mail tools

```
docker run -it --rm -v $(pwd):/app kyos0109/test-sendMail
```
# Requested file
1. edm.html (send email html content) or cli args -m FILE .
2. mail_list (MailTo List) or cli args -l FILE.
3. sender.json (Sender Config) or cli args -c FILE.

# WorkDir
```
/app
```

# Config File Example
mail_list
```
test01@gmail.com
test02@gmail.com
....
...
..
```

sender.json
```
{
  "SendersProfile": [
    {
      "SMTPHost": "127.0.0.1",
      "UserName": "test_sender",
      "Passowrd": "password",
      "MailFrom": "notify@test01.example.com",
      "MailFromName": "noreply",
      "Enable": true
    },
    {
      "SMTPHost": "192.168.0.10",
      "UserName": "test_sender",
      "Passowrd": "password",
      "MailFrom": "notify@test02.example.com",
      "MailFromName": "noreply",
      "Enable": false
    },
  ]
}
```
