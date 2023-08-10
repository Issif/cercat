package worker

import (
	"cercat/config"
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"cercat/pkg/screenshot"
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"time"

	tld "github.com/jpillora/go-tld"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"golang.org/x/net/idna"
)

// RunCertCheckWorker parses certificates and raises alert if matches config
func RunCertCheckWorker(cfg *config.Configuration) {
	for {
		msg := <-cfg.Messages
		result, err := parseResultCertificate(msg)
		if err != nil {
			cfg.Log.Warnf("Error parsing message: %s", err)
			continue
		}
		if result == nil {
			continue
		}
		if !isMatchingCert(cfg, result) {
			continue
		}
		var domain string
		domains := []string{result.Domain}
		domains = append(domains, result.SAN...)
		for _, i := range domains {
			if strings.Contains(i, "*") {
				continue
			}
			domain = i
			break
		}
		if domain != "" && !strings.Contains(domain, "*") {
			result.Registrar, result.CreationDate = getWhoIs(result.Domain, cfg)
			if result.CreationDate == "" {
				continue
			}
			creationDate := strings.Split(result.CreationDate, "-")
			if len(creationDate) < 3 {
				continue
			}
			if time.Since(date(creationDate[0], creationDate[1], creationDate[2])).Hours()/24 > float64(cfg.IgnoreOlderThan) {
				continue
			}
			result.Addresses = fetchIPv4Addresses(domain, cfg)
			if cfg.TakeScreenshot {
				result.Screenshot = screenshot.TakeScreenshot(domain, cfg)
			}
			cfg.Buffer <- result
		}
	}
}

// parseResultCertificate parses certificate details
func parseResultCertificate(msg []byte) (*model.Result, error) {
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

// fetchIPv4Addresses resolves domain to get IP
func fetchIPv4Addresses(domain string, cfg *config.Configuration) []string {
	var ipsList []string

	ips, err := net.LookupIP(domain)
	if err != nil || len(ips) == 0 {
		cfg.Log.Debugf("Could not fetch IPv4 addresses of domain %s", domain)
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

// isMatchingCert checks if certificate matches the regexp
func isMatchingCert(cfg *config.Configuration, result *model.Result) bool {
	domainList := append(result.SAN, result.Domain)
	for _, domain := range domainList {
		if isIDN(domain) {
			result.IDN, _ = idna.ToUnicode(domain)
			domain = homoglyph.ReplaceHomoglyph(result.IDN, cfg.Homoglyph)
		}
		if cfg.RegexpC.MatchString(domain) {
			return true
		}
	}
	return false
}

// getWhoIs gets domain WHOIS details
func getWhoIs(domain string, cfg *config.Configuration) (registrar, creationDate string) {
	u, err := tld.Parse("https://" + domain)
	if u == nil || err != nil {
		return "", ""
	}
	if u.Domain == "" || u.TLD == "" {
		cfg.Log.Warningf("Could not get WHOIS details of domain %s", domain)
		return "", ""
	}
	whoisRaw, err := whois.Whois(u.Domain + "." + u.TLD)
	if err != nil {
		cfg.Log.Warningf("Could not get WHOIS details of domain %s: %v", domain, err)
		return "", ""
	}
	result, err := whoisparser.Parse(whoisRaw)
	if err != nil {
		cfg.Log.Warningf("Could not parse WHOIS details of domain %s: %v", domain, err)
		return "", ""
	}
	if result.Registrar == nil || result.Domain == nil {
		return "", ""
	}
	return result.Registrar.Name, strings.Split(result.Domain.CreatedDate, "T")[0]
}

// date return a time.Time
func date(year, month, day string) time.Time {
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)
	d, _ := strconv.Atoi(day)
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}
