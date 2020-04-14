package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type configuration struct {
	Workers         int
	SlackWebHookURL string
	SlackIconURL    string
	SlackUsername   string
	Regexp          string
}

func getConfig() *configuration {
	c := &configuration{}

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
			log.Printf("[ERROR] : Error when reading config file : %v\n", err)
			os.Exit(1)
		}
	}
	v.AutomaticEnv()
	v.Unmarshal(c)

	if c.SlackUsername == "" {
		c.SlackUsername = "Cercat"
	}
	if c.Regexp == "" {
		log.Println("[ERROR] : Regexp can't be empty")
		os.Exit(1)
	}
	if _, err := regexp.Compile(c.Regexp); err != nil {
		log.Println("[ERROR] : Bad regexp")
		os.Exit(1)
	}
	if c.Workers < -1 {
		log.Println("[ERROR] : Workers must be strictly a positive number")
		os.Exit(1)
	}

	return c
}
