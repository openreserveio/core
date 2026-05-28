package bus

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	"time"
)

type BusMessage struct {
	ID                string `json:"id"`
	Data              []byte `json:"data"`
	IsReply           bool   `json:"isReply"`
	ReplyStatus       int    `json:"replyStatus"`
	ReplyStatusDetail string `json:"replyStatusDetail"`
}

func DoMonitoringLoop(nc *nats.Conn, natsUrl string) {

	for {
		time.Sleep(5 * time.Second)
		status := nc.Status()
		if status == nats.CONNECTED {
			// log.Info("Connected to NATS")
			continue
		} else {
			log.Info("Not connected to NATS.  Connecting")
			newconn, err := nats.Connect(natsUrl)
			if err != nil {
				log.Error("Error connecting to NATS: %v", err)
				continue
			}
			nc = newconn
		}
	}

}
