package lib

import (
	"cercat/config"
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"cercat/pkg/slack"
	"context"
	"encoding/json"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
)

// the websocket stream from calidog
const certInput = "wss://certstream.calidog.io"

// CertCheckWorker parses certificates and raises alert if matches config
func CertCheckWorker(r string, homoglyph *map[string]string, msgChan chan []byte, bufferChan chan *model.Result) {
	reg, _ := regexp.Compile(r)

	for {
		msg := <-msgChan
		result, err := ParseResultCertificate(msg)
		if err != nil {
			log.Warnf("Error parsing message: %s", err)
			continue
		}
		if result == nil {
			continue
		}
		if !IsMatchingCert(homoglyph, result, reg) {
			continue
		}
		result.Addresses = fetchIPAddresses(result.Domain)
		bufferChan <- result
	}
}

// ParseResultCertificate parses certificate details
func ParseResultCertificate(msg []byte) (*model.Result, error) {
	var c model.Certificate
	var r *model.Result

	err := json.Unmarshal(msg, &c)
	if err != nil || c.MessageType == "heartbeat" {
		return nil, err
	}

	r = &model.Result{
		Domain:    c.Data.LeafCert.Subject["CN"],
		Issuer:    c.Data.LeafCert.Issuer["O"],
		SAN:       c.Data.LeafCert.AllDomains,
		Addresses: []string{},
	}

	return r, nil
}

// isIPv4Net checks if IP is IPv4
func isIPv4Net(host string) bool {
	return net.ParseIP(host) != nil
}

// fetchIPAddresses resolves domain to get IP
func fetchIPAddresses(name string) []string {
	var ipsList []string

	ips, err := net.LookupIP(name)
	if err != nil || len(ips) == 0 {
		log.Debugf("Could not fetch IP addresses of domain %s", name)
		return ipsList
	}
	for _, j := range ips {
		if isIPv4Net(j.String()) {
			ipsList = append(ipsList, j.String())
		}
	}
	return ipsList
}

// isIDN checks if domain is an IDN
func isIDN(domain string) bool {
	s := strings.Split(domain, ".")
	for _, i := range s {
		if strings.HasPrefix(i, "xn--") {
			return true
		}
	}
	return false
}

// IsMatchingCert checks if certificate matches the regexp
func IsMatchingCert(homoglyphs *map[string]string, result *model.Result, reg *regexp.Regexp) bool {
	domainList := append(result.SAN, result.Domain)
	for _, domain := range domainList {
		if isIDN(domain) {
			result.IDN, _ = idna.ToUnicode(domain)
			domain = homoglyph.ReplaceHomoglyph(result.IDN, *homoglyphs)
		}
		if reg.MatchString(domain) {
			return true
		}
	}
	return false
}

// LoopCertStream gathers messages from CertStream
func LoopCertStream(msgBuf chan []byte) {
	dial := ws.Dialer{
		ReadBufferSize:  8192,
		WriteBufferSize: 512,
		Timeout:         1 * time.Second,
	}
	for {
		// conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), certInput)
		conn, _, _, err := dial.Dial(context.Background(), certInput)
		if err != nil {
			log.Warn("Error connecting to CertStream! Sleeping a few seconds and reconnecting...")
			time.Sleep(1 * time.Second)
			// conn.Close()
			continue
		}
		for {
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				log.Warn("Error reading message from CertStream")
				break
			}
			msgBuf <- msg
		}
		conn.Close()
	}
}

// Notifier is a worker that receives cert, depduplicates and sends to Slack the event
func Notifier(cfg *config.Configuration) {
	for {
		result := <-cfg.Buffer
		duplicate := false
		cfg.PreviousCerts.Do(func(d interface{}) {
			if result.Domain == d {
				duplicate = true
			}
		})
		if !duplicate {
			cfg.PreviousCerts = cfg.PreviousCerts.Prev()
			cfg.PreviousCerts.Value = result.Domain
			j, _ := json.Marshal(result)
			log.Infof("A certificate for '%v' has been issued : %v\n", result.Domain, string(j))
			if cfg.SlackWebHookURL != "" {
				go func(c *config.Configuration, r *model.Result) {
					slack.NewPayload(c, result).Post(c)
				}(cfg, result)
			}
		}
	}
}
