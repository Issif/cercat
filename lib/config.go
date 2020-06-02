package lib

import (
	"container/ring"
	"path"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
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
	Buffer          chan *Result
	Homoglyph       map[string]string
}

// GetConfig provides a Configuration
func GetConfig() *Configuration {
	c := &Configuration{
		Homoglyph:     GetHomoglyphMap(),
		PreviousCerts: ring.New(20),
		Buffer:        make(chan *Result, 50),
	}

	configFile := kingpin.Flag("configfile", "config file").Short('c').ExistingFile()
	kingpin.Parse()

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
	if c.Workers < 1 {
		log.Fatal("Workers must be strictly a positive number")
	}

	return c
}
