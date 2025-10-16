package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config holds application-wide configuration loaded from YAML.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Mongo  MongoConfig  `yaml:"mongo"`
	Redis  RedisConfig  `yaml:"redis"`
}

// ServerConfig groups HTTP server settings.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// MongoConfig groups MongoDB connection settings.
type MongoConfig struct {
	URI            string `yaml:"uri"`
	Database       string `yaml:"database"`
	ConnectTimeout int    `yaml:"connectTimeoutSeconds"` // seconds
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	AuthSource     string `yaml:"authSource"`
}

// RedisConfig groups Redis connection settings.
type RedisConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	DialTimeout  int    `yaml:"dialTimeoutSeconds"`  // seconds
	ReadTimeout  int    `yaml:"readTimeoutSeconds"`  // seconds
	WriteTimeout int    `yaml:"writeTimeoutSeconds"` // seconds
}

var (
	once      sync.Once
	loadedCfg *Config
	loadErr   error
)

// Load reads configuration from the provided path. If path is empty, it resolves
// the path using the CONFIG_PATH env var or defaults to "config.yaml" in the
// project root. Load is safe to call multiple times; the first successful load
// wins and subsequent calls return the original result.
func Load(path string) error {
	once.Do(func() {
		resolvedPath := resolvePath(path)
		cfgBytes, err := os.ReadFile(resolvedPath)
		if err != nil {
			loadErr = err
			return
		}
		var cfg Config
		if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
			loadErr = err
			return
		}
		loadedCfg = &cfg
	})
	return loadErr
}

// MustLoad behaves like Load but panics if loading fails.
func MustLoad(path string) {
	if err := Load(path); err != nil {
		panic(err)
	}
}

// Get returns the loaded configuration. It returns an error if Load has not
// been called or failed.
func Get() (*Config, error) {
	if loadedCfg == nil {
		return nil, errors.New("config not loaded")
	}
	return loadedCfg, nil
}

// GetOrDefault returns the loaded configuration if available; otherwise it
// attempts to load using default resolution and returns that config or an error.
func GetOrDefault() (*Config, error) {
	if loadedCfg != nil {
		return loadedCfg, nil
	}
	if err := Load(""); err != nil {
		return nil, err
	}
	return loadedCfg, nil
}

func resolvePath(p string) string {
	if p != "" {
		return p
	}
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		return env
	}
	// Default to repo root config.yaml or cwd/config.yaml
	rootPath := filepath.Clean("config.yaml")
	if _, err := os.Stat(rootPath); err == nil {
		return rootPath
	}
	if errors.Is(os.ErrNotExist, fs.ErrNotExist) {
		return rootPath
	}
	return rootPath
}
