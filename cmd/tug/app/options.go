package app

import (
	"flag"
	"log"

	"github.com/pkg/errors"
	"github.com/supergate-hub/tug/internal/config"
)

// Options holds the command-line options and configuration.
type Options struct {
	ConfigPath  string
	ListenAddr  string
	SlurmURI    string
	JWTMode     string
	JWTKey      string
	ShowVersion bool

	// Config holds the loaded configuration.
	Config config.Config
}

// NewOptions creates a new Options with default values.
func NewOptions() *Options {
	return &Options{}
}

// AddFlags registers command-line flags.
func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.ConfigPath, "config", "", "Path to configuration file (yaml)")
	fs.StringVar(&o.ListenAddr, "listen-addr", "", "Address to listen on (overrides config file)")
	fs.StringVar(&o.SlurmURI, "slurm-uri", "", "Slurmrestd URI (overrides config file)")
	fs.StringVar(&o.JWTMode, "jwt-mode", "", "JWT mode: auto or static (overrides config file)")
	fs.StringVar(&o.JWTKey, "jwt-key", "", "Path to JWT private key (overrides config file)")
	fs.BoolVar(&o.ShowVersion, "version", false, "Show version information and exit")
}

// LoadConfig loads the configuration from file and overrides with flags.
func (o *Options) LoadConfig() error {
	var cfg config.Config
	var err error

	// 1. Load Base Configuration
	if o.ConfigPath != "" {
		var loadedCfg *config.Config
		loadedCfg, err = config.LoadConfigFromFile(o.ConfigPath)
		if err != nil {
			return errors.Wrap(err, "load config file")
		}
		cfg = *loadedCfg
		log.Printf("Loaded configuration from %s", o.ConfigPath)
	} else {
		cfg = config.DefaultConfig()
		log.Println("Using default configuration")
	}

	// 2. Override with Flags if provided
	if o.ListenAddr != "" {
		cfg.ListenAddr = o.ListenAddr
	}
	if o.SlurmURI != "" {
		cfg.Slurmrestd.URI = o.SlurmURI
	}
	if o.JWTMode != "" {
		cfg.Slurmrestd.JWTMode = config.Mode(o.JWTMode)
	}
	if o.JWTKey != "" {
		cfg.Slurmrestd.JWTKey = o.JWTKey
	}

	o.Config = cfg
	return nil
}

