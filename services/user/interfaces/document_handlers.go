package interfaces

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"our-los/services/user/domain"
	"our-los/services/user/pkg/errors"
)

// Document Management Handlers

func (h *UserHandler) UploadDocument(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "upload_document"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Invalid form data",
		})
		return
	}

	// Get document type
	documentType := c.PostForm("document_type")
	if documentType == "" {
		logger.Error("Missing document type")
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Document type is required",
			Field:       "document_type",
		})
		return
	}

	// Get uploaded file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error("Failed to get uploaded file", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "File upload is required",
			Field:       "file",
		})
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read file content", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_013,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_013, nil),
			Description: "Failed to read uploaded file",
		})
		return
	}

	// Validate file size
	if len(content) == 0 {
		logger.Error("Empty file uploaded")
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "File cannot be empty",
			Field:       "file",
		})
		return
	}

	// Detect mime type from file extension and content
	mimeType := h.detectMimeType(fileHeader.Filename, content)
	
	// Get client IP
	clientIP := h.getClientIP(c)

	// Create document upload request
	documentUpload := &domain.DocumentUpload{
		Type:     documentType,
		Content:  content,
		MimeType: mimeType,
		UploadIP: clientIP,
	}

	// Upload document
	document, err := h.userService.UploadDocument(c.Request.Context(), userID, documentUpload)
	if err != nil {
		logger.Error("Failed to upload document", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Document uploaded successfully", 
		zap.String("document_id", document.ID),
		zap.String("document_type", document.DocumentType),
		zap.Int64("file_size", document.FileSize),
	)

	h.respondSuccess(c, http.StatusCreated, document)
}

func (h *UserHandler) GetDocuments(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "get_documents"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	documents, err := h.userService.GetDocuments(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get documents", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Documents retrieved successfully", zap.Int("count", len(documents)))
	h.respondSuccess(c, http.StatusOK, gin.H{
		"documents": documents,
		"count":     len(documents),
	})
}

func (h *UserHandler) GetDocument(c *gin.Context) {
	userID := c.Param("id")
	documentID := c.Param("doc_id")
	logger := h.logger.With(
		zap.String("operation", "get_document"),
		zap.String("user_id", userID),
		zap.String("document_id", documentID),
		zap.String("request_id", c.GetString("request_id")),
	)

	document, err := h.userService.GetDocument(c.Request.Context(), userID, documentID)
	if err != nil {
		logger.Error("Failed to get document", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Document retrieved successfully")
	h.respondSuccess(c, http.StatusOK, document)
}

func (h *UserHandler) DownloadDocument(c *gin.Context) {
	userID := c.Param("id")
	documentID := c.Param("doc_id")
	logger := h.logger.With(
		zap.String("operation", "download_document"),
		zap.String("user_id", userID),
		zap.String("document_id", documentID),
		zap.String("request_id", c.GetString("request_id")),
	)

	documentStream, err := h.userService.DownloadDocument(c.Request.Context(), userID, documentID)
	if err != nil {
		logger.Error("Failed to download document", zap.Error(err))
		h.respondError(c, err)
		return
	}

	// Set headers for file download
	c.Header("Content-Type", documentStream.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", documentStream.FileName))
	c.Header("Content-Length", fmt.Sprintf("%d", documentStream.Size))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Write file content
	c.Data(http.StatusOK, documentStream.ContentType, documentStream.Content)

	logger.Info("Document downloaded successfully", 
		zap.String("file_name", documentStream.FileName),
		zap.Int64("file_size", documentStream.Size),
	)
}

func (h *UserHandler) DeleteDocument(c *gin.Context) {
	userID := c.Param("id")
	documentID := c.Param("doc_id")
	logger := h.logger.With(
		zap.String("operation", "delete_document"),
		zap.String("user_id", userID),
		zap.String("document_id", documentID),
		zap.String("request_id", c.GetString("request_id")),
	)

	err := h.userService.DeleteDocument(c.Request.Context(), userID, documentID)
	if err != nil {
		logger.Error("Failed to delete document", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Document deleted successfully")
	h.respondSuccess(c, http.StatusNoContent, nil)
}

// Helper methods for document handling

func (h *UserHandler) detectMimeType(filename string, content []byte) string {
	// First, try to detect from file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".tiff", ".tif":
		return "image/tiff"
	case ".bmp":
		return "image/bmp"
	}

	// If extension detection fails, try content detection
	if len(content) >= 4 {
		// PDF magic number
		if content[0] == 0x25 && content[1] == 0x50 && content[2] == 0x44 && content[3] == 0x46 {
			return "application/pdf"
		}
		
		// JPEG magic number
		if content[0] == 0xFF && content[1] == 0xD8 {
			return "image/jpeg"
		}
		
		// PNG magic number
		if len(content) >= 8 && content[0] == 0x89 && content[1] == 0x50 && content[2] == 0x4E && content[3] == 0x47 {
			return "image/png"
		}
		
		// GIF magic number
		if len(content) >= 6 && content[0] == 0x47 && content[1] == 0x49 && content[2] == 0x46 {
			return "image/gif"
		}
	}

	// Default to octet-stream if detection fails
	return "application/octet-stream"
}

func (h *UserHandler) getClientIP(c *gin.Context) string {
	// Try to get real IP from headers (for cases behind proxy/load balancer)
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIndex := strings.Index(ip, ","); commaIndex != -1 {
			return strings.TrimSpace(ip[:commaIndex])
		}
		return ip
	}
	
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	
	if ip := c.GetHeader("X-Client-IP"); ip != "" {
		return ip
	}
	
	// Fall back to RemoteAddr
	return c.ClientIP()
}

func (h *UserHandler) validateFileUpload(fileHeader *multipart.FileHeader, content []byte) error {
	// Check file size (10MB max)
	if len(content) > 10*1024*1024 {
		return &errors.ServiceError{
			Code:        domain.USER_012,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_012, nil),
			Description: "File size exceeds 10MB limit",
		}
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".tiff", ".tif", ".bmp"}
	
	isValidExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isValidExt = true
			break
		}
	}
	
	if !isValidExt {
		return &errors.ServiceError{
			Code:        domain.USER_017,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_017, nil),
			Description: "Unsupported file type",
		}
	}

	// Basic virus/malware check - scan for suspicious patterns
	if h.containsSuspiciousContent(content) {
		return &errors.ServiceError{
			Code:        domain.USER_018,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_018, nil),
			Description: "File contains suspicious content",
		}
	}

	return nil
}

func (h *UserHandler) containsSuspiciousContent(content []byte) bool {
	// This is a basic check - in production, you'd use a proper antivirus engine
	// Check for executable signatures
	suspiciousSignatures := [][]byte{
		{0x4D, 0x5A}, // DOS/Windows executable
		{0x7F, 0x45, 0x4C, 0x46}, // ELF executable
		{0xFE, 0xED, 0xFA, 0xCE}, // Mach-O executable (little endian)
		{0xCE, 0xFA, 0xED, 0xFE}, // Mach-O executable (big endian)
	}

	if len(content) < 4 {
		return false
	}

	for _, signature := range suspiciousSignatures {
		if len(content) >= len(signature) {
			match := true
			for i, b := range signature {
				if content[i] != b {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	return false
}
