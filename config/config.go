package config

import (
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"container/ring"
	"path"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration represents a configuration element
type Configuration struct {
	Workers         int
	SlackWebHookURL string
	SlackIconURL    string
	SlackUsername   string
	RegIP           string
	Regexp          string
	PreviousCerts   *ring.Ring
	Messages        chan []byte
	Buffer          chan *model.Result
	Homoglyph       map[string]string
}

// GetConfig provides a Configuration
func GetConfig(configFile *string) *Configuration {
	c := &Configuration{
		Workers:       50,
		Homoglyph:     homoglyph.GetHomoglyphMap(),
		PreviousCerts: ring.New(20),
		Messages:      make(chan []byte, 50),
		Buffer:        make(chan *model.Result, 50),
	}

	v := viper.New()
	v.SetDefault("SlackWebhookURL", "")
	v.SetDefault("SlackIconURL", "")
	v.SetDefault("SlackUsername", "Cercat")
	v.SetDefault("Regexp", "")
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

	if c.Regexp == "" {
		log.Fatal("Regexp can't be empty")
	}

	if _, err := regexp.Compile(c.Regexp); err != nil {
		log.Fatal("Bad regexp")
	}

	return c
}
