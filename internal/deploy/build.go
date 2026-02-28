package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/stax-ai/stax-cli/internal/build"
)

// BuildAndPush builds the container image and pushes it to the registry.
func BuildAndPush(ctx context.Context, projectRoot, image, builder string, skipBuild, skipPush bool) error {
	// Build the image
	if !skipBuild {
		excludes, err := build.ReadStaxignore(projectRoot)
		if err != nil {
			return fmt.Errorf("read .staxignore: %w", err)
		}

		opts := build.Options{
			ProjectRoot:  projectRoot,
			Image:         image,
			Builder:       builder,
			Exclude:       excludes,
			TrustBuilder:  true, // Trust gcr.io/buildpacks/builder
			Verbose:       true, // Stream logs for deploy
		}

		if err := build.Build(ctx, opts, nil); err != nil {
			return fmt.Errorf("build image: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Built: %s\n", image)
	}

	// Push the image
	if !skipPush {
		if err := pushImage(ctx, image); err != nil {
			return fmt.Errorf("push image: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Pushed: %s\n", image)
	}

	return nil
}

// pushImage pushes the image to the registry using docker CLI.
func pushImage(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "push", image)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	return nil
}
