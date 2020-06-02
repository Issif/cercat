package main

import (
	"cercat/lib"
	_ "expvar"
	"net/http"
	_ "net/http/pprof"
)

var config *lib.Configuration

func init() {
	config = lib.GetConfig()
}

func main() {
	go http.ListenAndServe("localhost:6060", nil)
	for i := 0; i < config.Workers; i++ {
		go lib.CertCheckWorker(config)
	}
	go lib.Notifier(config)
	lib.LoopCertStream(config)
}
