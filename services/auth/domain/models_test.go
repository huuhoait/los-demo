package domain

import (
	"testing"
	"time"

	"github.com/lendingplatform/los/services/auth/domain"
	"github.com/stretchr/testify/assert"
)

func TestUserRole_GetPermissions(t *testing.T) {
	tests := []struct {
		name     string
		role     domain.UserRole
		expected []domain.Permission
	}{
		{
			name: "applicant role",
			role: RoleApplicant,
			expected: []Permission{
				PermissionViewProfile,
				PermissionEditProfile,
				PermissionUploadDocs,
				PermissionCreateApp,
				PermissionViewApp,
				PermissionEditApp,
				PermissionSubmitApp,
			},
		},
		{
			name: "junior reviewer role",
			role: RoleJuniorReviewer,
			expected: []Permission{
				PermissionViewProfile,
				PermissionViewDecisions,
				PermissionMakeDecision,
				PermissionViewQueue,
			},
		},
		{
			name: "admin role",
			role: RoleAdmin,
			expected: []Permission{
				PermissionViewProfile,
				PermissionEditProfile,
				PermissionViewDecisions,
				PermissionMakeDecision,
				PermissionViewQueue,
				PermissionViewApp,
				PermissionViewAudit,
				PermissionManageRules,
				PermissionManageUsers,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := tt.role.GetPermissions()
			assert.ElementsMatch(t, tt.expected, permissions)
		})
	}
}

func TestUserRole_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       domain.UserRole
		permission domain.Permission
		expected   bool
	}{
		{
			name:       "applicant can view profile",
			role:       domain.RoleApplicant,
			permission: domain.PermissionViewProfile,
			expected:   true,
		},
		{
			name:       "applicant cannot manage users",
			role:       domain.RoleApplicant,
			permission: domain.PermissionManageUsers,
			expected:   false,
		},
		{
			name:       "admin can manage users",
			role:       domain.RoleAdmin,
			permission: domain.PermissionManageUsers,
			expected:   true,
		},
		{
			name:       "junior reviewer can make decisions",
			role:       domain.RoleJuniorReviewer,
			permission: domain.PermissionMakeDecision,
			expected:   true,
		},
		{
			name:       "junior reviewer cannot view audit",
			role:       domain.RoleJuniorReviewer,
			permission: domain.PermissionViewAudit,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.HasPermission(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthError(t *testing.T) {
	t.Run("create auth error", func(t *testing.T) {
		err := domain.NewAuthError(domain.AUTH_001, "Invalid credentials", "Email or password is incorrect")

		assert.Equal(t, domain.AUTH_001, err.Code)
		assert.Equal(t, "Invalid credentials", err.Message)
		assert.Equal(t, "Email or password is incorrect", err.Description)
		assert.Equal(t, "Invalid credentials", err.Error())
	})

	t.Run("auth error with field", func(t *testing.T) {
		err := domain.NewAuthError(domain.AUTH_001, "Invalid email", "Email format is invalid").
			WithField("email")

		assert.Equal(t, "email", err.Field)
	})

	t.Run("auth error with metadata", func(t *testing.T) {
		err := domain.NewAuthError(domain.AUTH_010, "Rate limit exceeded", "Too many requests").
			WithMetadata("retry_after", 60).
			WithMetadata("max_requests", 100)

		assert.Equal(t, int64(60), err.Metadata["retry_after"])
		assert.Equal(t, int64(100), err.Metadata["max_requests"])
	})
}

func TestJWTClaims(t *testing.T) {
	t.Run("create valid JWT claims", func(t *testing.T) {
		claims := &domain.JWTClaims{
			UserID:    "user_123",
			Email:     "test@example.com",
			Role:      "applicant",
			SessionID: "session_456",
		}

		assert.Equal(t, "user_123", claims.UserID)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.Equal(t, "applicant", claims.Role)
		assert.Equal(t, "session_456", claims.SessionID)
	})
}

func TestSession(t *testing.T) {
	t.Run("create session", func(t *testing.T) {
		now := time.Now()
		session := &domain.Session{
			ID:           "session_123",
			UserID:       "user_456",
			RefreshToken: "refresh_token_789",
			ExpiresAt:    now.Add(24 * time.Hour),
			CreatedAt:    now,
			IPAddress:    "192.168.1.1",
			UserAgent:    "Mozilla/5.0...",
		}

		assert.Equal(t, "session_123", session.ID)
		assert.Equal(t, "user_456", session.UserID)
		assert.Equal(t, "refresh_token_789", session.RefreshToken)
		assert.True(t, session.ExpiresAt.After(now))
		assert.Equal(t, "192.168.1.1", session.IPAddress)
	})
}

func TestUser(t *testing.T) {
	t.Run("create user", func(t *testing.T) {
		now := time.Now()
		user := &domain.User{
			ID:           "user_123",
			Email:        "test@example.com",
			PasswordHash: "$2a$12$...",
			FirstName:    "John",
			LastName:     "Doe",
			Role:         "applicant",
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		assert.Equal(t, "user_123", user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "applicant", user.Role)
		assert.Equal(t, "active", user.Status)
	})
}

func TestTokenResponse(t *testing.T) {
	t.Run("create token response", func(t *testing.T) {
		user := &domain.User{
			ID:        "user_123",
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role:      "applicant",
			Status:    "active",
		}

		expiresAt := time.Now().Add(15 * time.Minute)
		response := &domain.TokenResponse{
			AccessToken:  "access_token_123",
			RefreshToken: "refresh_token_456",
			TokenType:    "Bearer",
			ExpiresIn:    900, // 15 minutes
			ExpiresAt:    expiresAt,
			User:         user,
		}

		assert.Equal(t, "access_token_123", response.AccessToken)
		assert.Equal(t, "refresh_token_456", response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, int64(900), response.ExpiresIn)
		assert.Equal(t, user, response.User)
	})
}

func TestAuthEvent(t *testing.T) {
	t.Run("create auth event", func(t *testing.T) {
		now := time.Now()
		event := &domain.AuthEvent{
			ID:        "event_123",
			UserID:    "user_456",
			EventType: "login",
			SessionID: "session_789",
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0...",
			Success:   true,
			Timestamp: now,
		}

		assert.Equal(t, "event_123", event.ID)
		assert.Equal(t, "user_456", event.UserID)
		assert.Equal(t, "login", event.EventType)
		assert.Equal(t, "session_789", event.SessionID)
		assert.True(t, event.Success)
		assert.Equal(t, now, event.Timestamp)
	})

	t.Run("create failed auth event", func(t *testing.T) {
		event := &domain.AuthEvent{
			ID:           "event_456",
			UserID:       "user_789",
			EventType:    "failed_login",
			IPAddress:    "192.168.1.1",
			Success:      false,
			ErrorCode:    domain.AUTH_001,
			ErrorMessage: "Invalid credentials",
			Metadata:     map[string]interface{}{"email": "test@example.com"},
		}

		assert.False(t, event.Success)
		assert.Equal(t, domain.AUTH_001, event.ErrorCode)
		assert.Equal(t, "Invalid credentials", event.ErrorMessage)
		assert.Equal(t, "test@example.com", event.Metadata["email"])
	})
}

func TestSecurityEvent(t *testing.T) {
	t.Run("create security event", func(t *testing.T) {
		now := time.Now()
		event := &domain.SecurityEvent{
			ID:          "security_123",
			EventType:   "suspicious_login",
			UserID:      "user_456",
			IPAddress:   "192.168.1.1",
			Severity:    "high",
			Description: "Login from unusual location",
			Metadata:    map[string]interface{}{"country": "Unknown"},
			Timestamp:   now,
		}

		assert.Equal(t, "security_123", event.ID)
		assert.Equal(t, "suspicious_login", event.EventType)
		assert.Equal(t, "high", event.Severity)
		assert.Equal(t, "Login from unusual location", event.Description)
		assert.Equal(t, "Unknown", event.Metadata["country"])
	})
}

func TestLoginRequest(t *testing.T) {
	t.Run("valid login request", func(t *testing.T) {
		req := &domain.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "password123", req.Password)
	})
}

func TestRefreshRequest(t *testing.T) {
	t.Run("valid refresh request", func(t *testing.T) {
		req := &domain.RefreshRequest{
			RefreshToken: "refresh_token_123",
		}

		assert.Equal(t, "refresh_token_123", req.RefreshToken)
	})
}

func TestAuthContext(t *testing.T) {
	t.Run("create auth context", func(t *testing.T) {
		ctx := &domain.AuthContext{
			UserID:    "user_123",
			Email:     "test@example.com",
			Role:      "applicant",
			SessionID: "session_456",
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0...",
		}

		assert.Equal(t, "user_123", ctx.UserID)
		assert.Equal(t, "test@example.com", ctx.Email)
		assert.Equal(t, "applicant", ctx.Role)
		assert.Equal(t, "session_456", ctx.SessionID)
	})
}

// Benchmark tests
func BenchmarkUserRole_HasPermission(b *testing.B) {
	role := domain.RoleAdmin
	permission := domain.PermissionManageUsers

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		role.HasPermission(permission)
	}
}

func BenchmarkUserRole_GetPermissions(b *testing.B) {
	role := domain.RoleAdmin

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		role.GetPermissions()
	}
}
