package nats

import (
	"time"

	"github.com/nats-io/nats.go"
)

// Connect opens a NATS connection that keeps reconnecting forever.
func Connect(url string) (*nats.Conn, error) {
	return nats.Connect(url,
		nats.Name("omni-chan"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
}
