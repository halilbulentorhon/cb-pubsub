package pubsub

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/halilbulentorhon/cb-pubsub/constant"
	util "github.com/halilbulentorhon/cb-pubsub/pkg"
)

type shutdownManager struct {
	ctx          context.Context
	cancel       context.CancelFunc
	signalCh     chan os.Signal
	shutdownOnce sync.Once
	logger       util.Logger
}

func newShutdownManager(logger util.Logger) *shutdownManager {
	ctx, cancel := context.WithCancel(context.Background())
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	return &shutdownManager{
		ctx:      ctx,
		cancel:   cancel,
		signalCh: signalCh,
		logger:   logger,
	}
}

func (sm *shutdownManager) Context() context.Context {
	return sm.ctx
}

func (sm *shutdownManager) SignalChannel() <-chan os.Signal {
	return sm.signalCh
}

func (sm *shutdownManager) IsClosed() bool {
	select {
	case <-sm.ctx.Done():
		return true
	default:
		return false
	}
}

func (sm *shutdownManager) Shutdown(cleanupFunc func(context.Context)) error {
	var err error
	sm.shutdownOnce.Do(func() {
		sm.logger.Info("graceful shutdown started")

		sm.cancel()

		if sm.signalCh != nil {
			signal.Stop(sm.signalCh)
		}
		if cleanupFunc != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						sm.logger.Error("cleanup function panicked", "panic", r)
					}
				}()

				shutdownCtx, cancel := context.WithTimeout(context.Background(), constant.DefaultShutdownTimeout)
				defer cancel()
				cleanupFunc(shutdownCtx)
			}()
		}

		sm.logger.Info("graceful shutdown completed")
	})
	return err
}
