package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pack "github.com/buildpacks/pack/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild_RequiresProjectRoot(t *testing.T) {
	err := Build(context.Background(), Options{Image: "myreg.io/app:latest"}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project root")
}

func TestBuild_RequiresImage(t *testing.T) {
	dir := t.TempDir()
	err := Build(context.Background(), Options{ProjectRoot: dir}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image")
}

func TestBuild_UntrustedBuilderRejected(t *testing.T) {
	dir := t.TempDir()
	err := Build(context.Background(), Options{
		ProjectRoot:  dir,
		Image:        "myreg.io/app:latest",
		Builder:      "evil.io/random-builder:latest",
		TrustBuilder: false,
	}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in the trusted list")
}

func TestBuild_TrustBuilderBypass(t *testing.T) {
	// With TrustBuilder, any builder is accepted; we use a mock so we don't need Docker.
	dir := t.TempDir()
	called := false
	mock := &mockImpl{
		build: func(ctx context.Context, opts pack.BuildOptions) error {
			called = true
			assert.Equal(t, "evil.io/builder:1", opts.Builder)
			assert.Equal(t, dir, opts.AppPath)
			assert.Equal(t, "myreg.io/app:latest", opts.Image)
			return nil
		},
	}
	err := Build(context.Background(), Options{
		ProjectRoot:   dir,
		Image:         "myreg.io/app:latest",
		Builder:       "evil.io/builder:1",
		TrustBuilder:  true,
		Verbose:       false,
	}, mock)
	require.NoError(t, err)
	assert.True(t, called)
}

func TestBuild_DefaultBuilderAndExcludes(t *testing.T) {
	dir := t.TempDir()
	var captured pack.BuildOptions
	mock := &mockImpl{
		build: func(ctx context.Context, opts pack.BuildOptions) error {
			captured = opts
			return nil
		},
	}
	err := Build(context.Background(), Options{
		ProjectRoot: dir,
		Image:       "reg.io/img:latest",
		Builder:     "", // default
		Exclude:     []string{"  .git  ", "", "tests"},
	}, mock)
	require.NoError(t, err)
	assert.Equal(t, DefaultBuilderImage, captured.Builder)
	assert.Equal(t, []string{".git", "tests"}, captured.ProjectDescriptor.Build.Exclude)
	assert.Equal(t, "[::]:8080", captured.Env["BPE_DEFAULT_LISTEN_ADDRESS"])
}

func TestReadStaxignore_MissingFile(t *testing.T) {
	dir := t.TempDir()
	excludes, err := ReadStaxignore(dir)
	require.NoError(t, err)
	assert.Nil(t, excludes)
}

func TestReadStaxignore_ReadsAndFilters(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".staxignore")
	err := os.WriteFile(path, []byte(".git\n\n# comment\n__pycache__\n  .venv  \n"), 0644)
	require.NoError(t, err)
	excludes, err := ReadStaxignore(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{".git", "__pycache__", ".venv"}, excludes)
}

func TestTrustBuilder_Localhost(t *testing.T) {
	assert.True(t, trustBuilder("localhost:5000/my-builder:1"))
	assert.True(t, trustBuilder("127.0.0.1:5000/b:1"))
}

func TestTrustBuilder_KnownPrefixes(t *testing.T) {
	assert.True(t, trustBuilder("ghcr.io/knative/builder-jammy-base:v2"))
	assert.True(t, trustBuilder("docker.io/paketobuildpacks/builder-jammy-base"))
	assert.False(t, trustBuilder("evil.io/random:1"))
}

type mockImpl struct {
	build func(context.Context, pack.BuildOptions) error
}

func (m *mockImpl) Build(ctx context.Context, opts pack.BuildOptions) error {
	return m.build(ctx, opts)
}
