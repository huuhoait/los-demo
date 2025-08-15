package domain

import (
	"context"
	"io"
	"time"
)

// UserError represents a domain-specific error with localization support
type UserError struct {
	Code         string                 `json:"code"`
	Message      string                 `json:"message"`
	Field        string                 `json:"field,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	Cause        error                  `json:"-"`
}

// Error implements the error interface
func (e *UserError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// Unwrap returns the underlying cause
func (e *UserError) Unwrap() error {
	return e.Cause
}

// NewUserError creates a new UserError
func NewUserError(code, message string) *UserError {
	return &UserError{
		Code:    code,
		Message: message,
	}
}

// NewUserErrorWithField creates a new UserError with field information
func NewUserErrorWithField(code, message, field string) *UserError {
	return &UserError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// NewUserErrorWithData creates a new UserError with template data
func NewUserErrorWithData(code, message string, templateData map[string]interface{}) *UserError {
	return &UserError{
		Code:         code,
		Message:      message,
		TemplateData: templateData,
	}
}

// NewUserErrorWithCause creates a new UserError with an underlying cause
func NewUserErrorWithCause(code, message string, cause error) *UserError {
	return &UserError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error
	DeleteUser(ctx context.Context, userID string) error

	// User profile operations
	CreateProfile(ctx context.Context, profile *UserProfile) error
	GetProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error

	// User search and listing
	ListUsers(ctx context.Context, offset, limit int) ([]*User, error)
	SearchUsers(ctx context.Context, criteria map[string]interface{}) ([]*User, error)
}

// KYCRepository defines the interface for KYC operations
type KYCRepository interface {
	// KYC verification operations
	CreateKYCVerification(ctx context.Context, verification *KYCVerification) error
	GetKYCVerification(ctx context.Context, userID, verificationType string) (*KYCVerification, error)
	UpdateKYCVerification(ctx context.Context, verificationID string, updates map[string]interface{}) error
	ListKYCVerifications(ctx context.Context, userID string) ([]*KYCVerification, error)

	// KYC status tracking
	GetKYCStatus(ctx context.Context, userID string) (map[string]KYCStatus, error)
	UpdateKYCStatus(ctx context.Context, userID, verificationType string, status KYCStatus) error
}

// DocumentRepository defines the interface for document operations
type DocumentRepository interface {
	// Document CRUD operations
	CreateDocument(ctx context.Context, document *Document) error
	GetDocument(ctx context.Context, documentID string) (*Document, error)
	GetDocumentsByUserID(ctx context.Context, userID string) ([]*Document, error)
	GetDocumentsByType(ctx context.Context, userID, documentType string) ([]*Document, error)
	UpdateDocument(ctx context.Context, documentID string, updates map[string]interface{}) error
	DeleteDocument(ctx context.Context, documentID string) error
}

// DocumentStorageService defines the interface for file storage operations
type DocumentStorageService interface {
	// File operations
	UploadFile(ctx context.Context, key string, content io.Reader, contentType string, metadata map[string]string) error
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, key string) error
	GetFileMetadata(ctx context.Context, key string) (map[string]string, error)

	// Security operations
	GeneratePresignedURL(ctx context.Context, key string, expiration int) (string, error)
	EncryptContent(content []byte) ([]byte, string, error) // returns encrypted content and key
	DecryptContent(encryptedContent []byte, key string) ([]byte, error)
}

// EncryptionService defines the interface for data encryption
type EncryptionService interface {
	// Field-level encryption for PII
	EncryptField(plaintext string) (string, error)
	DecryptField(ciphertext string) (string, error)

	// File encryption
	EncryptFile(content []byte) ([]byte, string, error) // returns encrypted content and key
	DecryptFile(encryptedContent []byte, key string) ([]byte, error)

	// Key management
	GenerateKey() (string, error)
	RotateKey(oldKey string) (string, error)
}

// KYCProviderService defines the interface for external KYC providers
type KYCProviderService interface {
	// Identity verification
	InitiateIdentityVerification(ctx context.Context, userID string, personalInfo *UserProfile) (*KYCSession, error)
	GetVerificationStatus(ctx context.Context, providerReference string) (*KYCVerification, error)

	// Document verification
	VerifyDocument(ctx context.Context, documentType string, documentData []byte) (*KYCVerification, error)

	// Address verification
	VerifyAddress(ctx context.Context, address *Address) (*KYCVerification, error)

	// Provider capabilities
	GetSupportedDocumentTypes() []string
	GetProviderName() string
}

// NotificationService defines the interface for user notifications
type NotificationService interface {
	// Email notifications
	SendWelcomeEmail(ctx context.Context, userID, email, firstName string) error
	SendEmailVerification(ctx context.Context, userID, email, verificationCode string) error
	SendPasswordReset(ctx context.Context, userID, email, resetToken string) error

	// SMS notifications
	SendPhoneVerification(ctx context.Context, userID, phone, verificationCode string) error
	SendSecurityAlert(ctx context.Context, userID, phone, alertMessage string) error

	// Push notifications
	SendPushNotification(ctx context.Context, userID, title, message string, data map[string]interface{}) error
}

// ValidationService defines the interface for data validation
type ValidationService interface {
	// Personal information validation
	ValidateEmail(email string) error
	ValidatePhone(phone string) error
	ValidateSSN(ssn string) error
	ValidateDateOfBirth(dob time.Time) error
	ValidateAddress(address *Address) error

	// Document validation
	ValidateDocument(docType string, content []byte, mimeType string) error
	ValidateDocumentType(docType string) error
	ValidateFileSize(size int64) error
	ValidateMimeType(mimeType string) error
}

// AuditService defines the interface for audit logging
type AuditService interface {
	// User action logging
	LogUserCreated(ctx context.Context, userID, email string, metadata map[string]interface{}) error
	LogUserUpdated(ctx context.Context, userID string, changes map[string]interface{}) error
	LogProfileUpdated(ctx context.Context, userID string, changes map[string]interface{}) error
	LogDocumentUploaded(ctx context.Context, userID, documentID, documentType string) error
	LogKYCStatusChanged(ctx context.Context, userID, verificationType string, oldStatus, newStatus KYCStatus) error

	// Security events
	LogSecurityEvent(ctx context.Context, userID, eventType string, metadata map[string]interface{}) error
	LogDataAccess(ctx context.Context, userID, accessedBy, dataType string) error
}

// CacheService defines the interface for caching operations
type CacheService interface {
	// User data caching
	CacheUser(ctx context.Context, userID string, user *User, ttl int) error
	GetCachedUser(ctx context.Context, userID string) (*User, error)
	InvalidateUserCache(ctx context.Context, userID string) error

	// Profile caching
	CacheProfile(ctx context.Context, userID string, profile *UserProfile, ttl int) error
	GetCachedProfile(ctx context.Context, userID string) (*UserProfile, error)
	InvalidateProfileCache(ctx context.Context, userID string) error

	// KYC status caching
	CacheKYCStatus(ctx context.Context, userID string, status map[string]KYCStatus, ttl int) error
	GetCachedKYCStatus(ctx context.Context, userID string) (map[string]KYCStatus, error)
	InvalidateKYCStatus(ctx context.Context, userID string) error
}

// UserService defines the main business logic interface
type UserService interface {
	// User management
	CreateUser(ctx context.Context, request *CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, userID string, request *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, userID string) error

	// Profile management
	GetProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, request *UpdateProfileRequest) (*UserProfile, error)

	// Email and phone verification
	SendEmailVerification(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, userID, verificationCode string) error
	SendPhoneVerification(ctx context.Context, userID string) error
	VerifyPhone(ctx context.Context, userID, verificationCode string) error

	// KYC operations
	InitiateKYC(ctx context.Context, userID string) (*KYCSession, error)
	GetKYCStatus(ctx context.Context, userID string) (map[string]KYCStatus, error)
	UpdateKYCStatus(ctx context.Context, userID, verificationType string, status KYCStatus, data map[string]interface{}) error

	// Document management
	UploadDocument(ctx context.Context, userID string, document *DocumentUpload) (*Document, error)
	GetDocuments(ctx context.Context, userID string) ([]*Document, error)
	GetDocument(ctx context.Context, userID, documentID string) (*Document, error)
	DownloadDocument(ctx context.Context, userID, documentID string) (*DocumentStream, error)
	DeleteDocument(ctx context.Context, userID, documentID string) error

	// Search and listing
	SearchUsers(ctx context.Context, criteria map[string]interface{}, offset, limit int) ([]*User, error)
	ListUsers(ctx context.Context, offset, limit int) ([]*User, error)
}

// Error code constants for user service
const (
	// User validation errors
	USER_001 = "USER_001" // Invalid email format
	USER_002 = "USER_002" // Invalid phone format
	USER_003 = "USER_003" // Invalid SSN format
	USER_004 = "USER_004" // Invalid date of birth
	USER_005 = "USER_005" // Missing required field
	USER_006 = "USER_006" // Email already exists
	USER_007 = "USER_007" // Phone already exists
	USER_008 = "USER_008" // SSN already exists
	USER_009 = "USER_009" // User under minimum age
	USER_010 = "USER_010" // KYC already completed

	// Document errors
	USER_011 = "USER_011" // Invalid document format
	USER_012 = "USER_012" // File too large
	USER_013 = "USER_013" // Upload failed
	USER_014 = "USER_014" // Document not found
	USER_015 = "USER_015" // Encryption failed
	USER_016 = "USER_016" // S3 upload failed
	USER_017 = "USER_017" // Unsupported document type
	USER_018 = "USER_018" // Virus detected in file
	USER_019 = "USER_019" // Document expired
	USER_020 = "USER_020" // Document already exists

	// KYC errors
	USER_021 = "USER_021" // KYC provider error
	USER_022 = "USER_022" // KYC session expired
	USER_023 = "USER_023" // KYC verification failed
	USER_024 = "USER_024" // KYC manual review required
	USER_025 = "USER_025" // KYC provider unavailable

	// System errors
	USER_026 = "USER_026" // Database error
	USER_027 = "USER_027" // Cache error
	USER_028 = "USER_028" // Encryption error
	USER_029 = "USER_029" // Notification error
	USER_030 = "USER_030" // User not found
	USER_031 = "USER_031" // Profile not found
	USER_032 = "USER_032" // Unauthorized access
	USER_033 = "USER_033" // Rate limit exceeded
	USER_034 = "USER_034" // Service unavailable
	USER_035 = "USER_035" // Data integrity error
)
