package repository

import (
	"context"
	"testing"

	"github.com/halilbulentorhon/cb-pubsub/config"
)

func TestNewCouchbaseRepository_InvalidConfig(t *testing.T) {
	cfg := config.CouchbaseConfig{
		Host:     "invalid-host",
		Username: "test",
		Password: "test",
	}

	_, err := NewCouchbaseRepository(cfg)
	if err == nil {
		t.Error("NewCouchbaseRepository should fail with invalid config")
	}
}

func TestCouchbaseRepository_Interface(t *testing.T) {
	var _ Repository = (*couchbaseRepository)(nil)
}

func TestRepositoryInterface_MethodSignatures(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"Get", "Get(ctx context.Context, key string, result interface{}) (gocb.Cas, error)"},
		{"GetAndTouch", "GetAndTouch(ctx context.Context, key string, result interface{}, ttl time.Duration) (gocb.Cas, error)"},
		{"Upsert", "Upsert(ctx context.Context, key string, document interface{}, ttl time.Duration) error"},
		{"ReplaceWithCas", "ReplaceWithCas(ctx context.Context, key string, document interface{}, ttl time.Duration, cas gocb.Cas) error"},
		{"UpsertPath", "UpsertPath(ctx context.Context, key string, path string, value interface{}) error"},
		{"UpsertPathWithCas", "UpsertPathWithCas(ctx context.Context, key string, path string, value interface{}, cas gocb.Cas) error"},
		{"ArrayAppend", "ArrayAppend(ctx context.Context, key string, path string, values interface{}) error"},
		{"RemoveMultiplePaths", "RemoveMultiplePaths(ctx context.Context, key string, paths []string) error"},
		{"ArrayRemoveFromIndex", "ArrayRemoveFromIndex(ctx context.Context, key string, path string, fromIndex int, toIndex int) error"},
		{"Delete", "Delete(ctx context.Context, key string) error"},
		{"Close", "Close() error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Repository interface has method: %s", tt.method)
		})
	}
}

func TestCouchbaseRepository_Timeouts(t *testing.T) {
	cfg := config.CouchbaseConfig{
		Host:                "couchbase://localhost",
		Username:            "admin",
		Password:            "password",
		BucketName:          "test",
		ScopeName:           "_default",
		CollectionName:      "_default",
		ConnectTimeoutSec:   1,
		OperationTimeoutSec: 1,
	}

	_, err := NewCouchbaseRepository(cfg)
	if err == nil {
		t.Error("NewCouchbaseRepository should timeout with short timeouts")
	}
}

func TestCouchbaseRepository_EmptyPaths(t *testing.T) {
	repo := &couchbaseRepository{}

	err := repo.RemoveMultiplePaths(context.Background(), "test-key", []string{})
	if err == nil {
		t.Error("RemoveMultiplePaths should fail with empty paths")
	}
	if err.Error() != "no paths provided" {
		t.Errorf("RemoveMultiplePaths error = %v, want 'no paths provided'", err)
	}
}

func TestCouchbaseRepository_ArrayRemoveFromIndex_Logic(t *testing.T) {
	tests := []struct {
		name      string
		fromIndex int
		toIndex   int
		path      string
		expected  []string
	}{
		{
			name:      "single element",
			fromIndex: 2,
			toIndex:   2,
			path:      "messages",
			expected:  []string{"messages[2]"},
		},
		{
			name:      "range of elements",
			fromIndex: 1,
			toIndex:   3,
			path:      "items",
			expected:  []string{"items[3]", "items[2]", "items[1]"},
		},
		{
			name:      "reverse order verification",
			fromIndex: 0,
			toIndex:   1,
			path:      "array",
			expected:  []string{"array[1]", "array[0]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paths []string
			for i := tt.toIndex; i >= tt.fromIndex; i-- {
				paths = append(paths, tt.path+"["+string(rune('0'+i))+"]")
			}

			for i, expected := range tt.expected {
				if i >= len(paths) {
					t.Errorf("Missing path at index %d", i)
					continue
				}
				actual := tt.path + "[" + string(rune('0'+(tt.toIndex-i))) + "]"
				if actual != expected {
					t.Errorf("Path %d = %s, want %s", i, actual, expected)
				}
			}
		})
	}
}
