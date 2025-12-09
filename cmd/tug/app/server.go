package app

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/supergate-hub/tug/internal/server"
)

var (
	// These variables are populated by ldflags during build.
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Run runs the Tug application.
func Run() error {
	opts := NewOptions()
	
	// Parse flags
	opts.AddFlags(flag.CommandLine)
	flag.Parse()

	if opts.ShowVersion {
		log.Printf("Tug version: %s, commit: %s, built at: %s", Version, Commit, Date)
		return nil
	}

	// Load configuration
	if err := opts.LoadConfig(); err != nil {
		return err
	}

	cfg := opts.Config
	log.Printf("Starting Tug daemon (Version: %s, Target: %s, AuthMode: %s)",
		Version, cfg.Slurmrestd.URI, cfg.Slurmrestd.JWTMode)

	// Initialize Server
	srv, err := server.NewServer(cfg)
	if err != nil {
		return err
	}

	// Start Server with Graceful Shutdown
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
		return err
	}

	log.Println("Server stopped successfully")
	return nil
}

