package constant

import "time"

const (
	AssignmentDocName = "_pubsub_all"
	SelfDocPrefix     = "_pubsub_instance_"
	MessagesPath      = "messages"
)

const (
	MaxConsecutiveFailures    = 10
	CleanupIntervalMultiplier = 1.5
	MaxCleanupInterval        = 5 * time.Minute
	SelfDocTtlSeconds         = 600
)

const (
	DefaultShutdownTimeout = 10 * time.Second
)

const (
	DefaultSubscribeRetryInitialDelay = 1 * time.Second
	DefaultSubscribeRetryMaxDelay     = 8 * time.Second
	DefaultSubscribeRetryMultiplier   = 2.0
)

const (
	DefaultCleanupRetryInitialDelay = 2 * time.Second
	DefaultCleanupRetryMaxDelay     = 30 * time.Second
	DefaultCleanupRetryMultiplier   = 2.0
)
