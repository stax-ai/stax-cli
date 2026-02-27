package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) TestInit_withExplicitFile_validYAML() {
	dir := s.T().TempDir()
	path := filepath.Join(dir, ".stax.yaml")
	err := os.WriteFile(path, []byte("api_url: https://api.example.com\nprofile: dev\n"), 0644)
	s.Require().NoError(err)

	s.Require().NoError(Init(path))
	c, err := Get()
	s.Require().NoError(err)
	s.Equal("https://api.example.com", c.APIURL)
	s.Equal("dev", c.Profile)
}

func (s *ConfigSuite) TestInit_withExplicitFile_notFound_returnsNil() {
	err := Init(filepath.Join(s.T().TempDir(), "nonexistent.yaml"))
	s.NoError(err)
}

func (s *ConfigSuite) TestInit_withExplicitFile_invalidYAML_returnsError() {
	dir := s.T().TempDir()
	path := filepath.Join(dir, ".stax.yaml")
	err := os.WriteFile(path, []byte("invalid: [[[ yaml"), 0644)
	s.Require().NoError(err)

	err = Init(path)
	s.Error(err)
}

func (s *ConfigSuite) TestInit_withEmptyFile_searchesCurrentDir() {
	dir := s.T().TempDir()
	path := filepath.Join(dir, ".stax.yaml")
	err := os.WriteFile(path, []byte("api_url: from-current\n"), 0644)
	s.Require().NoError(err)

	prev, err := os.Getwd()
	s.Require().NoError(err)
	s.Require().NoError(os.Chdir(dir))
	defer func() { _ = os.Chdir(prev) }()

	s.Require().NoError(Init(""))
	c, err := Get()
	s.Require().NoError(err)
	s.Equal("from-current", c.APIURL)
}

func (s *ConfigSuite) TestGet_emptyViper_returnsEmptyConfig() {
	c, err := Get()
	s.Require().NoError(err)
	s.NotNil(c)
	s.Empty(c.APIURL)
	s.Empty(c.Profile)
}

func (s *ConfigSuite) TestDefaultWritePath() {
	p := DefaultWritePath()
	s.NotEmpty(p)
	s.Equal(".stax.yaml", filepath.Base(p))
}
