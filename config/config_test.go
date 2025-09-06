package config

import (
	"testing"
)

func TestPubSubConfig_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    PubSubConfig
		expected PubSubConfig
	}{
		{
			name:  "empty config gets all defaults",
			input: PubSubConfig{},
			expected: PubSubConfig{
				PollIntervalSeconds:    1,
				CleanupIntervalSeconds: 15,
				SubscribeRetryAttempts: 3,
				CleanupRetryAttempts:   5,
				CouchbaseConfig: CouchbaseConfig{
					ConnectTimeoutSec:   10,
					OperationTimeoutSec: 5,
				},
			},
		},
		{
			name: "partial config keeps existing values",
			input: PubSubConfig{
				PollIntervalSeconds: 5,
				CouchbaseConfig: CouchbaseConfig{
					Host:       "localhost",
					Username:   "admin",
					BucketName: "test",
				},
			},
			expected: PubSubConfig{
				PollIntervalSeconds:    5,
				CleanupIntervalSeconds: 15,
				SubscribeRetryAttempts: 3,
				CleanupRetryAttempts:   5,
				CouchbaseConfig: CouchbaseConfig{
					Host:                "localhost",
					Username:            "admin",
					BucketName:          "test",
					ConnectTimeoutSec:   10,
					OperationTimeoutSec: 5,
				},
			},
		},
		{
			name: "negative values get defaults",
			input: PubSubConfig{
				PollIntervalSeconds:    -1,
				CleanupIntervalSeconds: 0,
				SubscribeRetryAttempts: -2,
				CleanupRetryAttempts:   0,
				CouchbaseConfig: CouchbaseConfig{
					ConnectTimeoutSec:   -1,
					OperationTimeoutSec: 0,
				},
			},
			expected: PubSubConfig{
				PollIntervalSeconds:    1,
				CleanupIntervalSeconds: 15,
				SubscribeRetryAttempts: 3,
				CleanupRetryAttempts:   5,
				CouchbaseConfig: CouchbaseConfig{
					ConnectTimeoutSec:   10,
					OperationTimeoutSec: 5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.input
			cfg.ApplyDefaults()

			if cfg.PollIntervalSeconds != tt.expected.PollIntervalSeconds {
				t.Errorf("PollIntervalSeconds = %d, want %d", cfg.PollIntervalSeconds, tt.expected.PollIntervalSeconds)
			}
			if cfg.CleanupIntervalSeconds != tt.expected.CleanupIntervalSeconds {
				t.Errorf("CleanupIntervalSeconds = %d, want %d", cfg.CleanupIntervalSeconds, tt.expected.CleanupIntervalSeconds)
			}
			if cfg.SubscribeRetryAttempts != tt.expected.SubscribeRetryAttempts {
				t.Errorf("SubscribeRetryAttempts = %d, want %d", cfg.SubscribeRetryAttempts, tt.expected.SubscribeRetryAttempts)
			}
			if cfg.CleanupRetryAttempts != tt.expected.CleanupRetryAttempts {
				t.Errorf("CleanupRetryAttempts = %d, want %d", cfg.CleanupRetryAttempts, tt.expected.CleanupRetryAttempts)
			}
			if cfg.CouchbaseConfig.ConnectTimeoutSec != tt.expected.CouchbaseConfig.ConnectTimeoutSec {
				t.Errorf("ConnectTimeoutSec = %d, want %d", cfg.CouchbaseConfig.ConnectTimeoutSec, tt.expected.CouchbaseConfig.ConnectTimeoutSec)
			}
			if cfg.CouchbaseConfig.OperationTimeoutSec != tt.expected.CouchbaseConfig.OperationTimeoutSec {
				t.Errorf("OperationTimeoutSec = %d, want %d", cfg.CouchbaseConfig.OperationTimeoutSec, tt.expected.CouchbaseConfig.OperationTimeoutSec)
			}
		})
	}
}

func TestCouchbaseConfig_Fields(t *testing.T) {
	cfg := CouchbaseConfig{
		Host:                "couchbase://localhost",
		Username:            "testuser",
		Password:            "testpass",
		BucketName:          "testbucket",
		ScopeName:           "testscope",
		CollectionName:      "testcollection",
		ConnectTimeoutSec:   15,
		OperationTimeoutSec: 10,
	}

	if cfg.Host != "couchbase://localhost" {
		t.Errorf("Host = %s, want couchbase://localhost", cfg.Host)
	}
	if cfg.Username != "testuser" {
		t.Errorf("Username = %s, want testuser", cfg.Username)
	}
	if cfg.BucketName != "testbucket" {
		t.Errorf("BucketName = %s, want testbucket", cfg.BucketName)
	}
}
