// Package build provides buildpack-based container image building for Stax automation projects.
package build

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	pack "github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/logging"
	"github.com/buildpacks/pack/pkg/project/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/heroku/color"
)

const (
	// DefaultBuilderImage is the default Python-capable builder (same as knative/func).
	DefaultBuilderImage = "ghcr.io/knative/builder-jammy-base:v2"
	// DefaultLifecycleImage is the buildpacks lifecycle image used by pack.
	DefaultLifecycleImage = "docker.io/buildpacksio/lifecycle:553c041"
)

// Trusted builder image prefixes (trailing slash required). See GHSA-5336-2g3f-9g3m.
var trustedBuilderPrefixes = []string{
	"quay.io/boson/",
	"gcr.io/paketo-buildpacks/",
	"docker.io/paketobuildpacks/",
	"gcr.io/buildpacks/",
	"ghcr.io/knative/",
	"docker.io/heroku/",
}

// Impl allows the pack build implementation to be mocked in tests.
type Impl interface {
	Build(context.Context, pack.BuildOptions) error
}

// Options configures a build.
type Options struct {
	// ProjectRoot is the path to the project (must contain e.g. main.py, pyproject.toml).
	ProjectRoot string
	// Image is the full OCI image name to build (e.g. myreg.io/myapp:latest).
	Image string
	// Builder is the builder image (default: DefaultBuilderImage).
	Builder string
	// Env is optional build-time env vars (e.g. BPE_DEFAULT_LISTEN_ADDRESS).
	Env map[string]string
	// Exclude is the list of paths to exclude from the build context (e.g. from .staxignore).
	Exclude []string
	// TrustBuilder when true trusts any builder image; otherwise only trusted prefixes are allowed.
	TrustBuilder bool
	// Verbose streams pack logs to stderr.
	Verbose bool
}

// Build produces an OCI image from the project at opts.ProjectRoot using Cloud Native Buildpacks.
// Requires a running Docker (or compatible) daemon.
func Build(ctx context.Context, opts Options, impl Impl) (err error) {
	if opts.ProjectRoot == "" {
		return fmt.Errorf("project root is required")
	}
	if opts.Image == "" {
		return fmt.Errorf("image name is required")
	}
	appPath, err := filepath.Abs(opts.ProjectRoot)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}
	if opts.Builder == "" {
		opts.Builder = DefaultBuilderImage
	}
	if !opts.TrustBuilder && !trustBuilder(opts.Builder) {
		return fmt.Errorf("builder image %q is not in the trusted list; use --trust-builder to override", opts.Builder)
	}

	env := make(map[string]string)
	if opts.Env != nil {
		for k, v := range opts.Env {
			env[k] = v
		}
	}
	if _, ok := env["BPE_DEFAULT_LISTEN_ADDRESS"]; !ok {
		env["BPE_DEFAULT_LISTEN_ADDRESS"] = "[::]:8080"
	}

	// Clean exclude list (no empty lines)
	var exclude []string
	for _, s := range opts.Exclude {
		s = strings.TrimSpace(s)
		if s != "" {
			exclude = append(exclude, s)
		}
	}

	buildOpts := pack.BuildOptions{
		AppPath:         appPath,
		Image:           opts.Image,
		Builder:         opts.Builder,
		Buildpacks:      nil, // use builder defaults
		LifecycleImage:  DefaultLifecycleImage,
		ProjectDescriptor: types.Descriptor{
			Build: types.Build{
				Exclude: exclude,
			},
		},
		ContainerConfig: struct {
			Network string
			Volumes []string
		}{Network: "", Volumes: nil},
		Env:          env,
		TrustBuilder: func(string) bool { return opts.TrustBuilder || trustBuilder(opts.Builder) },
	}
	if runtime.GOOS == "linux" {
		buildOpts.ContainerConfig.Network = "host"
	}

	var outBuf bytes.Buffer
	var logger logging.Logger
	if opts.Verbose {
		logger = logging.NewLogWithWriters(color.Stdout(), color.Stderr(), logging.WithVerbose())
	} else {
		logger = logging.NewSimpleLogger(&outBuf)
	}

	if impl == nil {
		cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("cannot create Docker client: %w (is Docker running?)", err)
		}
		defer cli.Close()

		packClient, err := pack.NewClient(pack.WithLogger(logger), pack.WithDockerClient(cli))
		if err != nil {
			return fmt.Errorf("cannot create pack client: %w", err)
		}
		impl = packClient
	}

	if err = impl.Build(ctx, buildOpts); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !opts.Verbose && outBuf.Len() > 0 {
			fmt.Fprintln(os.Stderr, "")
			_, _ = io.Copy(os.Stderr, &outBuf)
			fmt.Fprintln(os.Stderr, "")
		}
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}

func trustBuilder(builder string) bool {
	if isLocalhost(builder) {
		return true
	}
	for _, p := range trustedBuilderPrefixes {
		prefix := p
		if !strings.HasSuffix(prefix, "/") {
			prefix = prefix + "/"
		}
		if strings.HasPrefix(builder, prefix) {
			return true
		}
	}
	return false
}

func isLocalhost(img string) bool {
	localhostRE := regexp.MustCompile(`^(localhost|127\.0\.0\.1|\[::1\])(:\d+)?/.+$`)
	return localhostRE.MatchString(img)
}

// ReadStaxignore reads the project's .staxignore file and returns non-empty lines as exclude patterns.
// If the file does not exist, returns nil, nil.
func ReadStaxignore(projectRoot string) ([]string, error) {
	path := filepath.Join(projectRoot, ".staxignore")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out, nil
}
