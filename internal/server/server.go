package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	slurmv0040jobs "github.com/supergate-hub/slurm-client/pkg/slurm/v0040/jobs"

	"github.com/supergate-hub/slurm-client/pkg/client"
	"github.com/supergate-hub/tug/internal/auth"
	"github.com/supergate-hub/tug/internal/config"
)

// key is an unexported type for context keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

const (
	// userNameKey is the context key for Slurm user name.
	userNameKey key = iota
)

// Server represents the Tug HTTP server.
type Server struct {
	cfg         config.Config
	auth        *auth.Authenticator
	slurmClient *client.UnifiedClient
	httpServer  *http.Server
}

// NewServer creates a new Tug server.
func NewServer(cfg config.Config) (*Server, error) {
	// Initialize Authenticator
	authenticator, err := auth.NewAuthenticator(cfg.Slurmrestd)
	if err != nil {
		return nil, errors.Wrap(err, "initialize authenticator")
	}

	// Initialize Slurm Client (Rule 6.1)
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	// Check if URI scheme is unix
	if strings.HasPrefix(cfg.Slurmrestd.URI, "unix://") {
		socketPath := strings.TrimPrefix(cfg.Slurmrestd.URI, "unix://")

		// Setup custom dialer for Unix socket
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}
	}

	clientCfg := client.Config{
		Server:  cfg.Slurmrestd.URI,
		Version: cfg.Slurmrestd.Version,
		// AuthToken is required for validation but overridden by authTransport.
		AuthToken: "dummy-token-for-validation",
		HTTPClient: &http.Client{
			Transport: &authTransport{
				base: transport, // Use our configured transport (which may have Unix dialer)
				auth: authenticator,
			},
			Timeout: 10 * time.Second, // Add explicit timeout
		},
	}

	slurmCli, err := client.NewUnifiedClient(clientCfg)
	if err != nil {
		return nil, errors.Wrap(err, "create slurm client")
	}

	return &Server{
		cfg:         cfg,
		auth:        authenticator,
		slurmClient: slurmCli,
	}, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/job/submit", s.handleJobSubmit)
	// Add more handlers as needed

	s.httpServer = &http.Server{
		Addr:    s.cfg.ListenAddr,
		Handler: mux,
	}

	log.Printf("Tug daemon listening on %s", s.cfg.ListenAddr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.WithStack(err)
	}
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// authTransport injects authentication headers into requests.
type authTransport struct {
	base http.RoundTripper
	auth *auth.Authenticator
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Retrieve user from context, defaulting to config if not present
	var username string
	if val := req.Context().Value(userNameKey); val != nil {
		username = val.(string)
	}

	headers := t.auth.Headers(username)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Remove dummy Authorization header injected by SDK
	req.Header.Del("Authorization")

	// [DEBUG] Print headers
	log.Printf("Sending request to %s", req.URL.String())
	for k, v := range req.Header {
		log.Printf("Header: %s = %v", k, v)
	}

	return t.base.RoundTrip(req)
}

// handleJobSubmit proxies the job submission to slurmrestd.
// Client Request: POST /job/submit
// Tug Action: Parse request -> Call slurmClient -> Return response
func (s *Server) handleJobSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract X-SLURM-USER-NAME from request header
	targetUser := r.Header.Get("X-SLURM-USER-NAME")

	// Read body (Rule 6.2: Ensure body is closed)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	// r.Body is automatically closed by the server, but explicit Close is fine if we were a client.
	// As a server handler, we don't strictly need to close r.Body, but it's good practice to drain it if we didn't read all.
	// Here we read all with io.ReadAll.

	// Parse body into v0040 SubmitOpts for now
	// In a real multi-version scenario, we would switch based on s.cfg.Slurmrestd.Version
	var submitReq slurmv0040jobs.SubmitOpts

	// We assume the body IS the SubmitOpts JSON (containing "script", "job", "jobs" etc.)
	if err := json.Unmarshal(body, &submitReq); err != nil {
		log.Printf("failed to unmarshal body: %v", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Create context with user info
	ctx := r.Context()
	if targetUser != "" {
		ctx = context.WithValue(ctx, userNameKey, targetUser)
	} else {
		http.Error(w, "X-SLURM-USER-NAME header is required", http.StatusBadRequest)
		return
	}

	// Call SDK with the new context
	resp, err := s.slurmClient.Slurm.Jobs().Submit(ctx, submitReq)
	if err != nil {
		log.Printf("submit failed: %v", err)
		// Don't expose internal error details ideally, but for now we keep it
		http.Error(w, fmt.Sprintf("submit failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
