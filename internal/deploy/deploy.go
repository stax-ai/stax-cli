package deploy

import (
	"context"
	"fmt"
	"os"
)

// Options configures a deployment.
type Options struct {
	ProjectRoot string
	Config      *Config
	ConfigPath  string
	Builder     string
	SkipBuild   bool
	SkipPush    bool
	DryRun      bool
}

// Deploy orchestrates the full deployment workflow:
// 1. Build image (if not skipped)
// 2. Push image (if not skipped)
// 3. Generate DopplerSecret
// 4. Generate Knative Service
// 5. Generate Pipe
// 6. Memoize resources to deploy/ folder
// 7. Apply resources (if not dry-run)
func Deploy(ctx context.Context, opts Options) error {
	// Get full image name
	image, err := opts.Config.GetFullImageName()
	if err != nil {
		return fmt.Errorf("get image name: %w", err)
	}

	// Build and push image
	if !opts.SkipBuild || !opts.SkipPush {
		if err := BuildAndPush(ctx, opts.ProjectRoot, image, opts.Builder, opts.SkipBuild, opts.SkipPush); err != nil {
			return fmt.Errorf("build and push: %w", err)
		}
	}

	// Get namespace
	namespace := opts.Config.GetNamespace()

	// Generate DopplerSecret
	dopplerSecret, err := GenerateDopplerSecret(opts.Config, namespace)
	if err != nil {
		return fmt.Errorf("generate doppler secret: %w", err)
	}

	// Generate Knative Service
	knativeService, err := GenerateKnativeService(opts.Config, image)
	if err != nil {
		return fmt.Errorf("generate knative service: %w", err)
	}

	// Generate Pipe (replaces KameletBinding)
	pipe, err := GeneratePipe(opts.Config, namespace)
	if err != nil {
		return fmt.Errorf("generate pipe: %w", err)
	}

	// Memoize resources to deploy/ folder
	if err := MemoizeResources(opts.ProjectRoot, knativeService, pipe, nil, dopplerSecret); err != nil {
		return fmt.Errorf("memoize resources: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Generated manifests saved to deploy/ folder\n")

	// Ensure Pub/Sub topic and subscription exist
	if !opts.DryRun {
		// Ensure topic exists
		if err := EnsurePubSubTopic(ctx, opts.Config.PubSubTopic); err != nil {
			return fmt.Errorf("ensure pubsub topic: %w", err)
		}

		// Ensure subscription exists (subscription name is service-name-subscription)
		subscriptionName := fmt.Sprintf("%s-subscription", opts.Config.ServiceName)
		if err := EnsurePubSubSubscription(ctx, subscriptionName, opts.Config.PubSubTopic); err != nil {
			return fmt.Errorf("ensure pubsub subscription: %w", err)
		}
	}

	// Apply resources if not dry-run
	if !opts.DryRun {
		// Check kubectl availability
		if err := CheckKubectl(); err != nil {
			return fmt.Errorf("kubectl check: %w", err)
		}

		// Ensure service-specific doppler-token-secret exists before applying resources
		if err := EnsureDopplerTokenSecret(ctx, opts.Config.ServiceName); err != nil {
			return fmt.Errorf("ensure doppler token secret: %w", err)
		}

		if err := ApplyResources(ctx, opts.ProjectRoot, namespace, false); err != nil {
			return fmt.Errorf("apply resources: %w", err)
		}
	}

	return nil
}
