package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/lendingplatform/los/services/auth/domain"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// HTTPSignatureValidator handles HTTP signature validation
type HTTPSignatureValidator struct {
	keyStore     domain.KeyStore
	logger       *zap.Logger
	localizer    *i18n.Localizer
	maxClockSkew time.Duration
}

// NewHTTPSignatureValidator creates a new HTTP signature validator
func NewHTTPSignatureValidator(
	keyStore domain.KeyStore,
	logger *zap.Logger,
	localizer *i18n.Localizer,
) *HTTPSignatureValidator {
	return &HTTPSignatureValidator{
		keyStore:     keyStore,
		logger:       logger,
		localizer:    localizer,
		maxClockSkew: time.Minute * 5, // 5 minutes clock skew tolerance
	}
}

// ValidateHTTPSignature validates HTTP signature according to RFC draft
func (h *HTTPSignatureValidator) ValidateHTTPSignature(ctx context.Context, signatureHeader, dateHeader, method, path string, body []byte) error {
	logger := h.logger.With(
		zap.String("operation", "validate_http_signature"),
		zap.String("method", method),
		zap.String("path", path),
	)

	if signatureHeader == "" {
		logger.Warn("Missing signature header")
		return domain.NewAuthError(domain.AUTH_012,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.missing_signature"}),
			"Signature header is required")
	}

	// Parse signature header
	sigParams, err := h.parseSignatureHeader(signatureHeader)
	if err != nil {
		logger.Warn("Invalid signature header format", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_011,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_signature"}),
			"Invalid signature header format")
	}

	// Validate required parameters
	keyID, exists := sigParams["keyId"]
	if !exists {
		logger.Warn("Missing keyId in signature")
		return domain.NewAuthError(domain.AUTH_011,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_signature"}),
			"Missing keyId in signature")
	}

	algorithm, exists := sigParams["algorithm"]
	if !exists {
		algorithm = "hmac-sha256" // Default algorithm
	}

	signature, exists := sigParams["signature"]
	if !exists {
		logger.Warn("Missing signature in header")
		return domain.NewAuthError(domain.AUTH_011,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_signature"}),
			"Missing signature value")
	}

	headers, exists := sigParams["headers"]
	if !exists {
		headers = "date" // Default to date header only
	}

	// Validate clock skew
	if err := h.validateClockSkew(dateHeader); err != nil {
		logger.Warn("Clock skew too large", zap.Error(err))
		return err
	}

	// Get signing key
	key, err := h.keyStore.GetKey(keyID)
	if err != nil {
		logger.Error("Failed to get signing key", zap.Error(err), zap.String("key_id", keyID))
		return domain.NewAuthError(domain.AUTH_013,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.key_not_found"}),
			"Signing key not found")
	}

	// Build signature string
	sigString, err := h.buildSignatureString(headers, method, path, dateHeader, body)
	if err != nil {
		logger.Error("Failed to build signature string", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_011,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_signature"}),
			"Failed to build signature string")
	}

	// Verify signature
	if err := h.verifySignature(algorithm, key, sigString, signature); err != nil {
		logger.Warn("Signature verification failed", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_011,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_signature"}),
			"Signature verification failed")
	}

	logger.Debug("HTTP signature validated successfully", zap.String("key_id", keyID))
	return nil
}

// parseSignatureHeader parses the Signature header into components
func (h *HTTPSignatureValidator) parseSignatureHeader(header string) (map[string]string, error) {
	params := make(map[string]string)
	
	// Split by comma and parse key=value pairs
	pairs := strings.Split(header, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		// Find the first = sign
		eqIndex := strings.Index(pair, "=")
		if eqIndex == -1 {
			return nil, fmt.Errorf("invalid parameter format: %s", pair)
		}
		
		key := strings.TrimSpace(pair[:eqIndex])
		value := strings.TrimSpace(pair[eqIndex+1:])
		
		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		
		params[key] = value
	}
	
	return params, nil
}

// validateClockSkew checks if the date header is within acceptable range
func (h *HTTPSignatureValidator) validateClockSkew(dateHeader string) error {
	if dateHeader == "" {
		return domain.NewAuthError(domain.AUTH_014,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.missing_date"}),
			"Date header is required")
	}

	requestTime, err := time.Parse(time.RFC1123, dateHeader)
	if err != nil {
		// Try alternative formats
		if requestTime, err = time.Parse(time.RFC3339, dateHeader); err != nil {
			return domain.NewAuthError(domain.AUTH_014,
				h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_date"}),
				"Invalid date format")
		}
	}

	now := time.Now()
	timeDiff := now.Sub(requestTime)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	if timeDiff > h.maxClockSkew {
		return domain.NewAuthError(domain.AUTH_014,
			h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.clock_skew"}),
			"Request time is too far from current time")
	}

	return nil
}

// buildSignatureString constructs the string to be signed
func (h *HTTPSignatureValidator) buildSignatureString(headers, method, path, date string, body []byte) (string, error) {
	var parts []string
	headerList := strings.Split(headers, " ")
	
	for _, header := range headerList {
		header = strings.ToLower(strings.TrimSpace(header))
		
		switch header {
		case "request-line":
			requestLine := fmt.Sprintf("%s %s HTTP/1.1", strings.ToUpper(method), path)
			parts = append(parts, fmt.Sprintf("request-line: %s", requestLine))
		case "date":
			parts = append(parts, fmt.Sprintf("date: %s", date))
		case "content-length":
			contentLength := strconv.Itoa(len(body))
			parts = append(parts, fmt.Sprintf("content-length: %s", contentLength))
		case "digest":
			if len(body) > 0 {
				hash := sha256.Sum256(body)
				digest := base64.StdEncoding.EncodeToString(hash[:])
				parts = append(parts, fmt.Sprintf("digest: SHA-256=%s", digest))
			}
		default:
			// Skip unknown headers for now
			h.logger.Debug("Skipping unknown header in signature", zap.String("header", header))
		}
	}
	
	return strings.Join(parts, "\n"), nil
}

// verifySignature verifies the signature using the specified algorithm
func (h *HTTPSignatureValidator) verifySignature(algorithm string, key []byte, message, signature string) error {
	expectedSig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid base64 signature: %w", err)
	}

	switch algorithm {
	case "hmac-sha256":
		mac := hmac.New(sha256.New, key)
		mac.Write([]byte(message))
		computedSig := mac.Sum(nil)
		
		if subtle.ConstantTimeCompare(expectedSig, computedSig) != 1 {
			return fmt.Errorf("signature mismatch")
		}
		return nil
		
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// Additional authentication service methods

// CheckRateLimit checks if the request should be rate limited
func (s *AuthService) CheckRateLimit(ctx context.Context, identifier string) error {
	key := "rate_limit:auth:" + identifier
	
	// Get current count
	count, err := s.cache.Get(ctx, key)
	if err != nil && err.Error() != "key not found" {
		s.logger.Error("Failed to get rate limit count", zap.Error(err))
		return nil // Allow request if cache is unavailable
	}
	
	currentCount := int64(0)
	if count != nil {
		if c, ok := count.(int64); ok {
			currentCount = c
		}
	}
	
	// Check limit (100 requests per hour)
	limit := int64(100)
	if currentCount >= limit {
		return domain.NewAuthError(domain.AUTH_010,
			s.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.rate_limit_exceeded"}),
			"Too many authentication requests. Please try again later.")
	}
	
	// Increment counter
	if _, err := s.cache.Increment(ctx, key); err != nil {
		s.logger.Error("Failed to increment rate limit counter", zap.Error(err))
	}
	
	// Set expiration for new counter
	if currentCount == 0 {
		s.cache.SetExpiration(ctx, key, time.Hour)
	}
	
	return nil
}

// CreateSession creates a new user session
func (s *AuthService) CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (*domain.Session, error) {
	refreshToken, err := s.tokenManager.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:           generateSessionID(),
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.sessionDuration),
		CreatedAt:    time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}

// InvalidateSession invalidates a specific session
func (s *AuthService) InvalidateSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// InvalidateUserSessions invalidates all sessions for a user
func (s *AuthService) InvalidateUserSessions(ctx context.Context, userID string) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// CleanExpiredSessions removes expired sessions
func (s *AuthService) CleanExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpired(ctx)
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// UpdateLastLogin updates the user's last login timestamp
func (s *AuthService) UpdateLastLogin(ctx context.Context, userID string) error {
	return s.userRepo.UpdateLastLogin(ctx, userID)
}

// LogSecurityEvent logs a security-related event
func (s *AuthService) LogSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error {
	return s.auditLogger.LogSecurityEvent(ctx, event)
}

// ValidateHTTPSignature validates HTTP signature
func (s *AuthService) ValidateHTTPSignature(ctx context.Context, signature, keyID string, request []byte) error {
	// This would integrate with the HTTPSignatureValidator
	// For now, return a placeholder implementation
	return domain.NewAuthError(domain.AUTH_011,
		s.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.signature_not_implemented"}),
		"HTTP signature validation not implemented")
}

// Helper function to generate session ID
func generateSessionID() string {
	return "sess_" + generateRandomString(32)
}

// Helper function to generate random string
func generateRandomString(length int) string {
	// This should use crypto/rand for production
	return "random_string_placeholder"
}
