package certstream

import (
	"cercat/config"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// the websocket stream from calidog
const certInput = "wss://certstream.calidog.io"

// StartLoopCertStream gathers messages from CertStream
func StartLoopCertStream(cfg *config.Configuration) {
	dial := ws.Dialer{
		ReadBufferSize:  2048,
		WriteBufferSize: 128,
		Timeout:         5 * time.Second,
	}
	for {
		conn, _, _, err := dial.Dial(context.Background(), certInput)
		if err != nil {
			cfg.Log.Warn("Error connecting to CertStream! Sleeping a few seconds and reconnecting")
			time.Sleep(1 * time.Second)
			continue
		}
		for {
			msg, _, err := wsutil.ReadServerData(conn)
			// fmt.Println(string(msg))
			if err != nil {
				cfg.Log.Warn(fmt.Sprintf("Error reading message from CertStream (%v)", strings.TrimSuffix(err.Error(), " ")))
				break
			}
			cfg.Messages <- msg
		}
		conn.Close()
	}
}
