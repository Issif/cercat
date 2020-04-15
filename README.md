# cercat

`certcat` is for **Certificate Catcher**. It's monitors issued certificates from [CertStream](https://certstream.calidog.io/) stream and send an alert to **Slack** if a domain matchs a specified **regexp**.

```bash
               websocket    +----------+   POST
CertSteam <-----------------> certcat  +-----------> Slack
                            | (regexp) |
                            +----------+
```

![screenshot](https://github.com/issif/cercat/raw/master/screenshot.png)

It's highly inspired by [CertStreamMonitor](https://github.com/AssuranceMaladieSec/CertStreamMonitor/blob/master/README.md), the first idea was to improve performances for catching with a **Golang** version.

## Configuration

Two methods are available for configuration and can be mixed :
- *config file*
- *environment variables* (they override values in *config file*)

### With config file

```bash
---
SlackWebhookURL: "" #Slack Webhook URL
SlackIconURL: "" #Slack Icon (Avatar) URL
SlackUsername: "" #Slack Username
Regexp: ".*\\.fr$" #Regexp to match. Can't be empty. It uses Golang regexp format
Workers: 20 #Number of workers for consuming feed from CertStream
DisplayErrors: false #Enable/Disable display of errors in logs
```

### With env vars

- **SLACKWEBHOOKURL**: Slack Webhook URL
- **SLACKICONURL**: Slack Icon (Avatar) URL
- **SLACKUSERNAME**: Slack Username
- **REGEXP**: Regexp to match, if empty, '.*' is used. Use Golang regexp format
- **WORKERS**: Number of workers for consuming feed from CertStream
- **DISPLAYERRORS**: Enable/Disable display of errors in logs

## Run

```
usage: cercat [<flags>]

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
  -c, --configfile=CONFIGFILE  config file
```

## Logs

```bash
2020/04/14 17:29:40 [INFO]  : A certificate for 'www.XXXX.fr' has been issued : {"domain":"www.XXXX.fr","SAN":["www.XXXX.fr"],"issuer":"Let's Encrypt","Addresses":["XX.XX.XX.183","XX.XX.XX.182"]}
2020/04/14 17:29:41 [INFO]  : A certificate for 'XXXX.fr' has been issued : {"domain":"XXXX.fr","SAN":["mail.XXXX.fr","XXXX.fr","www.XXXX.fr"],"issuer":"Let's Encrypt","Addresses":["XX.XX.XX.108"]}
```

## License

MIT

## Author

Thomas Labarussias - [@Issif](https://www.github.com/issif)
