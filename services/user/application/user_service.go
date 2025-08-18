package application

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/huuhoait/los-demo/services/user/domain"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
)

type UserServiceImpl struct {
	userRepo            domain.UserRepository
	kycRepo             domain.KYCRepository
	documentRepo        domain.DocumentRepository
	storageService      domain.DocumentStorageService
	encryptionService   domain.EncryptionService
	kycProvider         domain.KYCProviderService
	notificationService domain.NotificationService
	validationService   domain.ValidationService
	auditService        domain.AuditService
	cacheService        domain.CacheService
	logger              *zap.Logger
	localizer           *i18n.Localizer
}

func NewUserService(
	userRepo domain.UserRepository,
	kycRepo domain.KYCRepository,
	documentRepo domain.DocumentRepository,
	storageService domain.DocumentStorageService,
	encryptionService domain.EncryptionService,
	kycProvider domain.KYCProviderService,
	notificationService domain.NotificationService,
	validationService domain.ValidationService,
	auditService domain.AuditService,
	cacheService domain.CacheService,
	logger *zap.Logger,
	localizer *i18n.Localizer,
) domain.UserService {
	return &UserServiceImpl{
		userRepo:            userRepo,
		kycRepo:             kycRepo,
		documentRepo:        documentRepo,
		storageService:      storageService,
		encryptionService:   encryptionService,
		kycProvider:         kycProvider,
		notificationService: notificationService,
		validationService:   validationService,
		auditService:        auditService,
		cacheService:        cacheService,
		logger:              logger,
		localizer:           localizer,
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, request *domain.CreateUserRequest) (*domain.User, error) {
	logger := s.logger.With(
		zap.String("operation", "create_user"),
		zap.String("email", request.Email),
	)

	logger.Info("Starting user creation")

	// Validate input
	if err := s.validateCreateUserRequest(request); err != nil {
		logger.Error("User creation validation failed", zap.Error(err))
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, request.Email)
	if err != nil && err.Error() != "not found" {
		logger.Error("Failed to check existing user", zap.Error(err))
		return nil, &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	if existingUser != nil {
		logger.Warn("User already exists with email", zap.String("email", request.Email))
		return nil, &domain.UserError{
			Code:        domain.USER_006,
			Message:     s.localizer.Localize(context.Background(), domain.USER_006, nil),
			Field:       "email",
		}
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, &domain.UserError{
			Code:        domain.USER_028,
			Message:     s.localizer.Localize(context.Background(), domain.USER_028, nil),
		}
	}

	// Create user entity
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        strings.ToLower(strings.TrimSpace(request.Email)),
		PasswordHash: string(passwordHash),
		Phone:        request.Phone,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create user in database
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		logger.Error("Failed to create user in database", zap.Error(err))
		return nil, &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Create initial user profile
	profile := &domain.UserProfile{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Phone:     request.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateProfile(ctx, profile); err != nil {
		logger.Error("Failed to create user profile", zap.Error(err))
		// Note: User is already created, this is a partial failure
		// In production, you might want to implement compensating transactions
	}

	// Send welcome email
	if err := s.notificationService.SendWelcomeEmail(ctx, user.ID, user.Email, request.FirstName); err != nil {
		logger.Warn("Failed to send welcome email", zap.Error(err))
		// Non-critical error, don't fail the user creation
	}

	// Cache the user
	if err := s.cacheService.CacheUser(ctx, user.ID, user, 3600); err != nil {
		logger.Warn("Failed to cache user", zap.Error(err))
	}

	// Log audit event
	if err := s.auditService.LogUserCreated(ctx, user.ID, user.Email, map[string]interface{}{
		"phone":      user.Phone,
		"first_name": request.FirstName,
		"last_name":  request.LastName,
	}); err != nil {
		logger.Warn("Failed to log audit event", zap.Error(err))
	}

	logger.Info("User created successfully", zap.String("user_id", user.ID))

	// Remove password hash from response
	user.PasswordHash = ""
	return user, nil
}

func (s *UserServiceImpl) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	logger := s.logger.With(
		zap.String("operation", "get_user"),
		zap.String("user_id", userID),
	)

	// Try cache first
	if cachedUser, err := s.cacheService.GetCachedUser(ctx, userID); err == nil && cachedUser != nil {
		logger.Debug("User found in cache")
		return cachedUser, nil
	}

	// Get from database
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err != nil && err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:        domain.USER_030,
				Message:     s.localizer.Localize(context.Background(), domain.USER_030, nil),
			}
		}
		logger.Error("Failed to get user from database", zap.Error(err))
		return nil, &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Cache the user
	if err := s.cacheService.CacheUser(ctx, userID, user, 3600); err != nil {
		logger.Warn("Failed to cache user", zap.Error(err))
	}

	// Remove password hash from response
	user.PasswordHash = ""
	return user, nil
}

func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	logger := s.logger.With(
		zap.String("operation", "get_user_by_email"),
		zap.String("email", email),
	)

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if err != nil && err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:        domain.USER_030,
				Message:     s.localizer.Localize(context.Background(), domain.USER_030, nil),
			}
		}
		logger.Error("Failed to get user by email", zap.Error(err))
		return nil, &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Remove password hash from response
	user.PasswordHash = ""
	return user, nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, userID string, request *domain.UpdateUserRequest) (*domain.User, error) {
	logger := s.logger.With(
		zap.String("operation", "update_user"),
		zap.String("user_id", userID),
	)

	// Validate input
	if err := s.validateUpdateUserRequest(request); err != nil {
		return nil, err
	}

	// Get existing user
	existingUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err != nil && err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:        domain.USER_030,
				Message:     s.localizer.Localize(context.Background(), domain.USER_030, nil),
			}
		}
		return nil, &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Build updates map
	updates := make(map[string]interface{})
	changes := make(map[string]interface{})

	if request.Phone != nil && *request.Phone != existingUser.Phone {
		updates["phone"] = *request.Phone
		updates["phone_verified"] = false
		changes["phone"] = map[string]interface{}{
			"old": existingUser.Phone,
			"new": *request.Phone,
		}
	}

	updates["updated_at"] = time.Now()

	// Apply updates
	if len(updates) > 1 { // More than just updated_at
		if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
			logger.Error("Failed to update user", zap.Error(err))
			return nil, &domain.UserError{
				Code:        domain.USER_026,
				Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
			}
		}

		// Invalidate cache
		if err := s.cacheService.InvalidateUserCache(ctx, userID); err != nil {
			logger.Warn("Failed to invalidate user cache", zap.Error(err))
		}

		// Log audit event
		if err := s.auditService.LogUserUpdated(ctx, userID, changes); err != nil {
			logger.Warn("Failed to log audit event", zap.Error(err))
		}

		logger.Info("User updated successfully")
	}

	// Get updated user
	return s.GetUser(ctx, userID)
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	logger := s.logger.With(
		zap.String("operation", "delete_user"),
		zap.String("user_id", userID),
	)

	// Check if user exists
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err != nil && err.Error() == "not found" {
			return &domain.UserError{
				Code:        domain.USER_030,
				Message:     s.localizer.Localize(context.Background(), domain.USER_030, nil),
			}
		}
		return &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Soft delete user
	updates := map[string]interface{}{
		"status":     "deleted",
		"updated_at": time.Now(),
	}

	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Invalidate cache
	if err := s.cacheService.InvalidateUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to invalidate user cache", zap.Error(err))
	}

	logger.Info("User deleted successfully")
	return nil
}

func (s *UserServiceImpl) GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	logger := s.logger.With(
		zap.String("operation", "get_profile"),
		zap.String("user_id", userID),
	)

	// Try cache first
	if cachedProfile, err := s.cacheService.GetCachedProfile(ctx, userID); err == nil && cachedProfile != nil {
		logger.Debug("Profile found in cache")
		return cachedProfile, nil
	}

	// Get from database
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		if err != nil && err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:        domain.USER_031,
				Message:     s.localizer.Localize(context.Background(), domain.USER_031, nil),
			}
		}
		logger.Error("Failed to get profile from database", zap.Error(err))
		return nil, &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Decrypt SSN if present
	if profile.SSNEncrypted != "" {
		decryptedSSN, err := s.encryptionService.DecryptField(profile.SSNEncrypted)
		if err != nil {
			logger.Error("Failed to decrypt SSN", zap.Error(err))
			// Don't return error, just log it
		} else {
			// Mask SSN for security (show only last 4 digits)
			if len(decryptedSSN) >= 4 {
				profile.SSNEncrypted = "***-**-" + decryptedSSN[len(decryptedSSN)-4:]
			}
		}
	}

	// Cache the profile
	if err := s.cacheService.CacheProfile(ctx, userID, profile, 3600); err != nil {
		logger.Warn("Failed to cache profile", zap.Error(err))
	}

	return profile, nil
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, userID string, request *domain.UpdateProfileRequest) (*domain.UserProfile, error) {
	logger := s.logger.With(
		zap.String("operation", "update_profile"),
		zap.String("user_id", userID),
	)

	// Validate input
	if err := s.validateUpdateProfileRequest(request); err != nil {
		return nil, err
	}

	// Get existing profile
	existingProfile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		if err != nil && err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:        domain.USER_031,
				Message:     s.localizer.Localize(context.Background(), domain.USER_031, nil),
			}
		}
		return nil, &domain.UserError{
			Code:        domain.USER_026,
			Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
		}
	}

	// Build updates map
	updates := make(map[string]interface{})
	changes := make(map[string]interface{})

	if request.DateOfBirth != nil && !request.DateOfBirth.Equal(existingProfile.DateOfBirth) {
		updates["date_of_birth"] = *request.DateOfBirth
		changes["date_of_birth"] = map[string]interface{}{
			"old": existingProfile.DateOfBirth,
			"new": *request.DateOfBirth,
		}
	}

	if request.Phone != nil && *request.Phone != existingProfile.Phone {
		updates["phone"] = *request.Phone
		changes["phone"] = map[string]interface{}{
			"old": existingProfile.Phone,
			"new": *request.Phone,
		}
	}

	if request.Address != nil {
		updates["address"] = *request.Address
		changes["address"] = map[string]interface{}{
			"old": existingProfile.Address,
			"new": *request.Address,
		}
	}

	if request.EmploymentInfo != nil {
		updates["employment_info"] = *request.EmploymentInfo
		changes["employment_info"] = "updated"
	}

	if request.FinancialInfo != nil {
		updates["financial_info"] = *request.FinancialInfo
		changes["financial_info"] = "updated"
	}

	updates["updated_at"] = time.Now()

	// Apply updates
	if len(updates) > 1 { // More than just updated_at
		if err := s.userRepo.UpdateProfile(ctx, userID, updates); err != nil {
			logger.Error("Failed to update profile", zap.Error(err))
			return nil, &domain.UserError{
				Code:        domain.USER_026,
				Message:     s.localizer.Localize(context.Background(), domain.USER_026, nil),
			}
		}

		// Invalidate cache
		if err := s.cacheService.InvalidateProfileCache(ctx, userID); err != nil {
			logger.Warn("Failed to invalidate profile cache", zap.Error(err))
		}

		// Log audit event
		if err := s.auditService.LogProfileUpdated(ctx, userID, changes); err != nil {
			logger.Warn("Failed to log audit event", zap.Error(err))
		}

		logger.Info("Profile updated successfully")
	}

	// Get updated profile
	return s.GetProfile(ctx, userID)
}

// Validation methods
func (s *UserServiceImpl) validateCreateUserRequest(request *domain.CreateUserRequest) error {
	if request.Email == "" {
		return &domain.UserError{
			Code:        domain.USER_005,
			Message:     s.localizer.Localize(context.Background(), domain.USER_005, nil),
			Field:       "email",
		}
	}

	if err := s.validationService.ValidateEmail(request.Email); err != nil {
		return &domain.UserError{
			Code:        domain.USER_001,
			Message:     s.localizer.Localize(context.Background(), domain.USER_001, nil),
			Field:       "email",
		}
	}

	if request.Password == "" {
		return &domain.UserError{
			Code:        domain.USER_005,
			Message:     s.localizer.Localize(context.Background(), domain.USER_005, nil),
			Field:       "password",
		}
	}

	if len(request.Password) < 8 {
		return &domain.UserError{
			Code:        domain.USER_005,
			Message:     s.localizer.Localize(context.Background(), domain.USER_005, nil),
			Field:       "password",
		}
	}

	if request.Phone != "" {
		if err := s.validationService.ValidatePhone(request.Phone); err != nil {
			return &domain.UserError{
				Code:        domain.USER_002,
				Message:     s.localizer.Localize(context.Background(), domain.USER_002, nil),
				Field:       "phone",
			}
		}
	}

	if request.FirstName == "" {
		return &domain.UserError{
			Code:        domain.USER_005,
			Message:     s.localizer.Localize(context.Background(), domain.USER_005, nil),
			Field:       "first_name",
		}
	}

	if request.LastName == "" {
		return &domain.UserError{
			Code:        domain.USER_005,
			Message:     s.localizer.Localize(context.Background(), domain.USER_005, nil),
			Field:       "last_name",
		}
	}

	return nil
}

func (s *UserServiceImpl) validateUpdateUserRequest(request *domain.UpdateUserRequest) error {
	if request.Phone != nil && *request.Phone != "" {
		if err := s.validationService.ValidatePhone(*request.Phone); err != nil {
			return &domain.UserError{
				Code:        domain.USER_002,
				Message:     s.localizer.Localize(context.Background(), domain.USER_002, nil),
				Field:       "phone",
			}
		}
	}

	return nil
}

func (s *UserServiceImpl) validateUpdateProfileRequest(request *domain.UpdateProfileRequest) error {
	if request.DateOfBirth != nil {
		if err := s.validationService.ValidateDateOfBirth(*request.DateOfBirth); err != nil {
			return &domain.UserError{
				Code:        domain.USER_004,
				Message:     s.localizer.Localize(context.Background(), domain.USER_004, nil),
				Field:       "date_of_birth",
			}
		}
	}

	if request.Phone != nil && *request.Phone != "" {
		if err := s.validationService.ValidatePhone(*request.Phone); err != nil {
			return &domain.UserError{
				Code:        domain.USER_002,
				Message:     s.localizer.Localize(context.Background(), domain.USER_002, nil),
				Field:       "phone",
			}
		}
	}

	if request.Address != nil {
		if err := s.validationService.ValidateAddress(request.Address); err != nil {
			return &domain.UserError{
				Code:        domain.USER_005,
				Message:     s.localizer.Localize(context.Background(), domain.USER_005, nil),
				Field:       "address",
			}
		}
	}

	return nil
}

// Additional validation helper methods
func (s *UserServiceImpl) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (s *UserServiceImpl) isValidPhone(phone string) bool {
	// Simple US phone number validation
	phoneRegex := regexp.MustCompile(`^\+?1?[2-9]\d{2}[2-9]\d{2}\d{4}$`)
	cleanPhone := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	return phoneRegex.MatchString(cleanPhone)
}
