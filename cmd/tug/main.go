package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/supergate-hub/tug/internal/config"
	"github.com/supergate-hub/tug/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Flags holds command-line flags.
type Flags struct {
	ConfigPath string
	ListenAddr string
	SlurmURI   string
	JWTMode    string
	JWTKey     string
	ShowVersion bool
}

func parseFlags() *Flags {
	f := &Flags{}
	flag.StringVar(&f.ConfigPath, "config", "", "Path to configuration file (yaml)")
	flag.StringVar(&f.ListenAddr, "listen-addr", "", "Address to listen on (overrides config file)")
	flag.StringVar(&f.SlurmURI, "slurm-uri", "", "Slurmrestd URI (overrides config file)")
	flag.StringVar(&f.JWTMode, "jwt-mode", "", "JWT mode: auto or static (overrides config file)")
	flag.StringVar(&f.JWTKey, "jwt-key", "", "Path to JWT private key (overrides config file)")
	flag.BoolVar(&f.ShowVersion, "version", false, "Show version information and exit")

	flag.Parse()
	return f
}

func loadConfig(f *Flags) (config.Config, error) {
	var cfg config.Config
	var err error

	// 1. Load Base Configuration
	if f.ConfigPath != "" {
		var loadedCfg *config.Config
		loadedCfg, err = config.LoadConfigFromFile(f.ConfigPath)
		if err != nil {
			return cfg, errors.Wrap(err, "load config file")
		}
		cfg = *loadedCfg
		log.Printf("Loaded configuration from %s", f.ConfigPath)
	} else {
		cfg = config.DefaultConfig()
		log.Println("Using default configuration")
	}

	// 2. Override with Flags if provided
	if f.ListenAddr != "" {
		cfg.ListenAddr = f.ListenAddr
	}
	if f.SlurmURI != "" {
		cfg.Slurmrestd.URI = f.SlurmURI
	}
	if f.JWTMode != "" {
		cfg.Slurmrestd.JWTMode = config.Mode(f.JWTMode)
	}
	if f.JWTKey != "" {
		cfg.Slurmrestd.JWTKey = f.JWTKey
	}

	return cfg, nil
}

func main() {
	// 1. Parse Flags
	flags := parseFlags()

	if flags.ShowVersion {
		log.Printf("Tug version: %s, commit: %s, built at: %s", version, commit, date)
		return
	}

	// 2. Load Configuration
	cfg, err := loadConfig(flags)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	log.Printf("Starting Tug daemon (Version: %s, Target: %s, AuthMode: %s)",
		version, cfg.Slurmrestd.URI, cfg.Slurmrestd.JWTMode)

	// 3. Initialize Server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	// 4. Start Server with Graceful Shutdown
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Wait for signal
	sig := <-stopCh
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("Server stopped successfully")
}
