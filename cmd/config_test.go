package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stax-ai/stax-cli/cmd"
	"github.com/stretchr/testify/suite"
)

type ConfigCmdSuite struct {
	suite.Suite
	configPath string
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
}

func TestConfigCmdSuite(t *testing.T) {
	suite.Run(t, new(ConfigCmdSuite))
}

func (s *ConfigCmdSuite) SetupTest() {
	dir := s.T().TempDir()
	s.configPath = filepath.Join(dir, ".stax.yaml")
	s.stdout = &bytes.Buffer{}
	s.stderr = &bytes.Buffer{}
}

func (s *ConfigCmdSuite) run(args ...string) error {
	fullArgs := append([]string{"--config-file", s.configPath}, args...)
	cmd.RootCmd().SetArgs(fullArgs)
	// Commands use fmt.Println/os.Stderr; capture via redirect
	oldOut, oldErr := os.Stdout, os.Stderr
	defer func() { os.Stdout, os.Stderr = oldOut, oldErr }()
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout = outW
	os.Stderr = errW
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.RootCmd().Execute()
		outW.Close()
		errW.Close()
	}()
	s.stdout.Reset()
	s.stderr.Reset()
	_, _ = s.stdout.ReadFrom(outR)
	_, _ = s.stderr.ReadFrom(errR)
	return <-errCh
}

func (s *ConfigCmdSuite) TestConfigGet_emptyConfig_printsEmptyObject() {
	s.Require().NoError(s.run("config", "get"))
	s.Equal("{}\n", s.stdout.String())
	s.Empty(s.stderr.String())
}

func (s *ConfigCmdSuite) TestConfigGet_missingKey_printsNotSet() {
	s.Require().NoError(s.run("config", "get", "api_url"))
	s.Equal("not set\n", s.stderr.String())
}

func (s *ConfigCmdSuite) TestConfigSet_andGet_roundtrip() {
	s.Require().NoError(s.run("config", "set", "api_url", "https://api.example.com"))
	s.stdout.Reset()
	s.stderr.Reset()
	s.Require().NoError(s.run("config", "get", "api_url"))
	s.Equal("https://api.example.com\n", s.stdout.String())

	// Full config should include the key
	s.stdout.Reset()
	s.Require().NoError(s.run("config", "get"))
	s.Contains(s.stdout.String(), "api.example.com")
}

func (s *ConfigCmdSuite) TestConfigSet_createsFileWhenMissing() {
	s.Require().NoError(s.run("config", "set", "profile", "dev"))
	_, err := os.Stat(s.configPath)
	s.Require().NoError(err)
}
