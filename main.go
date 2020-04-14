package main

import (
	"encoding/json"
	"log"
	"net"
	"regexp"

	"time"

	"github.com/gorilla/websocket"
)

type result struct {
	Domain    string   `json:"domain"`
	SAN       []string `json:"SAN"`
	Issuer    string   `json:"issuer"`
	Addresses []string `json:"Addresses"`
}

type certificate struct {
	MessageType string `json:"message_type"`
	Data        data   `json:"data"`
}

type data struct {
	UpdateType string            `json:"update_type"`
	LeafCert   leafCert          `json:"leaf_cert"`
	Chain      []leafCert        `json:"chain"`
	CertIndex  float32           `json:"cert_index"`
	Seen       float32           `json:"seen"`
	Source     map[string]string `json:"source"`
}

type leafCert struct {
	Subject      map[string]string `json:"subject"`
	Extensions   map[string]string `json:"extensions"`
	NotBefore    float32           `json:"not_before"`
	NotAfter     float32           `json:"not_after"`
	SerialNumber string            `json:"serial_number"`
	FingerPrint  string            `json:"fingerprint"`
	AsDer        string            `json:"as_der"`
	AllDomains   []string          `json:"all_domains"`
}

var config *configuration

func init() {
	config = getConfig()
}

func main() {
	msgChan := make(chan []byte, 10)
	for i := 0; i < config.Workers; i++ {
		go certCheckWorker(msgChan)
	}

	for {
		ws, _, err := websocket.DefaultDialer.Dial("wss://certstream.calidog.io", nil)
		defer ws.Close()

		if err != nil {
			log.Println("[INFO]  : Error connecting to certstream! Sleeping a few seconds and reconnecting...")
			time.Sleep(1 * time.Second)
			continue
		}
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Println("[ERROR] : Error reading message")
				continue
			}
			msgChan <- msg
		}
	}
}

func certCheckWorker(msgChan <-chan []byte) {
	reg, _ := regexp.Compile(config.Regexp)
	regIP, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

	var c certificate
	for {
		msg := <-msgChan
		json.Unmarshal(msg, &c)
		if c.MessageType == "heartbeat" {
			continue
		}
		if reg.MatchString(c.Data.LeafCert.Subject["CN"]) {
			r := result{
				Domain:    c.Data.LeafCert.Subject["CN"],
				Issuer:    c.Data.Chain[0].Subject["O"],
				SAN:       c.Data.LeafCert.AllDomains,
				Addresses: []string{"N/A"},
			}
			for _, i := range c.Data.LeafCert.AllDomains {
				ips, err := net.LookupIP(i)
				if err != nil {
					continue
				}
				ipsList := []string{}
				if len(ips) != 0 {
					for _, j := range ips {
						if regIP.MatchString(j.String()) {
							ipsList = append(ipsList, j.String())
						}
					}
					r.Addresses = ipsList
				}
				break
			}
			b, _ := json.Marshal(r)
			log.Printf("[INFO]  : A certificate for '%v' has been issued : %v\n", r.Domain, string(b))
			if config.SlackWebHookURL != "" {
				go newSlackPayload(r).Post()
			}
		}
	}
}
