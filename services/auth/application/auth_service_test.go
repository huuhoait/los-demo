package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/crypto/bcrypt"

	"github.com/huuhoait/los-demo/services/auth/domain"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Mock implementations
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) GenerateAccessToken(ctx context.Context, user *domain.User, sessionID string) (string, time.Time, error) {
	args := m.Called(ctx, user, sessionID)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockTokenManager) GenerateRefreshToken(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) ValidateAccessToken(ctx context.Context, token string) (*domain.JWTClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.JWTClaims), args.Error(1)
}

func (m *MockTokenManager) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenManager) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogAuthEvent(ctx context.Context, event *domain.AuthEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditLogger) LogSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Test setup helper
func setupAuthService(t *testing.T) (*AuthService, *MockUserRepository, *MockSessionRepository, *MockTokenManager, *MockCacheService, *MockAuditLogger) {
	userRepo := &MockUserRepository{}
	sessionRepo := &MockSessionRepository{}
	tokenManager := &MockTokenManager{}
	cache := &MockCacheService{}
	auditLogger := &MockAuditLogger{}
	logger := zaptest.NewLogger(t)
	
	bundle := i18n.NewBundle(language.English)
	localizer := i18n.NewLocalizer(bundle, language.English.String())

	authService := NewAuthService(
		userRepo,
		sessionRepo,
		tokenManager,
		cache,
		auditLogger,
		logger,
		localizer,
	)

	return authService, userRepo, sessionRepo, tokenManager, cache, auditLogger
}

func createTestUser() *domain.User {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	return &domain.User{
		ID:           "user_123",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
		FirstName:    "John",
		LastName:     "Doe",
		Role:         "applicant",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestAuthService_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		authService, userRepo, sessionRepo, tokenManager, cache, auditLogger := setupAuthService(t)
		
		user := createTestUser()
		ctx := context.Background()

		// Setup mocks
		cache.On("Exists", ctx, "lockout:user_123").Return(false, nil)
		userRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
		cache.On("Delete", ctx, "failed_attempts:user_123").Return(nil)
		sessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)
		tokenManager.On("GenerateRefreshToken", ctx).Return("refresh_token_123", nil)
		tokenManager.On("GenerateAccessToken", ctx, user, mock.AnythingOfType("string")).
			Return("access_token_123", time.Now().Add(15*time.Minute), nil)
		userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
		auditLogger.On("LogAuthEvent", ctx, mock.AnythingOfType("*domain.AuthEvent")).Return(nil)

		// Execute
		response, err := authService.Login(ctx, "test@example.com", "password123", "192.168.1.1", "Mozilla/5.0")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "access_token_123", response.AccessToken)
		assert.Equal(t, "refresh_token_123", response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, user, response.User)

		// Verify all mocks were called
		userRepo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		tokenManager.AssertExpectations(t)
		cache.AssertExpectations(t)
		auditLogger.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		authService, userRepo, _, _, cache, auditLogger := setupAuthService(t)
		
		user := createTestUser()
		ctx := context.Background()

		// Setup mocks
		cache.On("Exists", ctx, "lockout:user_123").Return(false, nil)
		userRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
		cache.On("Increment", ctx, "failed_attempts:user_123").Return(int64(1), nil)
		cache.On("SetExpiration", ctx, "failed_attempts:user_123", mock.AnythingOfType("time.Duration")).Return(nil)
		auditLogger.On("LogAuthEvent", ctx, mock.AnythingOfType("*domain.AuthEvent")).Return(nil)

		// Execute
		response, err := authService.Login(ctx, "test@example.com", "wrongpassword", "192.168.1.1", "Mozilla/5.0")

		// Assert
		require.Error(t, err)
		assert.Nil(t, response)
		
		authErr, ok := err.(*domain.AuthError)
		require.True(t, ok)
		assert.Equal(t, domain.AUTH_001, authErr.Code)

		userRepo.AssertExpectations(t)
		cache.AssertExpectations(t)
		auditLogger.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		authService, userRepo, _, _, cache, auditLogger := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		cache.On("Exists", ctx, mock.AnythingOfType("string")).Return(false, nil)
		userRepo.On("GetByEmail", ctx, "notfound@example.com").Return(nil, domain.NewAuthError(domain.AUTH_016, "User not found", "User not found"))
		auditLogger.On("LogAuthEvent", ctx, mock.AnythingOfType("*domain.AuthEvent")).Return(nil)

		// Execute
		response, err := authService.Login(ctx, "notfound@example.com", "password123", "192.168.1.1", "Mozilla/5.0")

		// Assert
		require.Error(t, err)
		assert.Nil(t, response)
		
		authErr, ok := err.(*domain.AuthError)
		require.True(t, ok)
		assert.Equal(t, domain.AUTH_001, authErr.Code)

		userRepo.AssertExpectations(t)
		auditLogger.AssertExpectations(t)
	})

	t.Run("account locked", func(t *testing.T) {
		authService, _, _, _, cache, auditLogger := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		cache.On("Exists", ctx, "lockout:user_123").Return(true, nil)
		auditLogger.On("LogAuthEvent", ctx, mock.AnythingOfType("*domain.AuthEvent")).Return(nil)

		// Execute
		response, err := authService.Login(ctx, "test@example.com", "password123", "192.168.1.1", "Mozilla/5.0")

		// Assert
		require.Error(t, err)
		assert.Nil(t, response)
		
		authErr, ok := err.(*domain.AuthError)
		require.True(t, ok)
		assert.Equal(t, domain.AUTH_002, authErr.Code)

		cache.AssertExpectations(t)
		auditLogger.AssertExpectations(t)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("successful token refresh", func(t *testing.T) {
		authService, userRepo, sessionRepo, tokenManager, _, auditLogger := setupAuthService(t)
		
		user := createTestUser()
		session := &domain.Session{
			ID:           "session_123",
			UserID:       user.ID,
			RefreshToken: "refresh_token_123",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}
		ctx := context.Background()

		// Setup mocks
		sessionRepo.On("GetByRefreshToken", ctx, "refresh_token_123").Return(session, nil)
		userRepo.On("GetByID", ctx, user.ID).Return(user, nil)
		tokenManager.On("GenerateRefreshToken", ctx).Return("new_refresh_token_456", nil)
		sessionRepo.On("Update", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)
		tokenManager.On("GenerateAccessToken", ctx, user, session.ID).
			Return("new_access_token_456", time.Now().Add(15*time.Minute), nil)
		auditLogger.On("LogAuthEvent", ctx, mock.AnythingOfType("*domain.AuthEvent")).Return(nil)

		// Execute
		response, err := authService.RefreshToken(ctx, "refresh_token_123", "192.168.1.1", "Mozilla/5.0")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "new_access_token_456", response.AccessToken)
		assert.Equal(t, "new_refresh_token_456", response.RefreshToken)
		assert.Equal(t, user, response.User)

		sessionRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		tokenManager.AssertExpectations(t)
		auditLogger.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		authService, _, sessionRepo, _, _, _ := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		sessionRepo.On("GetByRefreshToken", ctx, "invalid_token").Return(nil, domain.NewAuthError(domain.AUTH_007, "Invalid refresh token", "Invalid refresh token"))

		// Execute
		response, err := authService.RefreshToken(ctx, "invalid_token", "192.168.1.1", "Mozilla/5.0")

		// Assert
		require.Error(t, err)
		assert.Nil(t, response)
		
		authErr, ok := err.(*domain.AuthError)
		require.True(t, ok)
		assert.Equal(t, domain.AUTH_007, authErr.Code)

		sessionRepo.AssertExpectations(t)
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("successful logout", func(t *testing.T) {
		authService, _, sessionRepo, _, _, auditLogger := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		sessionRepo.On("Delete", ctx, "session_123").Return(nil)
		auditLogger.On("LogAuthEvent", ctx, mock.AnythingOfType("*domain.AuthEvent")).Return(nil)

		// Execute
		err := authService.Logout(ctx, "user_123", "session_123")

		// Assert
		require.NoError(t, err)

		sessionRepo.AssertExpectations(t)
		auditLogger.AssertExpectations(t)
	})

	t.Run("logout failure", func(t *testing.T) {
		authService, _, sessionRepo, _, _, _ := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		sessionRepo.On("Delete", ctx, "session_123").Return(domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to delete session"))

		// Execute
		err := authService.Logout(ctx, "user_123", "session_123")

		// Assert
		require.Error(t, err)
		
		authErr, ok := err.(*domain.AuthError)
		require.True(t, ok)
		assert.Equal(t, domain.AUTH_017, authErr.Code)

		sessionRepo.AssertExpectations(t)
	})
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		authService, _, sessionRepo, tokenManager, _, _ := setupAuthService(t)
		
		claims := &domain.JWTClaims{
			UserID:    "user_123",
			Email:     "test@example.com",
			Role:      "applicant",
			SessionID: "session_123",
		}
		session := &domain.Session{
			ID:        "session_123",
			UserID:    "user_123",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		ctx := context.Background()

		// Setup mocks
		tokenManager.On("ValidateAccessToken", ctx, "valid_token").Return(claims, nil)
		tokenManager.On("IsTokenRevoked", ctx, "valid_token").Return(false, nil)
		sessionRepo.On("GetByID", ctx, "session_123").Return(session, nil)

		// Execute
		authContext, err := authService.ValidateAccessToken(ctx, "valid_token")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, authContext)
		assert.Equal(t, "user_123", authContext.UserID)
		assert.Equal(t, "test@example.com", authContext.Email)
		assert.Equal(t, "applicant", authContext.Role)
		assert.Equal(t, "session_123", authContext.SessionID)

		tokenManager.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("invalid token", func(t *testing.T) {
		authService, _, _, tokenManager, _, _ := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		tokenManager.On("ValidateAccessToken", ctx, "invalid_token").Return(nil, domain.NewAuthError(domain.AUTH_004, "Invalid token", "Token is malformed"))

		// Execute
		authContext, err := authService.ValidateAccessToken(ctx, "invalid_token")

		// Assert
		require.Error(t, err)
		assert.Nil(t, authContext)
		
		authErr, ok := err.(*domain.AuthError)
		require.True(t, ok)
		assert.Equal(t, domain.AUTH_004, authErr.Code)

		tokenManager.AssertExpectations(t)
	})
}

func TestAuthService_CheckRateLimit(t *testing.T) {
	t.Run("within rate limit", func(t *testing.T) {
		authService, _, _, _, cache, _ := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		cache.On("Get", ctx, "rate_limit:auth:192.168.1.1").Return(int64(10), nil)

		// Execute
		err := authService.CheckRateLimit(ctx, "192.168.1.1")

		// Assert
		require.NoError(t, err)

		cache.AssertExpectations(t)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		authService, _, _, _, cache, _ := setupAuthService(t)
		
		ctx := context.Background()

		// Setup mocks
		cache.On("Get", ctx, "rate_limit:auth:192.168.1.1").Return(int64(100), nil)

		// Execute
		err := authService.CheckRateLimit(ctx, "192.168.1.1")

		// Assert
		require.Error(t, err)
		
		authErr, ok := err.(*domain.AuthError)
		require.True(t, ok)
		assert.Equal(t, domain.AUTH_010, authErr.Code)

		cache.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkAuthService_Login(b *testing.B) {
	authService, userRepo, sessionRepo, tokenManager, cache, auditLogger := setupAuthService(&testing.T{})
	
	user := createTestUser()
	ctx := context.Background()

	// Setup mocks
	cache.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	userRepo.On("GetByEmail", mock.Anything, mock.AnythingOfType("string")).Return(user, nil)
	cache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(nil)
	tokenManager.On("GenerateRefreshToken", mock.Anything).Return("refresh_token_123", nil)
	tokenManager.On("GenerateAccessToken", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return("access_token_123", time.Now().Add(15*time.Minute), nil)
	userRepo.On("UpdateLastLogin", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	auditLogger.On("LogAuthEvent", mock.Anything, mock.AnythingOfType("*domain.AuthEvent")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authService.Login(ctx, "test@example.com", "password123", "192.168.1.1", "Mozilla/5.0")
	}
}
