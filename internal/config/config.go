package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Mode represents the JWT authentication mode.
type Mode string

const (
	// ModeAuto generates tokens automatically using a private key.
	ModeAuto Mode = "auto"
	// ModeStatic uses a pre-generated token.
	ModeStatic Mode = "static"
)

// Config holds the configuration for the Tug daemon.
type Config struct {
	// Address to listen on (e.g. ":8080").
	ListenAddr string `yaml:"listenAddr"`

	// Slurmrestd configuration.
	Slurmrestd SlurmrestdConfig `yaml:"slurmrestd"`
}

// SlurmrestdConfig holds configuration for connecting to slurmrestd.
type SlurmrestdConfig struct {
	// URI is the base URL of slurmrestd.
	URI string `yaml:"uri"`

	// Version is the Slurm API version (e.g. "v0.0.44").
	Version string `yaml:"version"`

	// Auth configuration.
	JWTMode     Mode   `yaml:"jwtMode"`
	JWTUser     string `yaml:"jwtUser"`
	JWTLifespan int    `yaml:"jwtLifespan"` // in seconds
	JWTKey      string `yaml:"jwtKey"`      // path to private key file (for auto mode)
	JWTToken    string `yaml:"jwtToken"`    // static token (for static mode)
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		ListenAddr: ":8080",
		Slurmrestd: SlurmrestdConfig{
			URI:         "http://localhost:6820",
			Version:     "v0.0.40",
			JWTMode:     ModeAuto,
			JWTUser:     "slurm",
			JWTLifespan: 360,
			JWTKey:      "/etc/slurm/jwt_hs256.key",
		},
	}
}

// LoadConfigFromFile loads configuration from a YAML file.
func LoadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read config file")
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrap(err, "parse config file")
	}

	return &cfg, nil
}
