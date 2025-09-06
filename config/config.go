package config

type PubSubConfig struct {
	PollIntervalSeconds    int             `json:"pollIntervalSeconds"`
	SelfDocTTLSeconds      int             `json:"selfDocTTLSeconds"`
	CleanupIntervalSeconds int             `json:"cleanupIntervalSeconds"`
	SubscribeRetryAttempts int             `json:"subscribeRetryAttempts"`
	CleanupRetryAttempts   int             `json:"cleanupRetryAttempts"`
	ShutdownTimeoutSec     int             `json:"shutdownTimeoutSec"`
	InitTimeoutSec         int             `json:"initTimeoutSec"`
	CouchbaseConfig        CouchbaseConfig `json:"couchbaseConfig"`
}

type CouchbaseConfig struct {
	Host                string `json:"host"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	BucketName          string `json:"bucket"`
	ScopeName           string `json:"scope"`
	CollectionName      string `json:"collection"`
	ConnectTimeoutSec   int    `json:"connectTimeoutSec"`
	OperationTimeoutSec int    `json:"operationTimeoutSec"`
}

func (c *PubSubConfig) ApplyDefaults() {
	if c.PollIntervalSeconds <= 0 {
		c.PollIntervalSeconds = 1
	}
	if c.SelfDocTTLSeconds <= 0 {
		c.SelfDocTTLSeconds = 30
	}
	if c.CleanupIntervalSeconds <= 0 {
		c.CleanupIntervalSeconds = 15
	}
	if c.SubscribeRetryAttempts <= 0 {
		c.SubscribeRetryAttempts = 3
	}
	if c.CleanupRetryAttempts <= 0 {
		c.CleanupRetryAttempts = 5
	}
	if c.CouchbaseConfig.ConnectTimeoutSec <= 0 {
		c.CouchbaseConfig.ConnectTimeoutSec = 10
	}
	if c.CouchbaseConfig.OperationTimeoutSec <= 0 {
		c.CouchbaseConfig.OperationTimeoutSec = 5
	}
	if c.ShutdownTimeoutSec <= 0 {
		c.ShutdownTimeoutSec = 10
	}
	if c.InitTimeoutSec <= 0 {
		c.InitTimeoutSec = 30
	}
}
