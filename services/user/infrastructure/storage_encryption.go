package infrastructure

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"

	"our-los/services/user/domain"
)

// S3DocumentStorageService implements document storage using AWS S3
type S3DocumentStorageService struct {
	s3Client  S3Client
	bucket    string
	kmsKeyID  string
	logger    *zap.Logger
}

// S3Client interface for AWS S3 operations
type S3Client interface {
	UploadFile(ctx context.Context, bucket, key string, content io.Reader, contentType string, metadata map[string]string) error
	DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, bucket, key string) error
	GetFileMetadata(ctx context.Context, bucket, key string) (map[string]string, error)
	GeneratePresignedURL(ctx context.Context, bucket, key string, expiration int) (string, error)
}

func NewS3DocumentStorageService(s3Client S3Client, bucket, kmsKeyID string, logger *zap.Logger) domain.DocumentStorageService {
	return &S3DocumentStorageService{
		s3Client:  s3Client,
		bucket:    bucket,
		kmsKeyID:  kmsKeyID,
		logger:    logger,
	}
}

func (s *S3DocumentStorageService) UploadFile(ctx context.Context, key string, content io.Reader, contentType string, metadata map[string]string) error {
	logger := s.logger.With(
		zap.String("operation", "upload_file"),
		zap.String("bucket", s.bucket),
		zap.String("key", key),
	)

	logger.Info("Starting file upload to S3")

	// Add encryption metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["encryption"] = "server-side"
	metadata["kms-key-id"] = s.kmsKeyID

	err := s.s3Client.UploadFile(ctx, s.bucket, key, content, contentType, metadata)
	if err != nil {
		logger.Error("Failed to upload file to S3", zap.Error(err))
		return fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Info("File uploaded successfully to S3")
	return nil
}

func (s *S3DocumentStorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	logger := s.logger.With(
		zap.String("operation", "download_file"),
		zap.String("bucket", s.bucket),
		zap.String("key", key),
	)

	logger.Info("Starting file download from S3")

	reader, err := s.s3Client.DownloadFile(ctx, s.bucket, key)
	if err != nil {
		logger.Error("Failed to download file from S3", zap.Error(err))
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	logger.Info("File downloaded successfully from S3")
	return reader, nil
}

func (s *S3DocumentStorageService) DeleteFile(ctx context.Context, key string) error {
	logger := s.logger.With(
		zap.String("operation", "delete_file"),
		zap.String("bucket", s.bucket),
		zap.String("key", key),
	)

	logger.Info("Starting file deletion from S3")

	err := s.s3Client.DeleteFile(ctx, s.bucket, key)
	if err != nil {
		logger.Error("Failed to delete file from S3", zap.Error(err))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted successfully from S3")
	return nil
}

func (s *S3DocumentStorageService) GetFileMetadata(ctx context.Context, key string) (map[string]string, error) {
	logger := s.logger.With(
		zap.String("operation", "get_file_metadata"),
		zap.String("bucket", s.bucket),
		zap.String("key", key),
	)

	metadata, err := s.s3Client.GetFileMetadata(ctx, s.bucket, key)
	if err != nil {
		logger.Error("Failed to get file metadata from S3", zap.Error(err))
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return metadata, nil
}

func (s *S3DocumentStorageService) GeneratePresignedURL(ctx context.Context, key string, expiration int) (string, error) {
	logger := s.logger.With(
		zap.String("operation", "generate_presigned_url"),
		zap.String("bucket", s.bucket),
		zap.String("key", key),
		zap.Int("expiration", expiration),
	)

	url, err := s.s3Client.GeneratePresignedURL(ctx, s.bucket, key, expiration)
	if err != nil {
		logger.Error("Failed to generate presigned URL", zap.Error(err))
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	logger.Info("Presigned URL generated successfully")
	return url, nil
}

func (s *S3DocumentStorageService) EncryptContent(content []byte) ([]byte, string, error) {
	// Generate a random AES-256 key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, "", fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the content
	ciphertext := gcm.Seal(nonce, nonce, content, nil)

	// Encode the key as base64 for storage
	keyStr := base64.StdEncoding.EncodeToString(key)

	return ciphertext, keyStr, nil
}

func (s *S3DocumentStorageService) DecryptContent(encryptedContent []byte, keyStr string) ([]byte, error) {
	// Decode the key from base64
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract the nonce from the beginning of the ciphertext
	nonceSize := gcm.NonceSize()
	if len(encryptedContent) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedContent[:nonceSize], encryptedContent[nonceSize:]

	// Decrypt the content
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content: %w", err)
	}

	return plaintext, nil
}

// AESEncryptionService implements field-level encryption using AES
type AESEncryptionService struct {
	masterKey []byte
	logger    *zap.Logger
}

func NewAESEncryptionService(masterKeyString string, logger *zap.Logger) domain.EncryptionService {
	// Create a 32-byte key from the master key string using SHA-256
	hash := sha256.Sum256([]byte(masterKeyString))
	return &AESEncryptionService{
		masterKey: hash[:],
		logger:    logger,
	}
}

func (e *AESEncryptionService) EncryptField(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode as base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *AESEncryptionService) DecryptField(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract the nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt field: %w", err)
	}

	return string(plaintext), nil
}

func (e *AESEncryptionService) EncryptFile(content []byte) ([]byte, string, error) {
	// Generate a random AES-256 key for this file
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, "", fmt.Errorf("failed to generate file encryption key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the content
	ciphertext := gcm.Seal(nonce, nonce, content, nil)

	// Encrypt the file key with the master key for storage
	encryptedKey, err := e.EncryptField(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt file key: %w", err)
	}

	return ciphertext, encryptedKey, nil
}

func (e *AESEncryptionService) DecryptFile(encryptedContent []byte, encryptedKey string) ([]byte, error) {
	// Decrypt the file key
	keyStr, err := e.DecryptField(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file key: %w", err)
	}

	// Decode the key
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(encryptedContent) < nonceSize {
		return nil, fmt.Errorf("encrypted content too short")
	}

	nonce, ciphertext := encryptedContent[:nonceSize], encryptedContent[nonceSize:]

	// Decrypt the content
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file content: %w", err)
	}

	return plaintext, nil
}

func (e *AESEncryptionService) GenerateKey() (string, error) {
	// Generate a random 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

func (e *AESEncryptionService) RotateKey(oldKey string) (string, error) {
	// Simply generate a new key
	return e.GenerateKey()
}

// MockKYCProviderService implements a mock KYC provider for testing/development
type MockKYCProviderService struct {
	providerName string
	logger       *zap.Logger
}

func NewMockKYCProviderService(logger *zap.Logger) domain.KYCProviderService {
	return &MockKYCProviderService{
		providerName: "mock-kyc-provider",
		logger:       logger,
	}
}

func (m *MockKYCProviderService) InitiateIdentityVerification(ctx context.Context, userID string, personalInfo *domain.UserProfile) (*domain.KYCSession, error) {
	logger := m.logger.With(
		zap.String("operation", "initiate_identity_verification"),
		zap.String("user_id", userID),
		zap.String("provider", m.providerName),
	)

	logger.Info("Starting mock KYC identity verification")

	// Generate a mock session
	session := &domain.KYCSession{
		ID:                fmt.Sprintf("kyc_session_%s", userID),
		UserID:            userID,
		Provider:          m.providerName,
		SessionURL:        fmt.Sprintf("https://mock-kyc.example.com/verify/%s", userID),
		ProviderReference: fmt.Sprintf("mock_ref_%d", time.Now().Unix()),
		ExpiresAt:         time.Now().Add(24 * time.Hour),
		Status:            "pending",
		Metadata: map[string]interface{}{
			"verification_type": "identity",
			"initiated_at":      time.Now(),
		},
		CreatedAt: time.Now(),
	}

	logger.Info("Mock KYC session created successfully", zap.String("session_id", session.ID))
	return session, nil
}

func (m *MockKYCProviderService) GetVerificationStatus(ctx context.Context, providerReference string) (*domain.KYCVerification, error) {
	logger := m.logger.With(
		zap.String("operation", "get_verification_status"),
		zap.String("provider_reference", providerReference),
	)

	// Mock verification result
	verification := &domain.KYCVerification{
		ID:                providerReference,
		VerificationType:  "identity",
		Provider:          m.providerName,
		Status:            domain.KYCStatusVerified,
		ProviderReference: providerReference,
		VerificationData: map[string]interface{}{
			"confidence_score": 0.95,
			"verified_fields":  []string{"name", "date_of_birth", "address"},
			"verification_time": time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	logger.Info("Mock verification status retrieved", zap.String("status", string(verification.Status)))
	return verification, nil
}

func (m *MockKYCProviderService) VerifyDocument(ctx context.Context, documentType string, documentData []byte) (*domain.KYCVerification, error) {
	logger := m.logger.With(
		zap.String("operation", "verify_document"),
		zap.String("document_type", documentType),
		zap.Int("document_size", len(documentData)),
	)

	// Mock document verification
	verification := &domain.KYCVerification{
		ID:                fmt.Sprintf("doc_verify_%d", time.Now().Unix()),
		VerificationType:  "document",
		Provider:          m.providerName,
		Status:            domain.KYCStatusVerified,
		VerificationData: map[string]interface{}{
			"document_type":    documentType,
			"authenticity":     "verified",
			"confidence_score": 0.92,
			"extracted_data": map[string]interface{}{
				"document_number": "MOCK123456789",
				"expiry_date":     "2030-12-31",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	logger.Info("Mock document verification completed", zap.String("status", string(verification.Status)))
	return verification, nil
}

func (m *MockKYCProviderService) VerifyAddress(ctx context.Context, address *domain.Address) (*domain.KYCVerification, error) {
	logger := m.logger.With(
		zap.String("operation", "verify_address"),
		zap.String("city", address.City),
		zap.String("state", address.State),
	)

	// Mock address verification
	verification := &domain.KYCVerification{
		ID:                fmt.Sprintf("addr_verify_%d", time.Now().Unix()),
		VerificationType:  "address",
		Provider:          m.providerName,
		Status:            domain.KYCStatusVerified,
		VerificationData: map[string]interface{}{
			"address_match":    "exact",
			"deliverable":      true,
			"confidence_score": 0.98,
			"standardized_address": map[string]interface{}{
				"street":   address.Street,
				"city":     address.City,
				"state":    address.State,
				"zip_code": address.ZipCode,
				"country":  address.Country,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	logger.Info("Mock address verification completed", zap.String("status", string(verification.Status)))
	return verification, nil
}

func (m *MockKYCProviderService) GetSupportedDocumentTypes() []string {
	return []string{
		domain.DocumentTypeDriversLicense,
		domain.DocumentTypePassport,
		domain.DocumentTypePayStub,
		domain.DocumentTypeBankStatement,
		domain.DocumentTypeUtilityBill,
		domain.DocumentTypeW2,
		domain.DocumentType1099,
	}
}

func (m *MockKYCProviderService) GetProviderName() string {
	return m.providerName
}
