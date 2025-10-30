package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/google/uuid"
	"github.com/halilbulentorhon/cb-pubsub/config"
	"github.com/halilbulentorhon/cb-pubsub/constant"
	"github.com/halilbulentorhon/cb-pubsub/model"
	util "github.com/halilbulentorhon/cb-pubsub/pkg"
	"github.com/halilbulentorhon/cb-pubsub/repository"
)

type cbPubSub[T any] struct {
	cfg                  config.PubSubConfig
	repository           repository.Repository
	shutdownMgr          *shutdownManager
	logger               util.Logger
	subscribeRetryConfig util.RetryConfig
	cleanupRetryConfig   util.RetryConfig
	subscribeOnce        sync.Once
	channel              string
	instanceId           string
	selfDocId            string
	isSubscribed         bool
}

func (c *cbPubSub[T]) Publish(ctx context.Context, msg T) error {
	var allDoc model.AssignmentDoc
	_, err := c.repository.Get(ctx, constant.AssignmentDocName, &allDoc)
	if err != nil {
		return err
	}

	channel, found := allDoc[c.channel]
	if !found {
		return fmt.Errorf("publish error, channel not found")
	}

	for member := range channel {
		if member == c.instanceId {
			continue
		}
		key := fmt.Sprintf("%s%s", constant.SelfDocPrefix, member)
		err = c.repository.ArrayAppend(ctx, key, constant.MessagesPath, msg)
		if errors.Is(err, gocb.ErrDocumentNotFound) {
			continue
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (c *cbPubSub[T]) Subscribe(ctx context.Context, handler PubSubHandler[T]) error {
	if c.isSubscribed {
		return errors.New("subscribe already called")
	}

	var subscribeErr error
	c.subscribeOnce.Do(func() {
		c.isSubscribed = true
		subscribeErr = c.doSubscribe(ctx, handler)
	})

	return subscribeErr
}

func (c *cbPubSub[T]) doSubscribe(ctx context.Context, handler PubSubHandler[T]) error {
	ticker := time.NewTicker(time.Duration(c.cfg.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.shutdownMgr.Context().Done():
			return errors.New("graceful shutdown")
		case sig := <-c.shutdownMgr.SignalChannel():
			c.logger.Info("graceful shutdown initiated", "signal", sig.String())
			_ = c.Close()
			return errors.New("graceful shutdown")
		case <-ticker.C:
			var selfDoc model.PubSubDoc[T]
			selfDocTTL := time.Duration(constant.SelfDocTtlSeconds) * time.Second

			err := util.WithRetry(ctx, c.subscribeRetryConfig, func() error {
				_, err := c.repository.GetAndTouch(ctx, c.selfDocId, &selfDoc, selfDocTTL)
				if errors.Is(err, gocb.ErrDocumentNotFound) {
					c.logger.Info("self document not found, recreating...", "instance_id", c.instanceId, "channel", c.channel)
					return c.assign(ctx)
				}
				return err
			})

			if err != nil {
				c.logger.Error("failed to get self document after retries, shutting down", "error", err, "instance_id", c.instanceId, "channel", c.channel)
				return fmt.Errorf("subscribe failed after retries: %w", err)
			}

			messageCount := len(selfDoc.Messages)
			if messageCount == 0 {
				continue
			}

			err = handler(selfDoc.Messages)
			if err != nil {
				c.logger.Error("pubsub handler error", "error", err, "message_count", messageCount, "instance_id", c.instanceId)
				continue
			}

			err = util.WithRetry(ctx, c.subscribeRetryConfig, func() error {
				return c.repository.ArrayRemoveFromIndex(ctx, c.selfDocId, constant.MessagesPath, 0, messageCount-1)
			})
			if err != nil {
				c.logger.Error("failed to remove processed messages after retries", "error", err, "message_count", messageCount, "instance_id", c.instanceId)
				continue
			}
		}
	}
}

func (c *cbPubSub[T]) Close() error {
	var repoErr error

	err := c.shutdownMgr.Shutdown(func(shutdownCtx context.Context) {
		if c.repository != nil {
			_ = c.repository.Delete(shutdownCtx, c.selfDocId)
			pathToRemove := util.GetAssignmentPath(c.channel, c.instanceId)
			_ = c.repository.RemoveMultiplePaths(shutdownCtx, constant.AssignmentDocName, []string{pathToRemove})
			repoErr = c.repository.Close()
		}
	})

	if err != nil {
		return err
	}
	return repoErr
}

func (c *cbPubSub[T]) cleanOldMembers() error {
	cleanupInterval := time.Duration(c.cfg.CleanupIntervalSeconds) * time.Second
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	consecutiveFailures := 0

	for {
		select {
		case <-c.shutdownMgr.Context().Done():
			return c.shutdownMgr.Context().Err()
		case <-ticker.C:
			err := util.WithRetry(c.shutdownMgr.Context(), c.cleanupRetryConfig, func() error {
				return c.performCleanup(c.shutdownMgr.Context())
			})

			if err != nil {
				consecutiveFailures++
				c.logger.Error("cleanup failed after retries", "error", err, "consecutive_failures", consecutiveFailures, "instance_id", c.instanceId)

				if consecutiveFailures >= constant.MaxConsecutiveFailures {
					c.logger.Error("cleanup failed too many times, triggering shutdown", "consecutive_failures", consecutiveFailures, "instance_id", c.instanceId)
					return fmt.Errorf("cleanup failed %d consecutive times: %w", consecutiveFailures, err)
				}

				newInterval := time.Duration(float64(cleanupInterval) * constant.CleanupIntervalMultiplier)
				if newInterval > constant.MaxCleanupInterval {
					newInterval = constant.MaxCleanupInterval
				}
				ticker.Reset(newInterval)
				c.logger.Warn("cleanup interval increased due to failures", "new_interval", newInterval, "consecutive_failures", consecutiveFailures)
			} else {
				if consecutiveFailures > 0 {
					c.logger.Info("cleanup recovered after failures", "previous_failures", consecutiveFailures)
					consecutiveFailures = 0
					ticker.Reset(cleanupInterval)
				}
			}
		}
	}
}

func (c *cbPubSub[T]) performCleanup(ctx context.Context) error {
	var allDoc model.AssignmentDoc
	_, err := c.repository.Get(ctx, constant.AssignmentDocName, &allDoc)
	if err != nil {
		return fmt.Errorf("failed to get assignment document: %w", err)
	}

	inactiveMembers := make([]string, 0)
	for channel, memberMap := range allDoc {
		for memberId := range memberMap {
			var res interface{}
			_, err = c.repository.Get(ctx, fmt.Sprintf("%s%s", constant.SelfDocPrefix, memberId), &res)
			if errors.Is(err, gocb.ErrDocumentNotFound) {
				c.logger.Debug("inactive member detected", "member_id", memberId, "channel", channel)
				inactiveMembers = append(inactiveMembers, fmt.Sprintf("%s.%s", channel, memberId))
			}
		}
	}

	if len(inactiveMembers) > 0 {
		err = c.repository.RemoveMultiplePaths(ctx, constant.AssignmentDocName, inactiveMembers)
		if err != nil {
			return fmt.Errorf("failed to remove inactive members: %w", err)
		}

		c.logger.Info("cleaned up inactive members", "count", len(inactiveMembers), "members", inactiveMembers)
	}

	return nil
}

func (c *cbPubSub[T]) assign(ctx context.Context) error {
	selfDocTTL := time.Duration(constant.SelfDocTtlSeconds) * time.Second

	err := c.repository.Upsert(ctx, c.selfDocId, model.CreatePubSubDoc[T](), selfDocTTL)
	if err != nil {
		return err
	}

	currentTimestamp := time.Now().Unix()
	err = c.repository.UpsertPath(ctx, constant.AssignmentDocName, util.GetAssignmentPath(c.channel, c.instanceId), currentTimestamp)
	if err != nil {
		return err
	}

	return nil
}

func NewCbPubSub[T any](channel string, cfg config.PubSubConfig) (PubSub[T], error) {
	cfg.ApplyDefaults()

	id := uuid.NewString()
	logger := util.NewLogger("cb-pubsub").With("instance_id", id, "channel", channel)

	cbPS := &cbPubSub[T]{
		cfg:         cfg,
		channel:     channel,
		instanceId:  id,
		selfDocId:   fmt.Sprintf("%s%s", constant.SelfDocPrefix, id),
		logger:      logger,
		shutdownMgr: newShutdownManager(logger.With("component", "shutdown-manager")),
		subscribeRetryConfig: util.RetryConfig{
			MaxRetries:   cfg.SubscribeRetryAttempts,
			InitialDelay: constant.DefaultSubscribeRetryInitialDelay,
			MaxDelay:     constant.DefaultSubscribeRetryMaxDelay,
			Multiplier:   constant.DefaultSubscribeRetryMultiplier,
		},
		cleanupRetryConfig: util.RetryConfig{
			MaxRetries:   cfg.CleanupRetryAttempts,
			InitialDelay: constant.DefaultCleanupRetryInitialDelay,
			MaxDelay:     constant.DefaultCleanupRetryMaxDelay,
			Multiplier:   constant.DefaultCleanupRetryMultiplier,
		},
	}

	repo, err := repository.NewCouchbaseRepository(cfg.CouchbaseConfig)
	if err != nil {
		return nil, err
	}
	cbPS.repository = repo

	initCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.InitTimeoutSec)*time.Second)
	defer cancel()

	err = cbPS.assign(initCtx)
	if err != nil {
		return nil, err
	}

	go func() {
		err = cbPS.cleanOldMembers()
		if err != nil && !errors.Is(err, context.Canceled) {
			cbPS.logger.Error("cleanOldMembers failed, initiating graceful shutdown", "error", err)
			_ = cbPS.Close()
		}
	}()

	return cbPS, nil
}
