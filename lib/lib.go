package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	_ "net/http/pprof"

	_ "expvar"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Result struct {
	Domain    string   `json:"domain"`
	SAN       []string `json:"SAN"`
	Issuer    string   `json:"issuer"`
	Addresses []string `json:"Addresses"`
}

type Certificate struct {
	MessageType string `json:"message_type"`
	Data        Data   `json:"data"`
}

type Data struct {
	UpdateType string            `json:"update_type"`
	LeafCert   LeafCert          `json:"leaf_cert"`
	Chain      []LeafCert        `json:"chain"`
	CertIndex  float32           `json:"cert_index"`
	Seen       float32           `json:"seen"`
	Source     map[string]string `json:"source"`
}

type LeafCert struct {
	Subject      map[string]string      `json:"subject"`
	Extensions   map[string]interface{} `json:"extensions"`
	NotBefore    float32                `json:"not_before"`
	NotAfter     float32                `json:"not_after"`
	SerialNumber string                 `json:"serial_number"`
	FingerPrint  string                 `json:"fingerprint"`
	AsDer        string                 `json:"as_der"`
	AllDomains   []string               `json:"all_domains"`
}

// MsgChan is the communication channel between certCheckWorkers and LoopCheckCerts
var MsgChan chan []byte

const certInput = "wss://certstream.calidog.io"

// CertCheckWorker parses certificates and raises alert if matches config
func CertCheckWorker(config *Configuration) {
	reg, _ := regexp.Compile(config.Regexp)
	regIP, _ := regexp.Compile(config.RegIP)
	regIDN, _ := regexp.Compile(config.RegIDN)

	for {
		msg := <-MsgChan

		detailedCert, err := ParseResultCertificate(msg, regIP)
		if err != nil {
			log.Warnf("Error parsing message: %s", err)
			continue
		}
		if detailedCert == nil {
			continue
		}
		if !IsMatchingCert(detailedCert, reg, regIDN) {
			continue
		}
		notify(config, *detailedCert)
	}
}

func ParseResultCertificate(msg []byte, regIP *regexp.Regexp) (*Result, error) {
	var c Certificate
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
	r.Addresses = FetchIPAddresses(r.Domain, regIP)
	return r, nil
}

func FetchIPAddresses(name string, regIP *regexp.Regexp) []string {
	var ipsList []string

	ips, err := net.LookupIP(name)
	if err != nil || len(ips) == 0 {
		log.Debugf("Could not fetch IP addresses of domain %s", name)
		return ipsList
	}
	for _, j := range ips {
		if regIP.MatchString(j.String()) {
			ipsList = append(ipsList, j.String())
		}
	}
	return ipsList
}

func IsMatchingCert(cert *Result, reg, regIDN *regexp.Regexp) bool {

	domainList := append(cert.SAN, cert.Domain)
	for _, domain := range domainList {
		if isIDN(domain) && regIDN.MatchString(domain) {
			return true
		}
		if reg.MatchString(domain) {
			return true
		}
	}
	return false
}

func isIDN(domain string) bool {
	return strings.HasPrefix(domain, "xn--")
}

func notify(config *Configuration, detailedCert Result) {
	b, _ := json.Marshal(detailedCert)

	if config.SlackWebHookURL != "" {
		go newSlackPayload(detailedCert, config).Post(config)
	} else {
		fmt.Printf("A certificate for '%v' has been issued : %v\n", detailedCert.Domain, string(b))
	}
}

// LoopCheckCerts Loops on messages from source
func LoopCheckCerts(config *Configuration) {
	for {
		conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), certInput)
		defer conn.Close()
		if err != nil {
			log.Warn("Error connecting to certstream! Sleeping a few seconds and reconnecting...")
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
