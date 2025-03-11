package main

import (
	"cercat/config"
	"cercat/pkg/certstream"
	"cercat/pkg/model"
	"cercat/pkg/slack"
	"cercat/pkg/worker"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/arl/statsviz"
	"github.com/pkg/errors"
)

func init() {
	statsviz.RegisterDefault()
	go http.ListenAndServe("localhost:6060", nil)
}

func main() {
	a := kingpin.New(filepath.Base(os.Args[0]), "")
	configFile := a.Flag("configfile", "config file").Short('c').ExistingFile()
	a.HelpFlag.Short('h')

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	cfg := config.CreateConfig(configFile)
	// fmt.Printf("%#v\n", cfg)
	for i := 0; i < cfg.Workers; i++ {
		go worker.RunCertCheckWorker()
	}
	go runNotifierWorker()
	certstream.StartLoopCertStream()
}

// runNotifierWorker is a worker that receives cert, depduplicates and sends to Slack the event
func runNotifierWorker() {
	cfg := config.GetConfig()

	for {
		result := <-cfg.Buffer

		// j, _ := json.Marshal(result)
		cfg.Log.Infof("A %v phishing website '%v' has been detected\n", result.Status, result.Domain)
		if cfg.SlackWebHookURL != "" {
			go func(c *config.Configuration, r *model.Result) {
				slack.NewPayload(c, result).Post(c)
			}(cfg, result)
		}
	}
}
