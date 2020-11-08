package lib

import (
	"cercat/config"
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"cercat/pkg/slack"
	"context"
	"encoding/json"
	"net"
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
func CertCheckWorker(config *config.Configuration) {
	for {
		msg := <-config.Messages
		result, err := ParseResultCertificate(msg, &config.HomoglyphPatterns)
		if err != nil {
			log.Warnf("Error parsing message: %s", err)
			continue
		}
		if result == nil {
			continue
		}
		if match, attack := IsMatchingCert(result, config); match {
			result.Attack = attack
			config.Buffer <- result
		}
	}
}

// ParseResultCertificate parses certificate details
func ParseResultCertificate(msg []byte, homoglyphs *map[string]string) (*model.Result, error) {
	var c model.Certificate
	var r *model.Result

	err := json.Unmarshal(msg, &c)
	if err != nil || c.MessageType == "heartbeat" {
		return nil, err
	}
	r = &model.Result{
		Domain:    c.Data.LeafCert.Subject["CN"],
		Issuer:    c.Data.Chain[0].Subject["O"],
		SAN:       c.Data.LeafCert.AllDomains,
		Addresses: []string{},
	}
	r.Addresses = fetchIPAddresses(r.Domain)
	if isIDN(r.Domain) {
		r.IDN, _ = idna.ToUnicode(r.Domain)
		r.UnicodeIDN = homoglyph.ReplaceHomoglyph(r.IDN, *homoglyphs)
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
func IsMatchingCert(result *model.Result, config *config.Configuration) (bool, string) {
	domainsToCheck := []string{}
	var attack string
	if len(result.UnicodeIDN) != 0 {
		domainsToCheck = append(domainsToCheck, result.UnicodeIDN)
		attack = "homoglyph + "
	} else {
		domainsToCheck = append([]string{result.Domain}, result.SAN...)
	}
	for _, i := range config.Domains {
		for _, domain := range domainsToCheck {
			cleanedDomain := strings.ReplaceAll(strings.ReplaceAll(domain, "-", ""), ".", "")
			if isContained(cleanedDomain, config.InclusionPatterns[i]) {
				return true, attack + "inclusion"
			}
			if isContained(cleanedDomain, config.VowelSwapPatterns[i]) {
				return true, attack + "vowelswap"
			}
			if isContained(cleanedDomain, config.RepetitionPatterns[i]) {
				return true, attack + "repetition"
			}
			if isContained(cleanedDomain, config.TranspositionPatterns[i]) {
				return true, attack + "transposition"
			}
			if isContained(cleanedDomain, config.OmissionPatterns[i]) {
				return true, attack + "omission"
			}
			if isContained(cleanedDomain, config.BitsquattingPatterns[i]) {
				return true, attack + "bitsquatting"
			}
		}
	}
	return false, ""
}

func isContained(domain string, patterns []string) bool {
	for _, i := range patterns {
		if strings.Contains(domain, i) {
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
			conn.Close()
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

// Notifier is a worker that receives cert, deduplicates and sends to Slack the event
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
