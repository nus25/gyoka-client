// Command example authenticates to Gyoka through a Bluesky PDS and lists feeds.
//
// Run with:
//
//	GYOKA_HOST=editor.example.com BLUESKY_IDENTIFIER=sample.bsky.social BLUESKY_APP_PASSWORD=xxxx-xxxx-xxxx-xxxx go run ./example
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	client "github.com/nus25/gyoka-client/go-atproto"
)

func main() {
	ctx := context.Background()
	gyokaClient, err := client.New(
		ctx,
		requiredEnv("GYOKA_HOST"),
		requiredEnv("BLUESKY_IDENTIFIER"),
		requiredEnv("BLUESKY_APP_PASSWORD"),
	)
	if err != nil {
		log.Fatalf("create Gyoka client: %v", err)
	}

	ping, err := gyokaClient.Ping(ctx)
	if err != nil {
		log.Fatalf("ping Gyoka: %v", err)
	}
	fmt.Printf("Gyoka: %s\n", ping.Message)

	feeds, err := gyokaClient.ListFeeds(ctx)
	if err != nil {
		log.Fatalf("list feeds: %v", err)
	}
	for _, feed := range feeds.Feeds {
		fmt.Printf("%s active=%t langFilter=%t\n", feed.Uri, feed.IsActive, feed.LangFilter)
	}
}

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}
