package config

import (
	"cercat/pkg/bitsquatting"
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"cercat/pkg/omission"
	"cercat/pkg/repetition"
	"cercat/pkg/transposition"
	"cercat/pkg/vowelswap"
	"container/ring"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
)

// Configuration represents a configuration element
type Configuration struct {
	Workers               int
	SlackWebHookURL       string
	SlackIconURL          string
	SlackUsername         string
	RegIP                 string
	Domains               []string
	HomoglyphPatterns     map[string]string
	InclusionPatterns     map[string][]string
	BitsquattingPatterns  map[string][]string
	OmissionPatterns      map[string][]string
	RepetitionPatterns    map[string][]string
	TranspositionPatterns map[string][]string
	VowelSwapPatterns     map[string][]string
	PreviousCerts         *ring.Ring
	Messages              chan []byte
	Buffer                chan *model.Result
}

// GetConfig provides a Configuration
func GetConfig(configFile *string) *Configuration {
	c := &Configuration{
		Workers:               50,
		PreviousCerts:         ring.New(20),
		Messages:              make(chan []byte, 50),
		Buffer:                make(chan *model.Result, 50),
		HomoglyphPatterns:     homoglyph.GetHomoglyphMap(),
		OmissionPatterns:      make(map[string][]string),
		RepetitionPatterns:    make(map[string][]string),
		BitsquattingPatterns:  make(map[string][]string),
		TranspositionPatterns: make(map[string][]string),
		VowelSwapPatterns:     make(map[string][]string),
	}

	v := viper.New()
	v.SetDefault("SlackWebhookURL", "")
	v.SetDefault("SlackIconURL", "")
	v.SetDefault("SlackUsername", "Cercat")
	v.SetDefault("Domains", []string{})
	v.SetDefault("Workers", 20)

	if c.SlackUsername == "" {
		c.SlackUsername = "Cercat"
	}

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

	if len(c.Domains) == 0 {
		log.Fatal("Domain list can't be empty")
	}

	for _, domain := range c.Domains {
		p, _ := publicsuffix.PublicSuffix(domain)
		s := strings.Split(strings.ReplaceAll(strings.ReplaceAll(domain, "."+p, ""), "-", ""), ".")
		c.InclusionPatterns[domain] = []string{"s"}
		c.OmissionPatterns[domain] = omission.GetOmissionPatterns(s[len(s)-1])
		c.RepetitionPatterns[domain] = repetition.GetRepetitionPatterns(s[len(s)-1])
		c.BitsquattingPatterns[domain] = bitsquatting.GetBitsquattingPatterns(s[len(s)-1])
		c.TranspositionPatterns[domain] = transposition.GetTranspositionPatterns(s[len(s)-1])
		c.VowelSwapPatterns[domain] = vowelswap.GetVowelSwapPatterns(s[len(s)-1])
	}

	fmt.Printf("%#v\n", c.BitsquattingPatterns)
	return c
}
