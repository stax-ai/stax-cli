package doppler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EnsureProjectAndConfig verifies the Doppler project and config exist.
// Missing resources are created and local Doppler setup is updated.
func EnsureProjectAndConfig(ctx context.Context, project, config string) error {
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("doppler project cannot be empty")
	}
	if strings.TrimSpace(config) == "" {
		return fmt.Errorf("doppler config cannot be empty")
	}

	if err := ensureProject(ctx, project); err != nil {
		return err
	}
	if err := ensureConfig(ctx, project, config); err != nil {
		return err
	}
	if err := setupProjectConfig(ctx, project, config); err != nil {
		return err
	}

	return nil
}

func ensureProject(ctx context.Context, project string) error {
	if _, err := runDoppler(ctx, "projects", "get", project); err == nil {
		fmt.Fprintf(os.Stderr, "Doppler project %s exists\n", project)
		return nil
	}

	out, err := runDoppler(ctx, "projects", "create", project)
	if err != nil {
		return fmt.Errorf("create doppler project %s: %w (%s)", project, err, out)
	}

	fmt.Fprintf(os.Stderr, "Created Doppler project %s\n", project)
	return nil
}

func ensureConfig(ctx context.Context, project, config string) error {
	if _, err := runDoppler(ctx, "configs", "get", config, "--project", project); err == nil {
		fmt.Fprintf(os.Stderr, "Doppler config %s exists in project %s\n", config, project)
		return nil
	}

	out, err := runDoppler(ctx, "configs", "create", config, "--project", project)
	if err != nil {
		return fmt.Errorf("create doppler config %s in project %s: %w (%s)", config, project, err, out)
	}

	fmt.Fprintf(os.Stderr, "Created Doppler config %s in project %s\n", config, project)
	return nil
}

func setupProjectConfig(ctx context.Context, project, config string) error {
	out, err := runDoppler(ctx, "setup", "--project", project, "--config", config, "--no-interactive")
	if err != nil {
		return fmt.Errorf("set doppler project/config (%s/%s): %w (%s)", project, config, err, out)
	}
	fmt.Fprintf(os.Stderr, "Configured Doppler for project %s with config %s\n", project, config)
	return nil
}

func runDoppler(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "doppler", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
