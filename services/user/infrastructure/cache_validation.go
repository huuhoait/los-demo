package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/user/domain"
	"github.com/huuhoait/los-demo/services/shared/pkg/errors"
)

// RedisCacheService implements caching using Redis
type RedisCacheService struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisCacheService(client *redis.Client, logger *zap.Logger) domain.CacheService {
	return &RedisCacheService{
		client: client,
		logger: logger,
	}
}

func (r *RedisCacheService) CacheUser(ctx context.Context, userID string, user *domain.User, ttl int) error {
	key := fmt.Sprintf("user:%s", userID)

	data, err := json.Marshal(user)
	if err != nil {
		r.logger.Error("Failed to marshal user for cache", zap.Error(err))
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	err = r.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		r.logger.Error("Failed to cache user", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to cache user: %w", err)
	}

	r.logger.Debug("User cached successfully", zap.String("user_id", userID))
	return nil
}

func (r *RedisCacheService) GetCachedUser(ctx context.Context, userID string) (*domain.User, error) {
	key := fmt.Sprintf("user:%s", userID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.NewNotFoundError("user not in cache")
		}
		r.logger.Error("Failed to get cached user", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get cached user: %w", err)
	}

	var user domain.User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		r.logger.Error("Failed to unmarshal cached user", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cached user: %w", err)
	}

	return &user, nil
}

func (r *RedisCacheService) InvalidateUserCache(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s", userID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Failed to invalidate user cache", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}

	r.logger.Debug("User cache invalidated", zap.String("user_id", userID))
	return nil
}

func (r *RedisCacheService) CacheProfile(ctx context.Context, userID string, profile *domain.UserProfile, ttl int) error {
	key := fmt.Sprintf("profile:%s", userID)

	data, err := json.Marshal(profile)
	if err != nil {
		r.logger.Error("Failed to marshal profile for cache", zap.Error(err))
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	err = r.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		r.logger.Error("Failed to cache profile", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to cache profile: %w", err)
	}

	r.logger.Debug("Profile cached successfully", zap.String("user_id", userID))
	return nil
}

func (r *RedisCacheService) GetCachedProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	key := fmt.Sprintf("profile:%s", userID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.NewNotFoundError("profile not in cache")
		}
		r.logger.Error("Failed to get cached profile", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get cached profile: %w", err)
	}

	var profile domain.UserProfile
	err = json.Unmarshal([]byte(data), &profile)
	if err != nil {
		r.logger.Error("Failed to unmarshal cached profile", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cached profile: %w", err)
	}

	return &profile, nil
}

func (r *RedisCacheService) InvalidateProfileCache(ctx context.Context, userID string) error {
	key := fmt.Sprintf("profile:%s", userID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Failed to invalidate profile cache", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to invalidate profile cache: %w", err)
	}

	r.logger.Debug("Profile cache invalidated", zap.String("user_id", userID))
	return nil
}

func (r *RedisCacheService) CacheKYCStatus(ctx context.Context, userID string, status map[string]domain.KYCStatus, ttl int) error {
	key := fmt.Sprintf("kyc_status:%s", userID)

	data, err := json.Marshal(status)
	if err != nil {
		r.logger.Error("Failed to marshal KYC status for cache", zap.Error(err))
		return fmt.Errorf("failed to marshal KYC status: %w", err)
	}

	err = r.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		r.logger.Error("Failed to cache KYC status", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to cache KYC status: %w", err)
	}

	r.logger.Debug("KYC status cached successfully", zap.String("user_id", userID))
	return nil
}

func (r *RedisCacheService) GetCachedKYCStatus(ctx context.Context, userID string) (map[string]domain.KYCStatus, error) {
	key := fmt.Sprintf("kyc_status:%s", userID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.NewNotFoundError("KYC status not in cache")
		}
		r.logger.Error("Failed to get cached KYC status", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get cached KYC status: %w", err)
	}

	var status map[string]domain.KYCStatus
	err = json.Unmarshal([]byte(data), &status)
	if err != nil {
		r.logger.Error("Failed to unmarshal cached KYC status", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cached KYC status: %w", err)
	}

	return status, nil
}

func (r *RedisCacheService) InvalidateKYCStatus(ctx context.Context, userID string) error {
	key := fmt.Sprintf("kyc_status:%s", userID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Failed to invalidate KYC status cache", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to invalidate KYC status cache: %w", err)
	}

	r.logger.Debug("KYC status cache invalidated", zap.String("user_id", userID))
	return nil
}

// ValidationService implements data validation
type ValidationService struct {
	logger *zap.Logger
}

func NewValidationService(logger *zap.Logger) domain.ValidationService {
	return &ValidationService{
		logger: logger,
	}
}

func (v *ValidationService) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// RFC 5322 compliant email regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	if len(email) > 254 {
		return fmt.Errorf("email too long")
	}

	return nil
}

func (v *ValidationService) ValidatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone is required")
	}

	// Remove all non-digit characters
	cleanPhone := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// US phone number validation (10 or 11 digits)
	if len(cleanPhone) == 11 && cleanPhone[0] == '1' {
		cleanPhone = cleanPhone[1:] // Remove country code
	}

	if len(cleanPhone) != 10 {
		return fmt.Errorf("invalid phone number format")
	}

	// Check for valid area code (first digit 2-9)
	if cleanPhone[0] < '2' || cleanPhone[0] > '9' {
		return fmt.Errorf("invalid area code")
	}

	// Check for valid exchange code (fourth digit 2-9)
	if cleanPhone[3] < '2' || cleanPhone[3] > '9' {
		return fmt.Errorf("invalid exchange code")
	}

	return nil
}

func (v *ValidationService) ValidateSSN(ssn string) error {
	if ssn == "" {
		return fmt.Errorf("SSN is required")
	}

	// Remove all non-digit characters
	cleanSSN := regexp.MustCompile(`[^\d]`).ReplaceAllString(ssn, "")

	if len(cleanSSN) != 9 {
		return fmt.Errorf("SSN must be 9 digits")
	}

	// Check for invalid SSN patterns
	invalidSSNs := []string{
		"000000000", "111111111", "222222222", "333333333", "444444444",
		"555555555", "666666666", "777777777", "888888888", "999999999",
		"123456789", "987654321",
	}

	for _, invalid := range invalidSSNs {
		if cleanSSN == invalid {
			return fmt.Errorf("invalid SSN pattern")
		}
	}

	// Check area number (first 3 digits) - cannot be 000, 666, or 900-999
	areaNum, _ := strconv.Atoi(cleanSSN[:3])
	if areaNum == 0 || areaNum == 666 || areaNum >= 900 {
		return fmt.Errorf("invalid SSN area number")
	}

	// Check group number (middle 2 digits) - cannot be 00
	groupNum, _ := strconv.Atoi(cleanSSN[3:5])
	if groupNum == 0 {
		return fmt.Errorf("invalid SSN group number")
	}

	// Check serial number (last 4 digits) - cannot be 0000
	serialNum, _ := strconv.Atoi(cleanSSN[5:9])
	if serialNum == 0 {
		return fmt.Errorf("invalid SSN serial number")
	}

	return nil
}

func (v *ValidationService) ValidateDateOfBirth(dob time.Time) error {
	if dob.IsZero() {
		return fmt.Errorf("date of birth is required")
	}

	now := time.Now()

	// Check if date is in the future
	if dob.After(now) {
		return fmt.Errorf("date of birth cannot be in the future")
	}

	// Check minimum age (18 years)
	minAge := now.AddDate(-18, 0, 0)
	if dob.After(minAge) {
		return fmt.Errorf("must be at least 18 years old")
	}

	// Check maximum age (120 years)
	maxAge := now.AddDate(-120, 0, 0)
	if dob.Before(maxAge) {
		return fmt.Errorf("invalid date of birth - too old")
	}

	return nil
}

func (v *ValidationService) ValidateAddress(address *domain.Address) error {
	if address == nil {
		return fmt.Errorf("address is required")
	}

	if address.Street == "" {
		return fmt.Errorf("street address is required")
	}

	if len(address.Street) > 100 {
		return fmt.Errorf("street address too long")
	}

	if address.City == "" {
		return fmt.Errorf("city is required")
	}

	if len(address.City) > 50 {
		return fmt.Errorf("city name too long")
	}

	if address.State == "" {
		return fmt.Errorf("state is required")
	}

	// Validate US state code (2 letters)
	if len(address.State) != 2 {
		return fmt.Errorf("state must be 2-letter code")
	}

	if address.ZipCode == "" {
		return fmt.Errorf("zip code is required")
	}

	// Validate US zip code (5 digits or 5+4 format)
	zipRegex := regexp.MustCompile(`^\d{5}(-\d{4})?$`)
	if !zipRegex.MatchString(address.ZipCode) {
		return fmt.Errorf("invalid zip code format")
	}

	if address.Country == "" {
		address.Country = "US" // Default to US
	}

	return nil
}

func (v *ValidationService) ValidateDocument(docType string, content []byte, mimeType string) error {
	if err := v.ValidateDocumentType(docType); err != nil {
		return err
	}

	if err := v.ValidateFileSize(int64(len(content))); err != nil {
		return err
	}

	if err := v.ValidateMimeType(mimeType); err != nil {
		return err
	}

	// Additional document-specific validation could go here
	// For example, checking if the content actually matches the claimed mime type

	return nil
}

func (v *ValidationService) ValidateDocumentType(docType string) error {
	validTypes := []string{
		domain.DocumentTypeDriversLicense,
		domain.DocumentTypePassport,
		domain.DocumentTypePayStub,
		domain.DocumentTypeBankStatement,
		domain.DocumentTypeUtilityBill,
		domain.DocumentTypeW2,
		domain.DocumentType1099,
	}

	for _, validType := range validTypes {
		if docType == validType {
			return nil
		}
	}

	return fmt.Errorf("unsupported document type: %s", docType)
}

func (v *ValidationService) ValidateFileSize(size int64) error {
	const maxFileSize = 10 * 1024 * 1024 // 10 MB

	if size <= 0 {
		return fmt.Errorf("file is empty")
	}

	if size > maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	return nil
}

func (v *ValidationService) ValidateMimeType(mimeType string) error {
	allowedTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/tiff",
		"image/bmp",
	}

	mimeType = strings.ToLower(mimeType)

	for _, allowedType := range allowedTypes {
		if mimeType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("unsupported file type: %s", mimeType)
}
