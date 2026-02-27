package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stax-ai/stax-cli/internal/run"
)

var (
	runConfigPath    string
	runPayloadFile   string
	runPayloadString string
	runInteractive   bool
)

func init() {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the automation handler once with a CloudEvent (one-shot)",
		Long: `Run loads config.yml (current or parent dir, or -c), builds a CloudEvent
with the given payload as event data, and invokes the project's handler via
"uv run python main.py" with the CloudEvent JSON on stdin. The handler must
be decorated with stax-sdk so it validates the CloudEvent.

You must provide exactly one of:
  -f, --payload-file   Path to a JSON file whose contents become event data
  --payload            Inline JSON object for event data
  -i, --interactive    Prompt for each field defined in config.yml`,
		Example: `  stax run --payload '{}'
  stax run -f data.json
  stax run -c ./project/config.yml --payload '{"key":"value"}'`,
		RunE: runRun,
	}
	runCmd.Flags().StringVarP(&runConfigPath, "config", "c", "", "path to config.yml (default: search in current and parent dirs)")
	runCmd.Flags().StringVarP(&runPayloadFile, "payload-file", "f", "", "path to JSON file for event data")
	runCmd.Flags().StringVar(&runPayloadString, "payload", "", "inline JSON object for event data")
	runCmd.Flags().BoolVarP(&runInteractive, "interactive", "i", false, "prompt for each field from config")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	payload, err := run.ResolvePayload(runPayloadFile, runPayloadString)
	if err != nil {
		return err
	}
	if payload == nil && !runInteractive {
		return fmt.Errorf("provide --payload-file, --payload, or --interactive")
	}

	ctx := context.Background()
	stdout, stderr, err := run.Run(ctx, runConfigPath, payload, runInteractive, nil)
	if err != nil {
		return err
	}
	if len(stderr) > 0 {
		os.Stderr.Write(stderr)
	}
	if len(stdout) > 0 {
		os.Stdout.Write(stdout)
	}
	return nil
}
