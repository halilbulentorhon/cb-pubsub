package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/halilbulentorhon/cb-pubsub/config"
)

type couchbaseRepository struct {
	cluster    *gocb.Cluster
	collection *gocb.Collection
}

func (r *couchbaseRepository) Get(ctx context.Context, key string, result interface{}) (gocb.Cas, error) {
	getResult, err := r.collection.Get(key, &gocb.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get document with key %s: %w", key, err)
	}

	err = getResult.Content(result)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal document with key %s: %w", key, err)
	}

	return getResult.Cas(), nil
}

func (r *couchbaseRepository) GetAndTouch(ctx context.Context, key string, result interface{}, ttl time.Duration) (gocb.Cas, error) {
	getResult, err := r.collection.GetAndTouch(key, ttl, &gocb.GetAndTouchOptions{
		Context: ctx,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get document with key %s: %w", key, err)
	}

	err = getResult.Content(result)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal document with key %s: %w", key, err)
	}

	return getResult.Cas(), nil
}

func (r *couchbaseRepository) Upsert(ctx context.Context, key string, document interface{}, ttl time.Duration) error {
	opts := &gocb.UpsertOptions{
		Context: ctx,
	}
	if ttl > 0 {
		opts.Expiry = ttl
	}

	_, err := r.collection.Upsert(key, document, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert document with key '%s': %w", key, err)
	}

	return nil
}

func (r *couchbaseRepository) ReplaceWithCas(ctx context.Context, key string, document interface{}, ttl time.Duration, cas gocb.Cas) error {
	opts := &gocb.ReplaceOptions{
		Cas:     cas,
		Context: ctx,
	}

	if ttl > 0 {
		opts.Expiry = ttl
	}

	_, err := r.collection.Replace(key, document, opts)
	if err != nil {
		return fmt.Errorf("failed to replace document with key '%s' and cas '%d': %w", key, cas, err)
	}

	return nil
}

func (r *couchbaseRepository) UpsertPath(ctx context.Context, key string, path string, value interface{}) error {
	_, err := r.collection.MutateIn(key, []gocb.MutateInSpec{
		gocb.UpsertSpec(path, value, &gocb.UpsertSpecOptions{})}, &gocb.MutateInOptions{
		StoreSemantic:  gocb.StoreSemanticsUpsert,
		PreserveExpiry: true,
		Context:        ctx,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert path '%s' in document with key '%s': %w", path, key, err)
	}

	return nil
}

func (r *couchbaseRepository) UpsertPathWithCas(ctx context.Context, key string, path string, value interface{}, cas gocb.Cas) error {
	_, err := r.collection.MutateIn(key, []gocb.MutateInSpec{
		gocb.UpsertSpec(path, value, &gocb.UpsertSpecOptions{}),
	}, &gocb.MutateInOptions{
		StoreSemantic: gocb.StoreSemanticsUpsert,
		Cas:           cas,
		Context:       ctx,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert path '%s' with CAS in document with key '%s': %w", path, key, err)
	}

	return nil
}

func (r *couchbaseRepository) ArrayAppend(ctx context.Context, key string, path string, values interface{}) error {
	_, err := r.collection.MutateIn(key, []gocb.MutateInSpec{
		gocb.ArrayAppendSpec(path, values, &gocb.ArrayAppendSpecOptions{}),
	}, &gocb.MutateInOptions{
		PreserveExpiry: true,
		Context:        ctx,
	})

	if err != nil {
		return fmt.Errorf("failed to append to array at path '%s' in document with key '%s': %w", path, key, err)
	}

	return nil
}

func (r *couchbaseRepository) RemoveMultiplePaths(ctx context.Context, key string, paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths provided")
	}

	specs := make([]gocb.MutateInSpec, len(paths))
	for i, path := range paths {
		specs[i] = gocb.RemoveSpec(path, &gocb.RemoveSpecOptions{})
	}

	_, err := r.collection.MutateIn(key, specs, &gocb.MutateInOptions{
		PreserveExpiry: true,
		Context:        ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to remove paths %v from document with key '%s': %w", paths, key, err)
	}

	return nil
}

func (r *couchbaseRepository) ArrayRemoveFromIndex(ctx context.Context, key string, path string, fromIndex int, toIndex int) error {
	specs := make([]gocb.MutateInSpec, 0, toIndex-fromIndex+1)
	for i := toIndex; i >= fromIndex; i-- {
		fullPath := fmt.Sprintf("%s[%d]", path, i)
		specs = append(specs, gocb.RemoveSpec(fullPath, &gocb.RemoveSpecOptions{}))
	}

	_, err := r.collection.MutateIn(key, specs, &gocb.MutateInOptions{
		StoreSemantic:  gocb.StoreSemanticsReplace,
		PreserveExpiry: true,
		Context:        ctx,
	})

	if err != nil {
		return fmt.Errorf("failed to remove elements from index %d to %d from array at path '%s' in document with key '%s': %w", fromIndex, toIndex, path, key, err)
	}

	return nil
}

func (r *couchbaseRepository) Delete(ctx context.Context, key string) error {
	_, err := r.collection.Remove(key, &gocb.RemoveOptions{
		Context: ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to delete document with key %s: %w", key, err)
	}

	return nil
}

func (r *couchbaseRepository) Close() error {
	if r.cluster != nil {
		return r.cluster.Close(nil)
	}
	return nil
}

func NewCouchbaseRepository(cfg config.CouchbaseConfig) (Repository, error) {
	connectTimeout := time.Duration(cfg.ConnectTimeoutSec) * time.Second
	operationTimeout := time.Duration(cfg.OperationTimeoutSec) * time.Second

	opts := gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		},
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: connectTimeout,
			KVTimeout:      operationTimeout,
		},
	}

	cluster, err := gocb.Connect(cfg.Host, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to couchbase cluster: %w", err)
	}

	err = cluster.WaitUntilReady(connectTimeout, nil)
	if err != nil {
		return nil, fmt.Errorf("cluster not ready: %w", err)
	}

	bucket := cluster.Bucket(cfg.BucketName)
	err = bucket.WaitUntilReady(operationTimeout, nil)
	if err != nil {
		return nil, fmt.Errorf("bucket not ready: %w", err)
	}

	collection := bucket.Scope(cfg.ScopeName).Collection(cfg.CollectionName)

	return &couchbaseRepository{
		cluster:    cluster,
		collection: collection,
	}, nil
}
