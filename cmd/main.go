package main

import (
	"cercat/config"
	"cercat/pkg/certstream"
	"cercat/pkg/model"
	"cercat/pkg/slack"
	"cercat/pkg/worker"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	_ "net/http/pprof"

	"github.com/arl/statsviz"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
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
	for i := 0; i < cfg.Workers; i++ {
		go worker.RunCertCheckWorker(cfg)
	}
	go runNotifierWorker(cfg)
	certstream.StartLoopCertStream(cfg)
}

// runNotifierWorker is a worker that receives cert, depduplicates and sends to Slack the event
func runNotifierWorker(cfg *config.Configuration) {
	for {
		result := <-cfg.Buffer
		duplicate := false
		cfg.PreviousCerts.Do(func(d interface{}) {
			if result.Domain == d {
				duplicate = true
			}
		})
		if !duplicate {
			cfg.PreviousCerts = cfg.PreviousCerts.Prev()
			cfg.PreviousCerts.Value = result.Domain
			j, _ := json.Marshal(result)
			log.Infof("A certificate for '%v' has been issued : %v\n", result.Domain, string(j))
			if cfg.SlackWebHookURL != "" {
				go func(c *config.Configuration, r *model.Result) {
					slack.NewPayload(c, result).Post(c)
				}(cfg, result)
			}
		}
	}
}
