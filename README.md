# cercat

`certcat` is for **Certificate Catcher**. It monitors issued certificates from [CertStream](https://certstream.calidog.io/) stream and sends an alert to **Slack** if a domain matches a specified **regexp**.

```bash
               websocket    +----------+   POST
CertSteam <-----------------> cercat   +-----------> Slack
                            | (regexp) |
                            +----------+
```

![screenshot](https://github.com/issif/cercat/raw/master/screenshot.png)

It's highly inspired by [CertStreamMonitor](https://github.com/AssuranceMaladieSec/CertStreamMonitor/blob/master/README.md), the first idea was to improve performances for catching with a **Golang** version.

The regexp is applied on principal and SAN domains. If one of these domains is an [IDN](https://en.wikipedia.org/wiki/Internationalized_domain_name), it's converted in an equivalent in ASCII before applying the regexp.

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
```

### With env vars

- **SLACKWEBHOOKURL**: Slack Webhook URL
- **SLACKICONURL**: Slack Icon (Avatar) URL
- **SLACKUSERNAME**: Slack Username
- **REGEXP**: Regexp to match. Can't be empty. It uses Golang regexp format

## Run

```
usage: cercat [<flags>]

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
  -c, --configfile=CONFIGFILE  config file
```

## Docker

You can run with Docker :

```
docker run -d -e SLACKWEBHOOKURL=https://hooks.slack.com/services/XXXXX -e REGEXP=".*\\.fr$" issif/cercat:latest 
```

## Logs

```bash
INFO[0005] A certificate for 'xxxx.fr' has been issued : {"domain":"xxxx.fr","SAN":["xxxx.fr","www.xxxx.fr"],"issuer":"Let's Encrypt","Addresses":["X.X.X.129"]} 
INFO[0008] A certificate for 'xxxx.fr' has been issued : {"domain":"xxxx.fr","SAN":["xxxx.fr","www.xxxx.fr"],"issuer":"Let's Encrypt","Addresses":["X.X.X.116"]} 
```

## Profiles, Traces and Metrics

The service opens port `6060` for `profiles`, `traces` and `expvar`. Go to [http://localhost:6060/debug/pprof](http://localhost:6060/debug/pprof) and [http://localhost:6060/debug/vars](http://localhost:6060/debug/vars).

## License

MIT

## Authors

Thomas Labarussias - [@Issif](https://www.github.com/issif)
Ayoul Elaassal - [@Ayoul3](https://github.com/ayoul3)

