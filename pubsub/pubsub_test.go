package pubsub

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/halilbulentorhon/cb-pubsub/config"
	"github.com/halilbulentorhon/cb-pubsub/constant"
	"github.com/halilbulentorhon/cb-pubsub/mocks"
	"github.com/halilbulentorhon/cb-pubsub/model"
	util "github.com/halilbulentorhon/cb-pubsub/pkg"
	"go.uber.org/mock/gomock"
)

func createTestCbPubSub(t *testing.T, mockRepo *mocks.MockRepository) *cbPubSub[string] {
	cfg := config.PubSubConfig{
		PollIntervalSeconds:    1,
		CleanupIntervalSeconds: 15,
		SubscribeRetryAttempts: 3,
		CleanupRetryAttempts:   5,
	}

	logger := util.NewDevLogger("test")

	return &cbPubSub[string]{
		cfg:         cfg,
		repository:  mockRepo,
		channel:     "test-channel",
		instanceId:  "test-instance",
		selfDocId:   constant.SelfDocPrefix + "test-instance",
		shutdownMgr: newShutdownManager(logger.With("component", "shutdown-manager")),
		logger:      logger,
		subscribeRetryConfig: util.RetryConfig{
			MaxRetries:   3,
			InitialDelay: time.Millisecond,
			MaxDelay:     10 * time.Millisecond,
			Multiplier:   2.0,
		},
		cleanupRetryConfig: util.RetryConfig{
			MaxRetries:   5,
			InitialDelay: time.Millisecond,
			MaxDelay:     10 * time.Millisecond,
			Multiplier:   2.0,
		},
	}
}

func TestCbPubSub_Publish_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	assignmentDoc := model.AssignmentDoc{
		"test-channel": {
			"instance1": 1234567890,
			"instance2": 1234567891,
		},
	}

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.AssignmentDocName, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, result interface{}) (gocb.Cas, error) {
			*(result.(*model.AssignmentDoc)) = assignmentDoc
			return gocb.Cas(123), nil
		})

	mockRepo.EXPECT().
		ArrayAppend(gomock.Any(), constant.SelfDocPrefix+"instance1", constant.MessagesPath, "test-message").
		Return(nil)

	mockRepo.EXPECT().
		ArrayAppend(gomock.Any(), constant.SelfDocPrefix+"instance2", constant.MessagesPath, "test-message").
		Return(nil)

	err := pubsub.Publish(context.Background(), "test-message")
	if err != nil {
		t.Errorf("Publish returned error: %v", err)
	}
}

func TestCbPubSub_Publish_channelNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	assignmentDoc := model.AssignmentDoc{
		"other-channel": {
			"instance1": 1234567890,
		},
	}

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.AssignmentDocName, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, result interface{}) (gocb.Cas, error) {
			*(result.(*model.AssignmentDoc)) = assignmentDoc
			return gocb.Cas(123), nil
		})

	err := pubsub.Publish(context.Background(), "test-message")
	if err == nil {
		t.Error("Publish should return error when channel not found")
	}
	if err.Error() != "publish error, channel not found" {
		t.Errorf("Publish error = %v, want 'publish error, channel not found'", err)
	}
}

func TestCbPubSub_Publish_DocumentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	assignmentDoc := model.AssignmentDoc{
		"test-channel": {
			"instance1": 1234567890,
			"instance2": 1234567891,
		},
	}

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.AssignmentDocName, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, result interface{}) (gocb.Cas, error) {
			*(result.(*model.AssignmentDoc)) = assignmentDoc
			return gocb.Cas(123), nil
		})

	mockRepo.EXPECT().
		ArrayAppend(gomock.Any(), constant.SelfDocPrefix+"instance1", constant.MessagesPath, "test-message").
		Return(gocb.ErrDocumentNotFound)

	mockRepo.EXPECT().
		ArrayAppend(gomock.Any(), constant.SelfDocPrefix+"instance2", constant.MessagesPath, "test-message").
		Return(nil)

	err := pubsub.Publish(context.Background(), "test-message")
	if err != nil {
		t.Errorf("Publish returned error: %v", err)
	}
}

func TestCbPubSub_PerformCleanup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	assignmentDoc := model.AssignmentDoc{
		"test-channel": {
			"active-instance":   1234567890,
			"inactive-instance": 1234567891,
		},
		"other-channel": {
			"another-inactive": 1234567892,
		},
	}

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.AssignmentDocName, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, result interface{}) (gocb.Cas, error) {
			*(result.(*model.AssignmentDoc)) = assignmentDoc
			return gocb.Cas(123), nil
		})

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.SelfDocPrefix+"active-instance", gomock.Any()).
		Return(gocb.Cas(456), nil)

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.SelfDocPrefix+"inactive-instance", gomock.Any()).
		Return(gocb.Cas(0), gocb.ErrDocumentNotFound)

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.SelfDocPrefix+"another-inactive", gomock.Any()).
		Return(gocb.Cas(0), gocb.ErrDocumentNotFound)

	mockRepo.EXPECT().
		RemoveMultiplePaths(gomock.Any(), constant.AssignmentDocName, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, paths []string) error {
			if len(paths) != 2 {
				t.Errorf("Expected 2 paths, got %d", len(paths))
			}

			expectedPaths := map[string]bool{
				"test-channel.inactive-instance": true,
				"other-channel.another-inactive": true,
			}

			for _, path := range paths {
				if !expectedPaths[path] {
					t.Errorf("Unexpected path: %s", path)
				}
			}

			return nil
		})

	err := pubsub.performCleanup(context.Background())
	if err != nil {
		t.Errorf("performCleanup returned error: %v", err)
	}
}

func TestCbPubSub_PerformCleanup_NoInactiveMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	assignmentDoc := model.AssignmentDoc{
		"test-channel": {
			"instance1": 1234567890,
			"instance2": 1234567891,
		},
	}

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.AssignmentDocName, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, result interface{}) (gocb.Cas, error) {
			*(result.(*model.AssignmentDoc)) = assignmentDoc
			return gocb.Cas(123), nil
		})

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.SelfDocPrefix+"instance1", gomock.Any()).
		Return(gocb.Cas(456), nil)

	mockRepo.EXPECT().
		Get(gomock.Any(), constant.SelfDocPrefix+"instance2", gomock.Any()).
		Return(gocb.Cas(789), nil)

	err := pubsub.performCleanup(context.Background())
	if err != nil {
		t.Errorf("performCleanup returned error: %v", err)
	}
}

func TestCbPubSub_Assign_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	expectedTTL := time.Duration(constant.SelfDocTtlSeconds) * time.Second
	mockRepo.EXPECT().
		Upsert(gomock.Any(), pubsub.selfDocId, gomock.Any(), expectedTTL).
		Return(nil)

	expectedPath := util.GetAssignmentPath(pubsub.channel, pubsub.instanceId)
	mockRepo.EXPECT().
		UpsertPath(gomock.Any(), constant.AssignmentDocName, expectedPath, gomock.Any()).
		Return(nil)

	err := pubsub.assign(context.Background())
	if err != nil {
		t.Errorf("assign returned error: %v", err)
	}
}

func TestCbPubSub_Close_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	mockRepo.EXPECT().Delete(gomock.Any(), pubsub.selfDocId).Return(nil)
	mockRepo.EXPECT().RemoveMultiplePaths(gomock.Any(), constant.AssignmentDocName, gomock.Any()).Return(nil)
	mockRepo.EXPECT().Close().Return(nil)

	err := pubsub.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

func TestCbPubSub_Close_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pubsub := createTestCbPubSub(t, mockRepo)

	expectedErr := errors.New("close error")
	mockRepo.EXPECT().Delete(gomock.Any(), pubsub.selfDocId).Return(nil)
	mockRepo.EXPECT().RemoveMultiplePaths(gomock.Any(), constant.AssignmentDocName, gomock.Any()).Return(nil)
	mockRepo.EXPECT().Close().Return(expectedErr)

	err := pubsub.Close()
	if err != expectedErr {
		t.Errorf("Close returned error: %v, want %v", err, expectedErr)
	}
}

func TestCbPubSub_Close_NilRepository(t *testing.T) {
	logger := util.NewDevLogger("test")
	pubsub := &cbPubSub[string]{
		shutdownMgr: newShutdownManager(logger.With("component", "shutdown-manager")),
	}

	err := pubsub.Close()
	if err != nil {
		t.Errorf("Close with nil repository returned error: %v", err)
	}
}

func TestNewCbPubSub_ConfigDefaults(t *testing.T) {
	cfg := config.PubSubConfig{
		CouchbaseConfig: config.CouchbaseConfig{
			Host:           "couchbase://localhost",
			Username:       "admin",
			Password:       "password",
			BucketName:     "test",
			ScopeName:      "scope",
			CollectionName: "collection",
		},
	}

	_, err := NewCbPubSub[string]("test-channel", cfg)
	if err == nil {
		t.Error("NewCbPubSub should fail with invalid connection")
	}
}

func TestPubSubHandler_Function(t *testing.T) {
	var handlerCalled bool
	var receivedMessages []string

	handler := func(messages []string) error {
		handlerCalled = true
		receivedMessages = messages
		return nil
	}

	testMessages := []string{"msg1", "msg2", "msg3"}
	err := handler(testMessages)

	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}
	if len(receivedMessages) != 3 {
		t.Errorf("Handler received %d messages, want 3", len(receivedMessages))
	}
	for i, msg := range testMessages {
		if receivedMessages[i] != msg {
			t.Errorf("Message %d = %s, want %s", i, receivedMessages[i], msg)
		}
	}
}

func TestPubSubHandler_Error(t *testing.T) {
	expectedErr := errors.New("handler error")
	handler := func(messages []string) error {
		return expectedErr
	}

	err := handler([]string{"test"})
	if err != expectedErr {
		t.Errorf("Handler returned error: %v, want %v", err, expectedErr)
	}
}
