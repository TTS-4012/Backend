package runner

import "time"

const (
	NatsTimeout = time.Minute
	TimeLimit   = time.Second * 10
	MemoryLimit = 256 * 1024 * 1024
)
