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
	Workers           int
	SlackWebHookURL   string
	SlackIconURL      string
	SlackUsername     string
	Regexp            string
	ScreenshotsFolder string
	TakeScreenshot    bool
	IgnoreOlderThan   int
	RegIP             string
	RegexpC           *regexp.Regexp
	PreviousCerts     *ring.Ring
	Messages          chan []byte
	Buffer            chan *model.Result
	Homoglyph         map[string]string
	Log               *logrus.Logger
}

// CreateConfig provides a Configuration
func CreateConfig(configFile *string) *Configuration {
	c := &Configuration{
		Workers:       50,
		Homoglyph:     homoglyph.GetHomoglyphMap(),
		PreviousCerts: ring.New(20),
		Messages:      make(chan []byte, 100),
		Buffer:        make(chan *model.Result, 100),
		Log:           logrus.New(),
	}

	c.Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	c.Log.SetOutput(os.Stdout)

	v := viper.New()
	v.SetDefault("SlackWebhookURL", "")
	v.SetDefault("SlackIconURL", "")
	v.SetDefault("SlackUsername", "Cercat")
	v.SetDefault("Regexp", "")
	v.SetDefault("Workers", 20)
	v.SetDefault("TakeScreenshot", false)
	v.SetDefault("ScreenshotsFolder", ".")
	v.SetDefault("IgnoreOlderThan", 10)

	if *configFile != "" {
		d, f := path.Split(*configFile)
		if d == "" {
			d = "."
		}
		v.SetConfigName(f[0 : len(f)-len(filepath.Ext(f))])
		v.AddConfigPath(d)
		err := v.ReadInConfig()
		if err != nil {
			c.Log.Fatalf("[ERROR] : Error when reading config file : %v\n", err)
		}
	}
	v.AutomaticEnv()
	v.Unmarshal(c)

	if c.SlackUsername == "" {
		c.SlackUsername = "Cercat"
	}

	if c.Regexp == "" {
		c.Log.Fatal("Regexp can't be empty")
	}

	reg, err := regexp.Compile(c.Regexp)
	if err != nil {
		c.Log.Fatal("Bad regexp")
	}
	c.RegexpC = reg
	return c
}
