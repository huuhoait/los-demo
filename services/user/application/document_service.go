package application

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/user/domain"
)

// Document management methods for UserServiceImpl

func (s *UserServiceImpl) UploadDocument(ctx context.Context, userID string, document *domain.DocumentUpload) (*domain.Document, error) {
	logger := s.logger.With(
		zap.String("operation", "upload_document"),
		zap.String("user_id", userID),
		zap.String("document_type", document.Type),
	)

	logger.Info("Starting document upload")

	// Validate document
	if err := s.validateDocumentUpload(document); err != nil {
		logger.Error("Document validation failed", zap.Error(err))
		return nil, err
	}

	// Check if user exists
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:    domain.USER_030,
				Message: s.localizer.Localize(ctx, domain.USER_030, nil),
			}
		}
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Check for existing document of same type
	existingDocs, err := s.documentRepo.GetDocumentsByType(ctx, userID, document.Type)
	if err != nil && err.Error() != "not found" {
		logger.Error("Failed to check existing documents", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	if len(existingDocs) > 0 {
		logger.Warn("Document of this type already exists",
			zap.String("document_type", document.Type),
			zap.Int("existing_count", len(existingDocs)),
		)
		return nil, &domain.UserError{
			Code:    domain.USER_020,
			Message: s.localizer.Localize(ctx, domain.USER_020, nil),
		}
	}

	// Encrypt document content
	encryptedContent, encryptionKey, err := s.encryptionService.EncryptFile(document.Content)
	if err != nil {
		logger.Error("Failed to encrypt document", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_015,
			Message: s.localizer.Localize(ctx, domain.USER_015, nil),
		}
	}

	// Generate storage key
	documentID := uuid.New().String()
	storageKey := fmt.Sprintf("users/%s/documents/%s/%s", userID, document.Type, documentID)

	// Upload to storage
	contentReader := strings.NewReader(string(encryptedContent))
	metadata := map[string]string{
		"user_id":       userID,
		"document_type": document.Type,
		"original_size": fmt.Sprintf("%d", len(document.Content)),
		"upload_ip":     document.UploadIP,
		"document_id":   documentID,
	}

	if err := s.storageService.UploadFile(ctx, storageKey, contentReader, document.MimeType, metadata); err != nil {
		logger.Error("Failed to upload document to storage", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_016,
			Message: s.localizer.Localize(ctx, domain.USER_016, nil),
		}
	}

	// Create document record
	doc := &domain.Document{
		ID:            documentID,
		UserID:        userID,
		DocumentType:  document.Type,
		FilePath:      storageKey,
		FileSize:      int64(len(document.Content)),
		MimeType:      document.MimeType,
		EncryptionKey: encryptionKey,
		UploadIP:      document.UploadIP,
		CreatedAt:     time.Now(),
	}

	if err := s.documentRepo.CreateDocument(ctx, doc); err != nil {
		logger.Error("Failed to create document record", zap.Error(err))

		// Cleanup uploaded file
		if deleteErr := s.storageService.DeleteFile(ctx, storageKey); deleteErr != nil {
			logger.Error("Failed to cleanup uploaded file after database error", zap.Error(deleteErr))
		}

		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Log audit event
	if err := s.auditService.LogDocumentUploaded(ctx, userID, documentID, document.Type); err != nil {
		logger.Warn("Failed to log document upload audit event", zap.Error(err))
	}

	logger.Info("Document uploaded successfully",
		zap.String("document_id", documentID),
		zap.String("storage_key", storageKey),
		zap.Int64("file_size", doc.FileSize),
	)

	// Remove encryption key from response
	doc.EncryptionKey = ""
	return doc, nil
}

func (s *UserServiceImpl) GetDocuments(ctx context.Context, userID string) ([]*domain.Document, error) {
	logger := s.logger.With(
		zap.String("operation", "get_documents"),
		zap.String("user_id", userID),
	)

	// Check if user exists
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:    domain.USER_030,
				Message: s.localizer.Localize(ctx, domain.USER_030, nil),
			}
		}
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Get documents
	documents, err := s.documentRepo.GetDocumentsByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get documents", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Remove encryption keys from response
	for _, doc := range documents {
		doc.EncryptionKey = ""
	}

	logger.Info("Retrieved documents", zap.Int("document_count", len(documents)))
	return documents, nil
}

func (s *UserServiceImpl) GetDocument(ctx context.Context, userID, documentID string) (*domain.Document, error) {
	logger := s.logger.With(
		zap.String("operation", "get_document"),
		zap.String("user_id", userID),
		zap.String("document_id", documentID),
	)

	// Get document
	document, err := s.documentRepo.GetDocument(ctx, documentID)
	if err != nil {
		if err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:    domain.USER_014,
				Message: s.localizer.Localize(ctx, domain.USER_014, nil),
			}
		}
		logger.Error("Failed to get document", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Verify document belongs to user
	if document.UserID != userID {
		logger.Warn("Unauthorized document access attempt",
			zap.String("document_user_id", document.UserID),
			zap.String("requesting_user_id", userID),
		)
		return nil, &domain.UserError{
			Code:    domain.USER_032,
			Message: s.localizer.Localize(ctx, domain.USER_032, nil),
		}
	}

	// Log audit event for data access
	if err := s.auditService.LogDataAccess(ctx, userID, userID, "document"); err != nil {
		logger.Warn("Failed to log data access audit event", zap.Error(err))
	}

	// Remove encryption key from response
	document.EncryptionKey = ""
	return document, nil
}

func (s *UserServiceImpl) DownloadDocument(ctx context.Context, userID, documentID string) (*domain.DocumentStream, error) {
	logger := s.logger.With(
		zap.String("operation", "download_document"),
		zap.String("user_id", userID),
		zap.String("document_id", documentID),
	)

	// Get document metadata
	document, err := s.documentRepo.GetDocument(ctx, documentID)
	if err != nil {
		if err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:    domain.USER_014,
				Message: s.localizer.Localize(ctx, domain.USER_014, nil),
			}
		}
		logger.Error("Failed to get document", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Verify document belongs to user
	if document.UserID != userID {
		logger.Warn("Unauthorized document download attempt",
			zap.String("document_user_id", document.UserID),
			zap.String("requesting_user_id", userID),
		)
		return nil, &domain.UserError{
			Code:    domain.USER_032,
			Message: s.localizer.Localize(ctx, domain.USER_032, nil),
		}
	}

	// Download from storage
	fileReader, err := s.storageService.DownloadFile(ctx, document.FilePath)
	if err != nil {
		logger.Error("Failed to download file from storage", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_014,
			Message: s.localizer.Localize(ctx, domain.USER_014, nil),
		}
	}
	defer fileReader.Close()

	// Read encrypted content
	encryptedContent, err := io.ReadAll(fileReader)
	if err != nil {
		logger.Error("Failed to read file content", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_014,
			Message: s.localizer.Localize(ctx, domain.USER_014, nil),
		}
	}

	// Decrypt content
	decryptedContent, err := s.encryptionService.DecryptFile(encryptedContent, document.EncryptionKey)
	if err != nil {
		logger.Error("Failed to decrypt document", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_015,
			Message: s.localizer.Localize(ctx, domain.USER_015, nil),
		}
	}

	// Generate filename
	fileName := fmt.Sprintf("%s_%s%s",
		document.DocumentType,
		document.ID[:8],
		getFileExtension(document.MimeType),
	)

	// Log audit event for document download
	if err := s.auditService.LogDataAccess(ctx, userID, userID, "document_download"); err != nil {
		logger.Warn("Failed to log document download audit event", zap.Error(err))
	}

	logger.Info("Document downloaded successfully",
		zap.String("file_name", fileName),
		zap.Int("content_size", len(decryptedContent)),
	)

	return &domain.DocumentStream{
		Content:     decryptedContent,
		ContentType: document.MimeType,
		FileName:    fileName,
		Size:        int64(len(decryptedContent)),
	}, nil
}

func (s *UserServiceImpl) DeleteDocument(ctx context.Context, userID, documentID string) error {
	logger := s.logger.With(
		zap.String("operation", "delete_document"),
		zap.String("user_id", userID),
		zap.String("document_id", documentID),
	)

	// Get document
	document, err := s.documentRepo.GetDocument(ctx, documentID)
	if err != nil {
		if err.Error() == "not found" {
			return &domain.UserError{
				Code:    domain.USER_014,
				Message: s.localizer.Localize(ctx, domain.USER_014, nil),
			}
		}
		logger.Error("Failed to get document", zap.Error(err))
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Verify document belongs to user
	if document.UserID != userID {
		logger.Warn("Unauthorized document deletion attempt",
			zap.String("document_user_id", document.UserID),
			zap.String("requesting_user_id", userID),
		)
		return &domain.UserError{
			Code:    domain.USER_032,
			Message: s.localizer.Localize(ctx, domain.USER_032, nil),
		}
	}

	// Delete from storage
	if err := s.storageService.DeleteFile(ctx, document.FilePath); err != nil {
		logger.Error("Failed to delete file from storage", zap.Error(err))
		// Continue with database deletion even if storage deletion fails
	}

	// Delete from database
	if err := s.documentRepo.DeleteDocument(ctx, documentID); err != nil {
		logger.Error("Failed to delete document from database", zap.Error(err))
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	logger.Info("Document deleted successfully")
	return nil
}

// KYC-related methods

func (s *UserServiceImpl) InitiateKYC(ctx context.Context, userID string) (*domain.KYCSession, error) {
	logger := s.logger.With(
		zap.String("operation", "initiate_kyc"),
		zap.String("user_id", userID),
	)

	logger.Info("Starting KYC initiation")

	// Get user profile
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		if err.Error() == "not found" {
			return nil, &domain.UserError{
				Code:    domain.USER_031,
				Message: s.localizer.Localize(ctx, domain.USER_031, nil),
			}
		}
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Check if profile is complete enough for KYC
	if !s.isProfileReadyForKYC(profile) {
		return nil, &domain.UserError{
			Code:    domain.USER_005,
			Message: s.localizer.Localize(ctx, domain.USER_005, nil),
		}
	}

	// Check if KYC is already completed
	kycStatus, err := s.kycRepo.GetKYCStatus(ctx, userID)
	if err != nil && err.Error() != "not found" {
		logger.Error("Failed to get KYC status", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	if kycStatus != nil {
		for _, status := range kycStatus {
			if status == domain.KYCStatusVerified {
				return nil, &domain.UserError{
					Code:    domain.USER_010,
					Message: s.localizer.Localize(ctx, domain.USER_010, nil),
				}
			}
		}
	}

	// Initiate KYC with provider
	session, err := s.kycProvider.InitiateIdentityVerification(ctx, userID, profile)
	if err != nil {
		logger.Error("Failed to initiate KYC with provider", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_021,
			Message: s.localizer.Localize(ctx, domain.USER_021, nil),
		}
	}

	// Create KYC verification record
	verification := &domain.KYCVerification{
		ID:                uuid.New().String(),
		UserID:            userID,
		VerificationType:  "identity",
		Provider:          s.kycProvider.GetProviderName(),
		Status:            domain.KYCStatusPending,
		ProviderReference: session.ProviderReference,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.kycRepo.CreateKYCVerification(ctx, verification); err != nil {
		logger.Error("Failed to create KYC verification record", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	logger.Info("KYC session initiated successfully",
		zap.String("session_id", session.ID),
		zap.String("provider", session.Provider),
	)

	return session, nil
}

func (s *UserServiceImpl) GetKYCStatus(ctx context.Context, userID string) (map[string]domain.KYCStatus, error) {
	logger := s.logger.With(
		zap.String("operation", "get_kyc_status"),
		zap.String("user_id", userID),
	)

	// Try cache first
	if cachedStatus, err := s.cacheService.GetCachedKYCStatus(ctx, userID); err == nil && cachedStatus != nil {
		logger.Debug("KYC status found in cache")
		return cachedStatus, nil
	}

	// Get from database
	status, err := s.kycRepo.GetKYCStatus(ctx, userID)
	if err != nil {
		if err.Error() == "not found" {
			// Return empty status map if no KYC records found
			return make(map[string]domain.KYCStatus), nil
		}
		logger.Error("Failed to get KYC status", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Cache the status
	if err := s.cacheService.CacheKYCStatus(ctx, userID, status, 1800); err != nil { // 30 minutes
		logger.Warn("Failed to cache KYC status", zap.Error(err))
	}

	return status, nil
}

func (s *UserServiceImpl) UpdateKYCStatus(ctx context.Context, userID, verificationType string, status domain.KYCStatus, data map[string]interface{}) error {
	logger := s.logger.With(
		zap.String("operation", "update_kyc_status"),
		zap.String("user_id", userID),
		zap.String("verification_type", verificationType),
		zap.String("new_status", string(status)),
	)

	// Get existing KYC verification
	existingVerification, err := s.kycRepo.GetKYCVerification(ctx, userID, verificationType)
	if err != nil {
		if err.Error() == "not found" {
			return &domain.UserError{
				Code:    domain.USER_021,
				Message: s.localizer.Localize(ctx, domain.USER_021, nil),
			}
		}
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	oldStatus := existingVerification.Status

	// Update KYC verification
	updates := map[string]interface{}{
		"status":            status,
		"verification_data": data,
		"updated_at":        time.Now(),
	}

	if err := s.kycRepo.UpdateKYCVerification(ctx, existingVerification.ID, updates); err != nil {
		logger.Error("Failed to update KYC verification", zap.Error(err))
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Invalidate cache
	if err := s.cacheService.InvalidateKYCStatus(ctx, userID); err != nil {
		logger.Warn("Failed to invalidate KYC status cache", zap.Error(err))
	}

	// Log audit event
	if err := s.auditService.LogKYCStatusChanged(ctx, userID, verificationType, oldStatus, status); err != nil {
		logger.Warn("Failed to log KYC status change audit event", zap.Error(err))
	}

	logger.Info("KYC status updated successfully")
	return nil
}

// Search and listing methods

func (s *UserServiceImpl) SearchUsers(ctx context.Context, criteria map[string]interface{}, offset, limit int) ([]*domain.User, error) {
	logger := s.logger.With(
		zap.String("operation", "search_users"),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	users, err := s.userRepo.SearchUsers(ctx, criteria)
	if err != nil {
		logger.Error("Failed to search users", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Remove password hashes from response
	for _, user := range users {
		user.PasswordHash = ""
	}

	logger.Info("User search completed", zap.Int("result_count", len(users)))
	return users, nil
}

func (s *UserServiceImpl) ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	logger := s.logger.With(
		zap.String("operation", "list_users"),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	users, err := s.userRepo.ListUsers(ctx, offset, limit)
	if err != nil {
		logger.Error("Failed to list users", zap.Error(err))
		return nil, &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Remove password hashes from response
	for _, user := range users {
		user.PasswordHash = ""
	}

	logger.Info("User listing completed", zap.Int("result_count", len(users)))
	return users, nil
}

// Verification methods (email and phone)

func (s *UserServiceImpl) SendEmailVerification(ctx context.Context, userID string) error {
	logger := s.logger.With(
		zap.String("operation", "send_email_verification"),
		zap.String("user_id", userID),
	)

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err.Error() == "not found" {
			return &domain.UserError{
				Code:    domain.USER_030,
				Message: s.localizer.Localize(ctx, domain.USER_030, nil),
			}
		}
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	if user.EmailVerified {
		return &domain.UserError{
			Code:    domain.USER_005,
			Message: s.localizer.Localize(ctx, domain.USER_005, nil),
		}
	}

	// Generate verification code (in real implementation, store this securely)
	verificationCode := generateVerificationCode()

	// Send verification email
	if err := s.notificationService.SendEmailVerification(ctx, userID, user.Email, verificationCode); err != nil {
		logger.Error("Failed to send email verification", zap.Error(err))
		return &domain.UserError{
			Code:    domain.USER_029,
			Message: s.localizer.Localize(ctx, domain.USER_029, nil),
		}
	}

	logger.Info("Email verification sent successfully")
	return nil
}

func (s *UserServiceImpl) VerifyEmail(ctx context.Context, userID, verificationCode string) error {
	logger := s.logger.With(
		zap.String("operation", "verify_email"),
		zap.String("user_id", userID),
	)

	// In real implementation, validate the verification code against stored value
	// For now, we'll assume validation passed

	// Update user's email verification status
	updates := map[string]interface{}{
		"email_verified": true,
		"updated_at":     time.Now(),
	}

	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		logger.Error("Failed to update email verification status", zap.Error(err))
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Invalidate user cache
	if err := s.cacheService.InvalidateUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to invalidate user cache", zap.Error(err))
	}

	logger.Info("Email verified successfully")
	return nil
}

func (s *UserServiceImpl) SendPhoneVerification(ctx context.Context, userID string) error {
	logger := s.logger.With(
		zap.String("operation", "send_phone_verification"),
		zap.String("user_id", userID),
	)

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if err.Error() == "not found" {
			return &domain.UserError{
				Code:    domain.USER_030,
				Message: s.localizer.Localize(ctx, domain.USER_030, nil),
			}
		}
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	if user.PhoneVerified {
		return &domain.UserError{
			Code:    domain.USER_005,
			Message: s.localizer.Localize(ctx, domain.USER_005, nil),
		}
	}

	if user.Phone == "" {
		return &domain.UserError{
			Code:    domain.USER_005,
			Message: s.localizer.Localize(ctx, domain.USER_005, nil),
		}
	}

	// Generate verification code
	verificationCode := generateVerificationCode()

	// Send verification SMS
	if err := s.notificationService.SendPhoneVerification(ctx, userID, user.Phone, verificationCode); err != nil {
		logger.Error("Failed to send phone verification", zap.Error(err))
		return &domain.UserError{
			Code:    domain.USER_029,
			Message: s.localizer.Localize(ctx, domain.USER_029, nil),
		}
	}

	logger.Info("Phone verification sent successfully")
	return nil
}

func (s *UserServiceImpl) VerifyPhone(ctx context.Context, userID, verificationCode string) error {
	logger := s.logger.With(
		zap.String("operation", "verify_phone"),
		zap.String("user_id", userID),
	)

	// In real implementation, validate the verification code against stored value
	// For now, we'll assume validation passed

	// Update user's phone verification status
	updates := map[string]interface{}{
		"phone_verified": true,
		"updated_at":     time.Now(),
	}

	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		logger.Error("Failed to update phone verification status", zap.Error(err))
		return &domain.UserError{
			Code:    domain.USER_026,
			Message: s.localizer.Localize(ctx, domain.USER_026, nil),
		}
	}

	// Invalidate user cache
	if err := s.cacheService.InvalidateUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to invalidate user cache", zap.Error(err))
	}

	logger.Info("Phone verified successfully")
	return nil
}

// Helper methods

func (s *UserServiceImpl) validateDocumentUpload(document *domain.DocumentUpload) error {
	if document.Type == "" {
		return &domain.UserError{
			Code:    domain.USER_005,
			Message: s.localizer.Localize(context.Background(), domain.USER_005, nil),
		}
	}

	if err := s.validationService.ValidateDocumentType(document.Type); err != nil {
		return &domain.UserError{
			Code:    domain.USER_017,
			Message: s.localizer.Localize(context.Background(), domain.USER_017, nil),
		}
	}

	if len(document.Content) == 0 {
		return &domain.UserError{
			Code:    domain.USER_005,
			Message: s.localizer.Localize(context.Background(), domain.USER_005, nil),
		}
	}

	if err := s.validationService.ValidateFileSize(int64(len(document.Content))); err != nil {
		return &domain.UserError{
			Code:    domain.USER_012,
			Message: s.localizer.Localize(context.Background(), domain.USER_012, nil),
		}
	}

	if err := s.validationService.ValidateMimeType(document.MimeType); err != nil {
		return &domain.UserError{
			Code:    domain.USER_011,
			Message: s.localizer.Localize(context.Background(), domain.USER_011, nil),
		}
	}

	if err := s.validationService.ValidateDocument(document.Type, document.Content, document.MimeType); err != nil {
		return &domain.UserError{
			Code:    domain.USER_011,
			Message: s.localizer.Localize(context.Background(), domain.USER_011, nil),
		}
	}

	return nil
}

func (s *UserServiceImpl) isProfileReadyForKYC(profile *domain.UserProfile) bool {
	return profile.FirstName != "" &&
		profile.LastName != "" &&
		!profile.DateOfBirth.IsZero() &&
		profile.Phone != ""
}

func getFileExtension(mimeType string) string {
	switch mimeType {
	case "application/pdf":
		return ".pdf"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}

func generateVerificationCode() string {
	// In real implementation, use crypto/rand for secure random generation
	return fmt.Sprintf("%06d", time.Now().Unix()%1000000)
}
