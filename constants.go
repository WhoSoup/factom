package factom

import "time"

const (
	RetryAttempts      int           = 120
	TimeBetweenRetries time.Duration = time.Second
)
