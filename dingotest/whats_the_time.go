package dingotest

import (
	"github.com/jonboulle/clockwork"
	"time"
)

type WhatsTheTime struct {
	clock clockwork.Clock
}

func (t *WhatsTheTime) InRFC1123() string {
	return t.clock.Now().Format(time.RFC1123)
}
