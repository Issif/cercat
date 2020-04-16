package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"regexp"

	"time"

	"net/http"
	_ "net/http/pprof"

	_ "expvar"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
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
	go http.ListenAndServe("localhost:6060", nil)

	msgChan := make(chan []byte, 10)
	for i := 0; i < config.Workers; i++ {
		go certCheckWorker(msgChan)
	}

	for {
		conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "wss://certstream.calidog.io")

		if err != nil {
			if config.DisplayErrors == "true" {
				log.Println("[ERROR] : Error connecting to certstream! Sleeping a few seconds and reconnecting...")
			}
			conn.Close()
			time.Sleep(1 * time.Second)
			continue
		}
		for {
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				if config.DisplayErrors == "true" {
					log.Println("[ERROR] : Error reading message from CertStream")
				}
				break
			}
			msgChan <- msg
		}
		conn.Close()
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
			ips, _ := net.LookupIP(r.Domain)
			if len(ips) != 0 {
				ipsList := []string{}
				for _, j := range ips {
					if regIP.MatchString(j.String()) {
						ipsList = append(ipsList, j.String())
					}
				}
				r.Addresses = ipsList
			}
			b, _ := json.Marshal(r)
			log.Printf("[INFO]  : A certificate for '%v' has been issued : %v\n", r.Domain, string(b))
			if config.SlackWebHookURL != "" {
				go newSlackPayload(r).Post()
			}
		}
	}
}
