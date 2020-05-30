package lib

import (
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

// Result represents a catched certificate
type Result struct {
	Domain    string   `json:"domain"`
	IDN       string   `json:"IDN,omitempty"`
	SAN       []string `json:"SAN"`
	Issuer    string   `json:"issuer"`
	Addresses []string `json:"Addresses"`
}

// certificate represents a certificate from CertStream
type certificate struct {
	MessageType string `json:"message_type"`
	Data        data   `json:"data"`
}

// data represents data field for a certificate from CertStream
type data struct {
	UpdateType string            `json:"update_type"`
	LeafCert   leafCert          `json:"leaf_cert"`
	Chain      []leafCert        `json:"chain"`
	CertIndex  float32           `json:"cert_index"`
	Seen       float32           `json:"seen"`
	Source     map[string]string `json:"source"`
}

// leafCert represents leaf_cert field from CertStream
type leafCert struct {
	Subject      map[string]string      `json:"subject"`
	Extensions   map[string]interface{} `json:"extensions"`
	NotBefore    float32                `json:"not_before"`
	NotAfter     float32                `json:"not_after"`
	SerialNumber string                 `json:"serial_number"`
	FingerPrint  string                 `json:"fingerprint"`
	AsDer        string                 `json:"as_der"`
	AllDomains   []string               `json:"all_domains"`
}

// MsgChan is the communication channel between certCheckWorkers and LoopCertStream
var MsgChan chan []byte

// the websocket stream from calidog
const certInput = "wss://certstream.calidog.io"

// CertCheckWorker parses certificates and raises alert if matches config
func CertCheckWorker(config *Configuration) {
	reg, _ := regexp.Compile(config.Regexp)

	for {
		msg := <-MsgChan

		detailedCert, err := ParseResultCertificate(msg)
		if err != nil {
			log.Warnf("Error parsing message: %s", err)
			continue
		}
		if detailedCert == nil {
			continue
		}
		if !IsMatchingCert(config, detailedCert, reg) {
			continue
		}

		j, _ := json.Marshal(detailedCert)
		log.Infof("A certificate for '%v' has been issued : %v\n", detailedCert.Domain, string(j))
		if config.SlackWebHookURL != "" {
			go func(c *Configuration, r *Result) {
				newSlackPayload(c, detailedCert).post(c)
			}(config, detailedCert)
		}
	}
}

// ParseResultCertificate parses certificate details
func ParseResultCertificate(msg []byte) (*Result, error) {
	var c certificate
	var r *Result

	err := json.Unmarshal(msg, &c)
	if err != nil || c.MessageType == "heartbeat" {
		return nil, err
	}

	r = &Result{
		Domain:    c.Data.LeafCert.Subject["CN"],
		Issuer:    c.Data.Chain[0].Subject["O"],
		SAN:       c.Data.LeafCert.AllDomains,
		Addresses: []string{"N/A"},
	}
	r.Addresses = fetchIPAddresses(r.Domain)
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
	return strings.HasPrefix(domain, "xn--")
}

// IsMatchingCert checks if certificate matches the regexp
func IsMatchingCert(config *Configuration, cert *Result, reg *regexp.Regexp) bool {
	domainList := append(cert.SAN, cert.Domain)
	for _, domain := range domainList {
		if isIDN(domain) {
			cert.IDN, _ = idna.ToUnicode(domain)
			domain = replaceHomoglyph(cert.IDN, config.Homoglyph)
		}
		if reg.MatchString(domain) {
			return true
		}
	}
	return false
}

// LoopCertStream gathers messages from CertStream
func LoopCertStream(config *Configuration) {
	for {
		conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), certInput)
		defer conn.Close()
		if err != nil {
			log.Warn("Error connecting to CertStream! Sleeping a few seconds and reconnecting...")
			time.Sleep(1 * time.Second)
			continue
		}
		for {
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				log.Warn("Error reading message from CertStream")
				break
			}
			MsgChan <- msg
		}
	}
}
