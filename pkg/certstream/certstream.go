package certstream

import (
	"cercat/config"
	"context"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
)

// the websocket stream from calidog
const certInput = "wss://certstream.calidog.io"

// StartLoopCertStream gathers messages from CertStream
func StartLoopCertStream(cfg *config.Configuration) {
	dial := ws.Dialer{
		ReadBufferSize:  8192,
		WriteBufferSize: 512,
		Timeout:         1 * time.Second,
	}
	for {
		// conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), certInput)
		conn, _, _, err := dial.Dial(context.Background(), certInput)
		if err != nil {
			log.Warn(err)
			log.Warn("Error connecting to CertStream! Sleeping a few seconds and reconnecting...")
			time.Sleep(1 * time.Second)
			// conn.Close()
			continue
		}
		for {
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				log.Warn("Error reading message from CertStream")
				break
			}
			cfg.Messages <- msg
		}
		conn.Close()
	}
}
