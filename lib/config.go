package lib

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type Configuration struct {
	Workers         int
	SlackWebHookURL string
	SlackIconURL    string
	SlackUsername   string
	DomainName      string
	RegIP           string
	Regexp          string
	RegIDN          string
	DisplayErrors   string
}

const RegStrIP = `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`

func GetConfig() *Configuration {
	c := &Configuration{
		RegIP: RegStrIP,
	}

	configFile := kingpin.Flag("configfile", "config file").Short('c').ExistingFile()
	kingpin.Parse()

	v := viper.New()
	v.SetDefault("SlackWebhookURL", "")
	v.SetDefault("SlackIconURL", "")
	v.SetDefault("SlackUsername", "Cercat")
	v.SetDefault("DomainName", "")
	v.SetDefault("Regexp", "")
	v.SetDefault("Workers", 20)
	v.SetDefault("DisplayErrors", "false")

	if *configFile != "" {
		d, f := path.Split(*configFile)
		if d == "" {
			d = "."
		}
		v.SetConfigName(f[0 : len(f)-len(filepath.Ext(f))])
		v.AddConfigPath(d)
		err := v.ReadInConfig()
		if err != nil {
			log.Fatalf("[ERROR] : Error when reading config file : %v\n", err)
		}
	}
	v.AutomaticEnv()
	v.Unmarshal(c)

	if c.SlackUsername == "" {
		c.SlackUsername = "Cercat"
	}
	if c.DisplayErrors == "" || c.DisplayErrors == "false" {
		log.SetLevel(log.DebugLevel)
	}
	if c.Regexp == "" {
		log.Fatal("Regexp can't be empty")
	}
	if c.DomainName == "" {
		log.Fatal("Specify the domain name to monitor for IDN homographs")
	}
	if _, err := regexp.Compile(c.Regexp); err != nil {
		log.Fatal("Bad regexp")
	}
	if c.Workers < -1 {
		log.Fatal("Workers must be strictly a positive number")
	}

	c.RegIDN = BuildIDNRegex(c.DomainName)

	return c
}

func BuildIDNRegex(name string) string {
	if len(name) < 2 {
		return ""
	}
	// Can detect up to two unicode characters in the domain name.
	// To adjust according to false positive rate & name length
	return fmt.Sprintf("[%s]{%d,%d}", strings.ToLower(name), len(name)-2, len(name)-1)
}
