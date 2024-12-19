package common

import "time"

const (
	Version                   = "1.37.0"
	DefaultURL                = "nats://127.0.0.1:4222"
	DefaultPort               = 4222
	DefaultMaxReconnect       = 60
	DefaultReconnectWait      = 2 * time.Second
	DefaultReconnectJitter    = 100 * time.Millisecond
	DefaultReconnectJitterTLS = time.Second
	DefaultTimeout            = 6 * time.Second
)
