package deploy

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2/google"
)

// EnsurePubSubTopic creates a Pub/Sub topic if it doesn't exist.
// Uses the project from Application Default Credentials.
func EnsurePubSubTopic(ctx context.Context, topicName string) error {
	// Get project from Application Default Credentials
	projectID, err := getGCPProject(ctx)
	if err != nil {
		return fmt.Errorf("get GCP project: %w", err)
	}

	// Create Pub/Sub client
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("create pubsub client: %w", err)
	}
	defer client.Close()

	// Check if topic exists
	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check topic existence: %w", err)
	}

	if !exists {
		// Create the topic
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return fmt.Errorf("create topic: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Created Pub/Sub topic: %s in project %s\n", topicName, projectID)
	} else {
		fmt.Fprintf(os.Stderr, "Pub/Sub topic already exists: %s in project %s\n", topicName, projectID)
	}

	return nil
}

// EnsurePubSubSubscription creates a Pub/Sub subscription if it doesn't exist.
// Uses the project from Application Default Credentials.
func EnsurePubSubSubscription(ctx context.Context, subscriptionName, topicName string) error {
	// Get project from Application Default Credentials
	projectID, err := getGCPProject(ctx)
	if err != nil {
		return fmt.Errorf("get GCP project: %w", err)
	}

	// Create Pub/Sub client
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("create pubsub client: %w", err)
	}
	defer client.Close()

	// Check if subscription exists
	sub := client.Subscription(subscriptionName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check subscription existence: %w", err)
	}

	if !exists {
		// Get or create the topic first
		topic := client.Topic(topicName)
		topicExists, err := topic.Exists(ctx)
		if err != nil {
			return fmt.Errorf("check topic existence: %w", err)
		}
		if !topicExists {
			// Create the topic if it doesn't exist
			topic, err = client.CreateTopic(ctx, topicName)
			if err != nil {
				return fmt.Errorf("create topic: %w", err)
			}
		}

		// Create the subscription
		sub, err = client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			return fmt.Errorf("create subscription: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Created Pub/Sub subscription: %s for topic %s in project %s\n", subscriptionName, topicName, projectID)
	} else {
		fmt.Fprintf(os.Stderr, "Pub/Sub subscription already exists: %s in project %s\n", subscriptionName, projectID)
	}

	return nil
}

// getGCPProject gets the GCP project ID from Application Default Credentials.
func getGCPProject(ctx context.Context) (string, error) {
	// Use the metadata service or credentials to get the project
	// The pubsub client will use ADC, but we need the project ID separately
	// We can use the compute metadata service or credentials
	
	// Try to get from environment variable first (common in CI/CD)
	if projectID := os.Getenv("GCP_PROJECT"); projectID != "" {
		return projectID, nil
	}
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		return projectID, nil
	}

	// Use the credentials to get project ID
	creds, err := google.FindDefaultCredentials(ctx, pubsub.ScopePubSub)
	if err != nil {
		return "", fmt.Errorf("get default credentials: %w", err)
	}

	if creds.ProjectID != "" {
		return creds.ProjectID, nil
	}

	return "", fmt.Errorf("could not determine GCP project ID from Application Default Credentials. Set GCP_PROJECT or GOOGLE_CLOUD_PROJECT environment variable")
}
