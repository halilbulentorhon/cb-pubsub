package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/halilbulentorhon/cb-pubsub/config"
	"github.com/halilbulentorhon/cb-pubsub/pubsub"
)

func main() {
	// Configure Couchbase connection
	cfg := config.PubSubConfig{
		CouchbaseConfig: config.CouchbaseConfig{
			Host:       "localhost:8091",
			Username:   "Administrator",
			Password:   "password",
			BucketName: "PubSub",
		},
	}

	// Create PubSub instance for string messages
	ps, err := pubsub.NewCbPubSub[string]("my-channel", cfg)
	if err != nil {
		log.Fatalf("Failed to create PubSub: %v", err)
	}
	defer ps.Close()

	// Start subscriber in background
	go func() {
		err = ps.Subscribe(context.Background(), func(messages []string) error {
			for _, msg := range messages {
				fmt.Printf("Received: %s\n", msg)
			}
			return nil
		})
		if err != nil {
			log.Printf("Subscribe error: %v", err)
		}
	}()

	// Give subscriber time to start
	time.Sleep(2 * time.Second)

	// Publish some messages
	messages := []string{
		"Hello, World!",
		"How are you?",
		"This is a test message",
	}

	for i, msg := range messages {
		err = ps.Publish(context.Background(), msg)
		if err != nil {
			log.Printf("Publish error: %v", err)
		} else {
			fmt.Printf("Published: %s\n", msg)
		}

		// Wait between messages
		if i < len(messages)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// Let messages be processed
	time.Sleep(3 * time.Second)
	fmt.Println("Example completed!")
	ps.Close()
	time.Sleep(3 * time.Second)
}
