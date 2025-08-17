package domain

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// User represents the user domain model
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Role         string    `json:"role" db:"role"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents an active user session
type Session struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// TokenResponse represents the authentication response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user"`
}

// RefreshRequest represents the token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// HTTPSignatureConfig represents HTTP signature configuration
type HTTPSignatureConfig struct {
	Algorithm    string        `json:"algorithm"`
	KeyStore     KeyStore      `json:"-"`
	MaxClockSkew time.Duration `json:"max_clock_skew"`
}

// KeyStore interface for managing cryptographic keys
type KeyStore interface {
	GetKey(keyID string) ([]byte, error)
	RotateKey(keyID string) error
	ValidateKey(keyID string) bool
}

// JWTClaims represents custom JWT claims
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// AuthContext represents the authenticated user context
type AuthContext struct {
	UserID    string
	Email     string
	Role      string
	SessionID string
	IPAddress string
	UserAgent string
}

// Permission represents a user permission
type Permission string

const (
	// User permissions
	PermissionViewProfile Permission = "user:view_profile"
	PermissionEditProfile Permission = "user:edit_profile"
	PermissionUploadDocs  Permission = "user:upload_documents"

	// Application permissions
	PermissionCreateApp Permission = "application:create"
	PermissionViewApp   Permission = "application:view"
	PermissionEditApp   Permission = "application:edit"
	PermissionSubmitApp Permission = "application:submit"

	// Decision permissions
	PermissionViewDecisions Permission = "decision:view"
	PermissionMakeDecision  Permission = "decision:make"
	PermissionViewQueue     Permission = "decision:view_queue"

	// Admin permissions
	PermissionManageUsers Permission = "admin:manage_users"
	PermissionViewAudit   Permission = "admin:view_audit"
	PermissionManageRules Permission = "admin:manage_rules"
)

// UserRole represents user role types
type UserRole string

const (
	RoleApplicant      UserRole = "applicant"
	RoleJuniorReviewer UserRole = "junior_reviewer"
	RoleSeniorReviewer UserRole = "senior_reviewer"
	RoleManager        UserRole = "manager"
	RoleAdmin          UserRole = "admin"
)

// GetPermissions returns permissions for a given role
func (r UserRole) GetPermissions() []Permission {
	switch r {
	case RoleApplicant:
		return []Permission{
			PermissionViewProfile,
			PermissionEditProfile,
			PermissionUploadDocs,
			PermissionCreateApp,
			PermissionViewApp,
			PermissionEditApp,
			PermissionSubmitApp,
		}
	case RoleJuniorReviewer:
		return []Permission{
			PermissionViewProfile,
			PermissionViewDecisions,
			PermissionMakeDecision,
			PermissionViewQueue,
		}
	case RoleSeniorReviewer:
		return []Permission{
			PermissionViewProfile,
			PermissionViewDecisions,
			PermissionMakeDecision,
			PermissionViewQueue,
			PermissionViewApp,
		}
	case RoleManager:
		return []Permission{
			PermissionViewProfile,
			PermissionViewDecisions,
			PermissionMakeDecision,
			PermissionViewQueue,
			PermissionViewApp,
			PermissionViewAudit,
			PermissionManageRules,
		}
	case RoleAdmin:
		return []Permission{
			PermissionViewProfile,
			PermissionEditProfile,
			PermissionViewDecisions,
			PermissionMakeDecision,
			PermissionViewQueue,
			PermissionViewApp,
			PermissionViewAudit,
			PermissionManageRules,
			PermissionManageUsers,
		}
	default:
		return []Permission{}
	}
}

// HasPermission checks if the role has a specific permission
func (r UserRole) HasPermission(permission Permission) bool {
	permissions := r.GetPermissions()
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
