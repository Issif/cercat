package config

import (
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"container/ring"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration represents a configuration element
type Configuration struct {
	Workers             int
	SlackWebHookURL     string
	SlackIconURL        string
	SlackUsername       string
	Regexp              string
	ScreenshotsFolder   string
	TakeScreenshot      bool
	IgnoreOlderThan     int
	WrongLinksThreshold int
	RegIP               string
	RegexpC             *regexp.Regexp
	PreviousCerts       *ring.Ring
	Messages            chan []byte
	Buffer              chan *model.Result
	Homoglyph           map[string]string
	Log                 *logrus.Logger
}

var config *Configuration

func init() {
	config = new(Configuration)
}

func GetConfig() *Configuration {
	return config
}

// CreateConfig provides a Configuration
func CreateConfig(configFile *string) *Configuration {
	config = &Configuration{
		Workers:       50,
		Homoglyph:     homoglyph.GetHomoglyphMap(),
		PreviousCerts: ring.New(200),
		Messages:      make(chan []byte, 100),
		Buffer:        make(chan *model.Result, 100),
		Log:           logrus.New(),
	}

	config.Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	config.Log.SetOutput(os.Stdout)

	v := viper.New()
	v.SetDefault("SlackWebhookURL", "")
	v.SetDefault("SlackIconURL", "")
	v.SetDefault("SlackUsername", "Cercat")
	v.SetDefault("Regexp", "")
	v.SetDefault("Workers", 20)
	v.SetDefault("TakeScreenshot", false)
	v.SetDefault("ScreenshotsFolder", ".")
	v.SetDefault("IgnoreOlderThan", 10)
	v.SetDefault("WrongLinksThreshold", 20)

	if *configFile != "" {
		d, f := path.Split(*configFile)
		if d == "" {
			d = "."
		}
		v.SetConfigName(f[0 : len(f)-len(filepath.Ext(f))])
		v.AddConfigPath(d)
		err := v.ReadInConfig()
		if err != nil {
			config.Log.Fatalf("[ERROR] : Error when reading config file : %v\n", err)
		}
	}
	v.AutomaticEnv()
	v.Unmarshal(config)

	if config.SlackUsername == "" {
		config.SlackUsername = "Cercat"
	}

	if config.Regexp == "" {
		config.Log.Fatal("Regexp can't be empty")
	}

	reg, err := regexp.Compile(config.Regexp)
	if err != nil {
		config.Log.Fatal("Bad regexp")
	}
	config.RegexpC = reg
	return config
}
