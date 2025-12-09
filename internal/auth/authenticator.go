package auth

import (
	"log"
	"os"
	"time"

	"github.com/supergate-hub/tug/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// TokenProvider defines the interface for providing JWT tokens.
type TokenProvider interface {
	GetToken() (string, error)
	GetTokenForUser(username string) (string, error)
	GetUser() string
	Headers(username string) map[string]string
}

// Authenticator manages JWT tokens for slurmrestd.
type Authenticator struct {
	cfg          config.SlurmrestdConfig
	jwtKey       []byte
	currentToken string
	expiration   int64
}

// NewAuthenticator creates a new Authenticator based on the configuration.
func NewAuthenticator(cfg config.SlurmrestdConfig) (*Authenticator, error) {
	auth := &Authenticator{
		cfg: cfg,
	}

	switch cfg.JWTMode {
	case config.ModeStatic:
		if cfg.JWTToken == "" {
			return nil, errors.New("missing token in configuration for static mode")
		}
		auth.currentToken = cfg.JWTToken

		// Validate static token expiration
		token, _, err := new(jwt.Parser).ParseUnverified(cfg.JWTToken, jwt.MapClaims{})
		if err != nil {
			return nil, errors.Wrap(err, "invalid static token")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				auth.expiration = int64(exp)
			}
		}
	case config.ModeAuto:
		// Load private key for signing
		keyData, err := os.ReadFile(cfg.JWTKey)
		if err != nil {
			return nil, errors.Wrapf(err, "read JWT key file %s", cfg.JWTKey)
		}
		auth.jwtKey = keyData
	default:
		return nil, errors.Errorf("unsupported JWT mode: %s", cfg.JWTMode)
	}

	return auth, nil
}

// GetToken returns a valid JWT token for the configured user.
func (a *Authenticator) GetToken() (string, error) {
	return a.GetTokenForUser(a.cfg.JWTUser)
}

// GetTokenForUser returns a valid JWT token for the specified user.
func (a *Authenticator) GetTokenForUser(username string) (string, error) {
	if a.cfg.JWTMode == config.ModeStatic {
		a.checkStaticExpiration()
		return a.currentToken, nil
	}

	// Auto mode
	// Note: For simplicity, we currently don't cache per-user tokens.
	// We generate a fresh token every time or if expired.
	// To support caching, we'd need a map[username]token.
	// For now, let's regenerate if username differs or token expired.

	// If we want to strictly follow the previous logic of 60s renewal:
	// We should ideally track tokens per user.

	// But since the requirement is "generate token based on header",
	// generating a new token (HS256 calculation) is very cheap.
	// Let's generate it on demand for now to keep it stateless per user.

	return a.generateToken(username)
}

// GetUser returns the configured JWT user name.
func (a *Authenticator) GetUser() string {
	return a.cfg.JWTUser
}

// Headers returns the HTTP headers required for authentication.
// If username is empty, uses the configured default user.
func (a *Authenticator) Headers(username string) map[string]string {
	targetUser := username
	if targetUser == "" {
		targetUser = a.cfg.JWTUser
	}

	token, err := a.GetTokenForUser(targetUser)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return map[string]string{}
	}

	return map[string]string{
		"X-SLURM-USER-NAME":  targetUser,
		"X-SLURM-USER-TOKEN": token,
	}
}

func (a *Authenticator) checkStaticExpiration() {
	gap := a.expiration - time.Now().Unix()
	if gap < 0 {
		log.Println("Error: Static JWT for slurmrestd authentication is expired")
	} else if gap < 3600 {
		log.Println("Warning: Static JWT for slurmrestd authentication will expire soon")
	}
}

func (a *Authenticator) generateToken(username string) (string, error) {
	now := time.Now()
	expiration := now.Add(time.Duration(a.cfg.JWTLifespan) * time.Second)

	// Create claims
	claims := jwt.MapClaims{
		"sun": username, // Slurm User Name
		"exp": expiration.Unix(),
		"iat": now.Unix(),
	}

	// Slurm usually uses HS256 with a shared key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(a.jwtKey)
	if err != nil {
		return "", errors.Wrap(err, "sign token")
	}

	// Note: We are not updating a.currentToken here because it's per-user.
	return signedToken, nil
}
