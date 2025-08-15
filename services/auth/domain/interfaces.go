package domain

import (
	"context"
	"time"
)

// AuthService defines the authentication service interface
type AuthService interface {
	// Authentication
	Login(ctx context.Context, email, password string, ipAddress, userAgent string) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string, ipAddress, userAgent string) (*TokenResponse, error)
	Logout(ctx context.Context, userID, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
	
	// Token validation
	ValidateAccessToken(ctx context.Context, token string) (*AuthContext, error)
	ValidateRefreshToken(ctx context.Context, token string) (*Session, error)
	
	// HTTP Signature validation
	ValidateHTTPSignature(ctx context.Context, signature, keyID string, request []byte) error
	
	// User management
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	
	// Session management
	CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	InvalidateSession(ctx context.Context, sessionID string) error
	InvalidateUserSessions(ctx context.Context, userID string) error
	CleanExpiredSessions(ctx context.Context) error
	
	// Security
	CheckRateLimit(ctx context.Context, identifier string) error
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
}

// UserRepository defines the user data access interface
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, userID string) error
}

// SessionRepository defines the session data access interface
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByRefreshToken(ctx context.Context, token string) (*Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

// TokenManager defines the token management interface
type TokenManager interface {
	GenerateAccessToken(ctx context.Context, user *User, sessionID string) (string, time.Time, error)
	GenerateRefreshToken(ctx context.Context) (string, error)
	ValidateAccessToken(ctx context.Context, token string) (*JWTClaims, error)
	RevokeToken(ctx context.Context, token string) error
	IsTokenRevoked(ctx context.Context, token string) (bool, error)
}

// CacheService defines the caching interface
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	SetExpiration(ctx context.Context, key string, expiration time.Duration) error
}

// AuditLogger defines the audit logging interface
type AuditLogger interface {
	LogAuthEvent(ctx context.Context, event *AuthEvent) error
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
}

// AuthEvent represents an authentication audit event
type AuthEvent struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	EventType    string                 `json:"event_type"` // "login", "logout", "refresh", "failed_login"
	SessionID    string                 `json:"session_id,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Success      bool                   `json:"success"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// SecurityEvent represents a security-related audit event
type SecurityEvent struct {
	ID           string                 `json:"id"`
	EventType    string                 `json:"event_type"` // "suspicious_login", "rate_limit_exceeded", "invalid_signature"
	UserID       string                 `json:"user_id,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Severity     string                 `json:"severity"` // "low", "medium", "high", "critical"
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// AuthError represents authentication-specific errors
type AuthError struct {
	Code        string
	Message     string
	Description string
	Field       string
	Metadata    map[string]interface{}
}

func (e *AuthError) Error() string {
	return e.Message
}

// Authentication error codes
const (
	AUTH_001 = "AUTH_001" // Invalid credentials
	AUTH_002 = "AUTH_002" // Account locked
	AUTH_003 = "AUTH_003" // Account disabled
	AUTH_004 = "AUTH_004" // Invalid token
	AUTH_005 = "AUTH_005" // Token expired
	AUTH_006 = "AUTH_006" // Token revoked
	AUTH_007 = "AUTH_007" // Invalid refresh token
	AUTH_008 = "AUTH_008" // Session not found
	AUTH_009 = "AUTH_009" // Session expired
	AUTH_010 = "AUTH_010" // Rate limit exceeded
	AUTH_011 = "AUTH_011" // Invalid HTTP signature
	AUTH_012 = "AUTH_012" // Missing signature header
	AUTH_013 = "AUTH_013" // Key not found
	AUTH_014 = "AUTH_014" // Clock skew too large
	AUTH_015 = "AUTH_015" // Insufficient permissions
	AUTH_016 = "AUTH_016" // User not found
	AUTH_017 = "AUTH_017" // Database error
	AUTH_018 = "AUTH_018" // Cache error
	AUTH_019 = "AUTH_019" // Token generation failed
	AUTH_020 = "AUTH_020" // Invalid request format
)

// NewAuthError creates a new authentication error
func NewAuthError(code, message, description string) *AuthError {
	return &AuthError{
		Code:        code,
		Message:     message,
		Description: description,
		Metadata:    make(map[string]interface{}),
	}
}

// WithField adds a field to the error
func (e *AuthError) WithField(field string) *AuthError {
	e.Field = field
	return e
}

// WithMetadata adds metadata to the error
func (e *AuthError) WithMetadata(key string, value interface{}) *AuthError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}
