package repository

import (
	"context"
	"github.com/couchbase/gocb/v2"
	"time"
)

type Repository interface {
	Get(ctx context.Context, key string, result interface{}) (gocb.Cas, error)
	GetAndTouch(ctx context.Context, key string, result interface{}, ttl time.Duration) (gocb.Cas, error)
	Upsert(ctx context.Context, key string, document interface{}, ttl time.Duration) error
	ReplaceWithCas(ctx context.Context, key string, document interface{}, ttl time.Duration, cas gocb.Cas) error
	UpsertPath(ctx context.Context, key string, path string, value interface{}) error
	UpsertPathWithCas(ctx context.Context, key string, path string, value interface{}, cas gocb.Cas) error
	ArrayAppend(ctx context.Context, key string, path string, values interface{}) error
	RemoveMultiplePaths(ctx context.Context, key string, paths []string) error
	ArrayRemoveFromIndex(ctx context.Context, key string, path string, fromIndex int, toIndex int) error
	Delete(ctx context.Context, key string) error
	Close() error
}
