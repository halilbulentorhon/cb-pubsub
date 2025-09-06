# Couchbase PubSub

A modern, type-safe publish-subscribe library for Couchbase built in Go with support for generics.

## Features

- 🔧 **Type-safe**: Full support for Go generics
- 🔄 **Auto-cleanup**: Automatic cleanup of inactive subscribers
- 🎯 **Context Support**: Proper context cancellation support
- 🔄 **Retry Logic**: Built-in retry mechanisms with exponential backoff

## Installation

```bash
go get github.com/halilbulentorhon/cb-pubsub@latest
```

## Quick Start

### Basic Configuration

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/halilbulentorhon/cb-pubsub/config"
    "github.com/halilbulentorhon/cb-pubsub/pubsub"
)

func main() {
    cfg := config.PubSubConfig{
        CouchbaseConfig: config.CouchbaseConfig{
            Host:                "localhost:8091",
            Username:            "admin",
            Password:            "password",
            BucketName:          "default",
            ScopeName:           "_default",
            CollectionName:      "_default",
            ConnectTimeoutSec:   10,
            OperationTimeoutSec: 5,
        },
        PollIntervalSeconds:    1,  // Optional, defaults to 1 second
        CleanupIntervalSeconds: 15, // Optional, defaults to 15 seconds
        SubscribeRetryAttempts: 3,  // Optional, defaults to 3
        CleanupRetryAttempts:   5,  // Optional, defaults to 5
        ShutdownTimeoutSec:     10, // Optional, defaults to 10 seconds
        InitTimeoutSec:         30, // Optional, defaults to 30 seconds
    }

    // Create a PubSub instance for string messages
    ps, err := pubsub.NewCbPubSub[string]("my-channel", cfg)
    if err != nil {
        panic(err)
    }
    defer ps.Close()

    // Subscribe to messages
    go func() {
        err := ps.Subscribe(context.Background(), func(messages []string) error {
            for _, msg := range messages {
                fmt.Println("Received:", msg)
            }
            return nil
        })
        if err != nil {
            fmt.Println("Subscribe error:", err)
        }
    }()

    // Publish a message (note: channel is set during NewCbPubSub)
    err = ps.Publish(context.Background(), "Hello, World!")
    if err != nil {
        fmt.Println("Publish error:", err)
    }
}
```

### Working with Custom Types

```go
type MyMessage struct {
    ID      string    `json:"id"`
    Content string    `json:"content"`
    Time    time.Time `json:"time"`
}

// Create PubSub for custom type
ps, err := pubsub.NewCbPubSub[MyMessage]("custom-channel", cfg)
if err != nil {
    panic(err)
}

// Subscribe with custom handler
err = ps.Subscribe(context.Background(), func(messages []MyMessage) error {
    for _, msg := range messages {
        fmt.Printf("Received message %s: %s\n", msg.ID, msg.Content)
    }
    return nil
})

// Publish custom message
msg := MyMessage{
    ID:      "msg-123",
    Content: "Hello from custom type!",
    Time:    time.Now(),
}
err = ps.Publish(context.Background(), msg)
```

## API Reference

### PubSub Interface

```go
type PubSub[T any] interface {
    Publish(ctx context.Context, msg T) error
    Subscribe(ctx context.Context, handler PubSubHandler[T]) error
    Close() error
}

type PubSubHandler[T any] func(messages []T) error
```

### Configuration

```go
type PubSubConfig struct {
    CouchbaseConfig        CouchbaseConfig `json:"couchbaseConfig"`
    PollIntervalSeconds    int             `json:"pollIntervalSeconds"`    // Defaults to 1
    CleanupIntervalSeconds int             `json:"cleanupIntervalSeconds"` // Defaults to 15
    SubscribeRetryAttempts int             `json:"subscribeRetryAttempts"` // Defaults to 3
    CleanupRetryAttempts   int             `json:"cleanupRetryAttempts"`   // Defaults to 5
    ShutdownTimeoutSec     int             `json:"shutdownTimeoutSec"`     // Defaults to 10
    InitTimeoutSec         int             `json:"initTimeoutSec"`         // Defaults to 30
}

type CouchbaseConfig struct {
    Host                string `json:"host"`
    Username            string `json:"username"`
    Password            string `json:"password"`
    BucketName          string `json:"bucket"`
    ScopeName           string `json:"scope"`
    CollectionName      string `json:"collection"`
    ConnectTimeoutSec   int    `json:"connectTimeoutSec"`   // Defaults to 10
    OperationTimeoutSec int    `json:"operationTimeoutSec"` // Defaults to 5
}
```

## Architecture

### How It Works

1. **Instance Registration**: Each PubSub instance registers itself in Couchbase with a unique UUID
2. **Assignment Document**: A global assignment document (`_pubsub_all`) tracks all active instances per channel
3. **Instance Documents**: Each instance has its own document (`_pubsub_instance_{uuid}`) containing messages
4. **Message Delivery**: Publishers append messages to all active instances in the target channel
5. **Polling**: Subscribers poll their instance documents for new messages at configurable intervals
6. **Auto-cleanup**: A background cleanup process removes inactive instances from assignment document
7. **TTL Management**: Instance documents are automatically expired (default: 600 seconds) if not refreshed

### Document Structure

**Assignment Document** (`_pubsub_all`):
```json
{
  "channel1": {
    "uuid-1": 1693123456,
    "uuid-2": 1693123457
  },
  "channel2": {
    "uuid-3": 1693123458
  }
}
```

**Instance Document** (`_pubsub_instance_{uuid}`):
```json
{
  "messages": ["msg1", "msg2", "msg3"],
  "creationDate": 1693123456
}
```

## Development

### Building and Testing

```bash
# Install dependencies
go mod download

# Generate mocks
make mocks

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Clean test cache
go clean -testcache

# Clean generated mocks
make clean-mocks
```

### Available Make Commands

```bash
make mocks       # Generate mocks for testing
make clean-mocks # Remove all generated mock files
```

## Examples

See the [examples/](examples/) directory for more usage examples.

## Requirements

- Go 1.21+
- Couchbase Server 6.0+
- Couchbase Go SDK v2

## Dependencies

- `github.com/couchbase/gocb/v2` - Couchbase Go SDK
- `github.com/google/uuid` - UUID generation
- `go.uber.org/mock` - Mock generation for testing

## Error Handling

The library includes comprehensive error handling with:
- Automatic retry mechanisms with exponential backoff
- Context-aware cancellation
- Graceful handling of network failures
- Structured logging for debugging

## Performance Considerations

- Messages are processed in batches for efficiency
- Configurable polling intervals to balance latency vs resource usage
- TTL-based cleanup prevents document bloat
- Connection pooling through Couchbase SDK
