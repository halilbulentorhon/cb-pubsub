package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/halilbulentorhon/cb-pubsub/config"
	"github.com/halilbulentorhon/cb-pubsub/pubsub"
)

func main() {
	log.Println("🚀 Starting cb-pubsub test application...")

	cfg := config.PubSubConfig{
		CouchbaseConfig: config.CouchbaseConfig{
			Host:                "localhost:8091",
			Username:            "Administrator",
			Password:            "password",
			BucketName:          "PubSub",
			ScopeName:           "_default",
			CollectionName:      "_default",
			ConnectTimeoutSec:   10,
			OperationTimeoutSec: 5,
		},
		PollIntervalSeconds:    1,
		SelfDocTTLSeconds:      30,
		CleanupIntervalSeconds: 15,
	}

	ps, err := pubsub.NewCbPubSub[string]("test-group", cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create PubSub instance: %v", err)
	}
	defer func() {
		log.Println("🧹 Cleaning up resources...")
		ps.Close()
	}()

	messageHandler := func(messages []string) error {
		for _, message := range messages {
			log.Printf("📨 Received message: %s", message)
		}
		return nil
	}

	subscribeErrCh := make(chan error, 1)
	go func() {
		log.Println("👂 Starting subscription...")
		err := ps.Subscribe(context.Background(), messageHandler)
		subscribeErrCh <- err
	}()

	startTime := time.Now()
	closeTime := startTime.Add(5 * time.Second)
	endTime := startTime.Add(30 * time.Second)

	publishTicker := time.NewTicker(500 * time.Millisecond)
	defer publishTicker.Stop()

	closeTimer := time.NewTimer(5 * time.Second)
	defer closeTimer.Stop()

	endTimer := time.NewTimer(30 * time.Second)
	defer endTimer.Stop()

	messageCounter := 1
	isClosed := false

	log.Println("📝 Starting message publishing (every 500ms)...")
	log.Printf("⏱️  Will call Close() at: %v (5s)", closeTime.Format("15:04:05"))
	log.Printf("🏁 Will exit at: %v (30s)", endTime.Format("15:04:05"))

	for {
		select {
		case err := <-subscribeErrCh:
			if err != nil && strings.Contains(err.Error(), "graceful shutdown") {
				log.Println("✅ Graceful shutdown completed")
			} else if err != nil {
				log.Printf("❌ Subscribe error: %v", err)
			} else {
				log.Println("🔄 Subscribe ended normally")
			}
			return

		case <-closeTimer.C:
			if !isClosed {
				elapsed := time.Since(startTime)
				log.Printf("🔒 Calling Close() after %v (at 5s mark)...", elapsed.Truncate(time.Millisecond))
				err := ps.Close()
				if err != nil {
					log.Printf("⚠️  Close() returned error: %v", err)
				} else {
					log.Println("✅ Close() completed successfully")
				}
				isClosed = true

			}

		case <-endTimer.C:
			elapsed := time.Since(startTime)
			log.Printf("⏰ Time's up! Exiting after %v (30s total)", elapsed.Truncate(time.Millisecond))
			log.Printf("📊 Total messages published: %d", messageCounter-1)
			return

		case <-publishTicker.C:
			if !isClosed {
				elapsed := time.Since(startTime)
				message := fmt.Sprintf("msg_%d_t%s", messageCounter, elapsed.Truncate(time.Millisecond))

				err := ps.Publish(context.Background(), message)
				if err != nil {
					log.Printf("❌ Publish error for message %d: %v", messageCounter, err)
				} else {
					log.Printf("✉️  Published #%d: %s (elapsed: %v)", messageCounter, message, elapsed.Truncate(time.Millisecond))
				}
				messageCounter++
			}
		}
	}
}
