package config

import (
	"cercat/pkg/model"
	"container/ring"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
)

// Configuration represents a configuration element
type Configuration struct {
	Workers         int
	SlackWebHookURL string
	SlackIconURL    string
	SlackUsername   string
	RegIP           string
	Domains         []string
	Regexp          map[string][]*regexp.Regexp
	PreviousCerts   *ring.Ring
	Messages        chan []byte
	Buffer          chan *model.Result
	Homoglyphs      map[string]string
}

// GetConfig provides a Configuration
func GetConfig(configFile *string) *Configuration {
	c := &Configuration{
		Workers:       50,
		PreviousCerts: ring.New(20),
		Messages:      make(chan []byte, 50),
		Buffer:        make(chan *model.Result, 50),
	}

	v := viper.New()
	v.SetDefault("SlackWebhookURL", "")
	v.SetDefault("SlackIconURL", "")
	v.SetDefault("SlackUsername", "Cercat")
	v.SetDefault("Domains", []string{})
	v.SetDefault("Workers", 20)

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

	if len(c.Domains) == 0 {
		log.Fatal("Domain list can't be empty")
	}

	c.Regexp = make(map[string][]*regexp.Regexp, len(c.Domains))
	for _, i := range c.Domains {
		p, _ := publicsuffix.PublicSuffix(i)
		s := strings.Split(strings.Replace(i, "."+p, "", -1), ".")
		c.Regexp[i] = GetRegexpList(s[len(s)-1])
	}
	return c
}

// GetRegexpList generates a list of Regexp
func GetRegexpList(pattern string) []*regexp.Regexp {
	list := []*regexp.Regexp{}
	for i := 0; i < len([]rune(pattern)); i++ {
		// s := ".*"
		var s string
		for k, l := range []rune(pattern) {
			if k == i {
				s += "[a-z]{1}"
			} else {
				s += string(l)
			}
		}
		// s += ".*"
		r, _ := regexp.Compile(s)
		list = append(list, r)
	}
	return list
}
