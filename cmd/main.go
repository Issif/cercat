package main

import (
	"cercat/config"
	"cercat/lib"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/arl/statsviz"
	"github.com/pkg/errors"
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

	cfg := config.GetConfig(configFile)
	for i := 0; i < cfg.Workers; i++ {
		go lib.CertCheckWorker(cfg.Regexp, &cfg.Homoglyph, cfg.Messages, cfg.Buffer)
	}
	go lib.Notifier(cfg)
	lib.LoopCertStream(cfg.Messages)
}
