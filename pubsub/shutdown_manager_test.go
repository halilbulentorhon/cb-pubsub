package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"

	util "github.com/halilbulentorhon/cb-pubsub/pkg"
)

func TestNewShutdownManager(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	if sm == nil {
		t.Fatal("shutdown manager should not be nil")
	}

	if sm.Context() == nil {
		t.Fatal("context should not be nil")
	}

	if sm.SignalChannel() == nil {
		t.Fatal("signal channel should not be nil")
	}

	if sm.IsClosed() {
		t.Fatal("should not be closed initially")
	}
}

func TestShutdownManager_IsClosed(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	if sm.IsClosed() {
		t.Fatal("should not be closed initially")
	}

	err := sm.Shutdown(nil)
	if err != nil {
		t.Fatalf("shutdown should not return error: %v", err)
	}

	if !sm.IsClosed() {
		t.Fatal("should be closed after shutdown")
	}
}

func TestShutdownManager_Context(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	ctx := sm.Context()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}

	select {
	case <-ctx.Done():
		t.Fatal("context should not be done initially")
	default:
	}

	err := sm.Shutdown(nil)
	if err != nil {
		t.Fatalf("shutdown should not return error: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
		t.Fatal("context should be done after shutdown")
	}
}

func TestShutdownManager_ShutdownOnce(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	var cleanupCallCount int
	var mu sync.Mutex

	cleanupFunc := func(ctx context.Context) {
		mu.Lock()
		cleanupCallCount++
		mu.Unlock()
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sm.Shutdown(cleanupFunc)
			if err != nil {
				t.Errorf("shutdown should not return error: %v", err)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	if cleanupCallCount != 1 {
		t.Fatalf("cleanup should be called exactly once, got %d", cleanupCallCount)
	}
	mu.Unlock()

	if !sm.IsClosed() {
		t.Fatal("should be closed after shutdown")
	}
}

func TestShutdownManager_CleanupWithTimeout(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	var cleanupCalled bool
	var cleanupContext context.Context

	cleanupFunc := func(ctx context.Context) {
		cleanupCalled = true
		cleanupContext = ctx

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("cleanup context should have deadline")
		}

		if time.Until(deadline) > 11*time.Second || time.Until(deadline) < 9*time.Second {
			t.Error("cleanup context should have ~10 second timeout")
		}
	}

	err := sm.Shutdown(cleanupFunc)
	if err != nil {
		t.Fatalf("shutdown should not return error: %v", err)
	}

	if !cleanupCalled {
		t.Fatal("cleanup function should be called")
	}

	if cleanupContext == nil {
		t.Fatal("cleanup context should not be nil")
	}
}

func TestShutdownManager_SignalChannel(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	sigCh := sm.SignalChannel()
	if sigCh == nil {
		t.Fatal("signal channel should not be nil")
	}

	select {
	case <-sigCh:
		t.Fatal("should not receive signal initially")
	default:
	}
}

func TestShutdownManager_NoCleanupFunction(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	err := sm.Shutdown(nil)
	if err != nil {
		t.Fatalf("shutdown should not return error: %v", err)
	}

	if !sm.IsClosed() {
		t.Fatal("should be closed after shutdown")
	}
}

func TestShutdownManager_CleanupPanic(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	cleanupFunc := func(ctx context.Context) {
		panic("cleanup panic")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("shutdown should not panic even if cleanup panics: %v", r)
		}
	}()

	err := sm.Shutdown(cleanupFunc)
	if err != nil {
		t.Fatalf("shutdown should not return error: %v", err)
	}

	if !sm.IsClosed() {
		t.Fatal("should be closed after shutdown")
	}
}

func TestShutdownManager_IntegrationTest(t *testing.T) {
	logger := util.NewLogger("test")
	sm := newShutdownManager(logger)

	var cleanupCalled bool
	var mu sync.Mutex
	done := make(chan bool, 1)

	go func() {
		select {
		case <-sm.Context().Done():
			done <- true
		case <-time.After(5 * time.Second):
			done <- false
		}
	}()

	cleanupFunc := func(ctx context.Context) {
		mu.Lock()
		cleanupCalled = true
		mu.Unlock()
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := sm.Shutdown(cleanupFunc)
		if err != nil {
			t.Errorf("shutdown should not return error: %v", err)
		}
	}()

	success := <-done
	if !success {
		t.Fatal("shutdown should complete within timeout")
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	called := cleanupCalled
	mu.Unlock()

	if !called {
		t.Fatal("cleanup function should be called")
	}

	if !sm.IsClosed() {
		t.Fatal("should be closed after shutdown")
	}
}
