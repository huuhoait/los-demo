package interfaces

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/user/domain"
	"github.com/huuhoait/los-demo/services/user/interfaces/middleware"
	"github.com/huuhoait/los-demo/services/shared/pkg/errors"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
)

type UserHandler struct {
	userService domain.UserService
	logger      *zap.Logger
	localizer   *i18n.Localizer
}

func NewUserHandler(userService domain.UserService, logger *zap.Logger, localizer *i18n.Localizer) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
		localizer:   localizer,
	}
}

// RegisterRoutes registers all user-related routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	// User management routes
	router.POST("/users", h.CreateUser)
	router.GET("/users/:id", h.GetUser)
	router.PUT("/users/:id", h.UpdateUser)
	router.DELETE("/users/:id", h.DeleteUser)
	router.GET("/users", h.ListUsers)
	router.POST("/users/search", h.SearchUsers)

	// Profile management routes
	router.GET("/users/:id/profile", h.GetProfile)
	router.PUT("/users/:id/profile", h.UpdateProfile)

	// Verification routes
	router.POST("/users/:id/verify-email", h.SendEmailVerification)
	router.POST("/users/:id/verify-email/confirm", h.VerifyEmail)
	router.POST("/users/:id/verify-phone", h.SendPhoneVerification)
	router.POST("/users/:id/verify-phone/confirm", h.VerifyPhone)

	// KYC routes
	router.POST("/users/:id/kyc/initiate", h.InitiateKYC)
	router.GET("/users/:id/kyc/status", h.GetKYCStatus)
	router.PUT("/users/:id/kyc/status", h.UpdateKYCStatus)

	// Document management routes
	router.POST("/users/:id/documents", h.UploadDocument)
	router.GET("/users/:id/documents", h.GetDocuments)
	router.GET("/users/:id/documents/:doc_id", h.GetDocument)
	router.GET("/users/:id/documents/:doc_id/download", h.DownloadDocument)
	router.DELETE("/users/:id/documents/:doc_id", h.DeleteDocument)
}

// User Management Handlers

func (h *UserHandler) CreateUser(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "create_user"),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request domain.CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		h.respondValidationError(c, map[string]string{
			"request_body": "invalid_format",
		})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &request)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("User created successfully", zap.String("user_id", user.ID))
	h.respondSuccessWithMessage(c, http.StatusCreated, "user_created", user, map[string]interface{}{
		"user_id": user.ID,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "get_user"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("User retrieved successfully")
	h.respondSuccess(c, http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "update_user"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Invalid request body",
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, &request)
	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("User updated successfully")
	h.respondSuccess(c, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "delete_user"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("User deleted successfully")
	h.respondSuccess(c, http.StatusNoContent, nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "list_users"),
		zap.String("request_id", c.GetString("request_id")),
	)

	// Parse pagination parameters
	offset := 0
	limit := 50 // Default limit

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	users, err := h.userService.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		logger.Error("Failed to list users", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Users listed successfully", zap.Int("count", len(users)))
	h.respondSuccess(c, http.StatusOK, gin.H{
		"users":  users,
		"offset": offset,
		"limit":  limit,
		"count":  len(users),
	})
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "search_users"),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request struct {
		Email         string `json:"email,omitempty"`
		Phone         string `json:"phone,omitempty"`
		Status        string `json:"status,omitempty"`
		EmailVerified *bool  `json:"email_verified,omitempty"`
		PhoneVerified *bool  `json:"phone_verified,omitempty"`
		Offset        int    `json:"offset,omitempty"`
		Limit         int    `json:"limit,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Invalid request body",
		})
		return
	}

	// Build search criteria
	criteria := make(map[string]interface{})
	if request.Email != "" {
		criteria["email"] = request.Email
	}
	if request.Phone != "" {
		criteria["phone"] = request.Phone
	}
	if request.Status != "" {
		criteria["status"] = request.Status
	}
	if request.EmailVerified != nil {
		criteria["email_verified"] = *request.EmailVerified
	}
	if request.PhoneVerified != nil {
		criteria["phone_verified"] = *request.PhoneVerified
	}

	// Default pagination
	offset := request.Offset
	limit := request.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	users, err := h.userService.SearchUsers(c.Request.Context(), criteria, offset, limit)
	if err != nil {
		logger.Error("Failed to search users", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("User search completed", zap.Int("count", len(users)))
	h.respondSuccess(c, http.StatusOK, gin.H{
		"users":    users,
		"criteria": criteria,
		"offset":   offset,
		"limit":    limit,
		"count":    len(users),
	})
}

// Profile Management Handlers

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "get_profile"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get profile", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Profile retrieved successfully")
	h.respondSuccess(c, http.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "update_profile"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Invalid request body",
		})
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, &request)
	if err != nil {
		logger.Error("Failed to update profile", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Profile updated successfully")
	h.respondSuccess(c, http.StatusOK, profile)
}

// Verification Handlers

func (h *UserHandler) SendEmailVerification(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "send_email_verification"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	err := h.userService.SendEmailVerification(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to send email verification", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Email verification sent successfully")
	h.respondSuccess(c, http.StatusOK, gin.H{
		"message": "Email verification sent",
	})
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "verify_email"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request struct {
		VerificationCode string `json:"verification_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Verification code is required",
		})
		return
	}

	err := h.userService.VerifyEmail(c.Request.Context(), userID, request.VerificationCode)
	if err != nil {
		logger.Error("Failed to verify email", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Email verified successfully")
	h.respondSuccess(c, http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

func (h *UserHandler) SendPhoneVerification(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "send_phone_verification"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	err := h.userService.SendPhoneVerification(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to send phone verification", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Phone verification sent successfully")
	h.respondSuccess(c, http.StatusOK, gin.H{
		"message": "Phone verification sent",
	})
}

func (h *UserHandler) VerifyPhone(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "verify_phone"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request struct {
		VerificationCode string `json:"verification_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Verification code is required",
		})
		return
	}

	err := h.userService.VerifyPhone(c.Request.Context(), userID, request.VerificationCode)
	if err != nil {
		logger.Error("Failed to verify phone", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("Phone verified successfully")
	h.respondSuccess(c, http.StatusOK, gin.H{
		"message": "Phone verified successfully",
	})
}

// KYC Handlers

func (h *UserHandler) InitiateKYC(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "initiate_kyc"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	session, err := h.userService.InitiateKYC(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to initiate KYC", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("KYC initiated successfully")
	h.respondSuccess(c, http.StatusOK, session)
}

func (h *UserHandler) GetKYCStatus(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "get_kyc_status"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	status, err := h.userService.GetKYCStatus(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get KYC status", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("KYC status retrieved successfully")
	h.respondSuccess(c, http.StatusOK, gin.H{
		"user_id": userID,
		"status":  status,
	})
}

func (h *UserHandler) UpdateKYCStatus(c *gin.Context) {
	userID := c.Param("id")
	logger := h.logger.With(
		zap.String("operation", "update_kyc_status"),
		zap.String("user_id", userID),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request struct {
		VerificationType string                 `json:"verification_type" binding:"required"`
		Status           domain.KYCStatus       `json:"status" binding:"required"`
		Data             map[string]interface{} `json:"data,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		h.respondError(c, &errors.ServiceError{
			Code:        domain.USER_005,
			Message:     h.localizer.GetErrorMessage("en", domain.USER_005, nil),
			Description: "Invalid request body",
		})
		return
	}

	err := h.userService.UpdateKYCStatus(c.Request.Context(), userID, request.VerificationType, request.Status, request.Data)
	if err != nil {
		logger.Error("Failed to update KYC status", zap.Error(err))
		h.respondError(c, err)
		return
	}

	logger.Info("KYC status updated successfully")
	h.respondSuccess(c, http.StatusOK, gin.H{
		"message": "KYC status updated successfully",
	})
}

// Helper methods

func (h *UserHandler) respondSuccess(c *gin.Context, status int, data interface{}) {
	// Create localized success response
	lang := middleware.GetLanguageFromGinContext(c)

	response := gin.H{
		"success":    true,
		"data":       data,
		"language":   lang,
		"request_id": c.GetString("request_id"),
		"timestamp":  c.GetTime("timestamp"),
		"service":    "user-service",
	}

	c.JSON(status, response)
}

func (h *UserHandler) respondSuccessWithMessage(c *gin.Context, status int, messageKey string, data interface{}, templateData map[string]interface{}) {
	// Create localized success response with message
	lang := middleware.GetLanguageFromGinContext(c)
	message := h.localizer.LocalizeMessage(lang, messageKey, templateData)

	response := gin.H{
		"success":    true,
		"message":    message,
		"data":       data,
		"language":   lang,
		"request_id": c.GetString("request_id"),
		"timestamp":  c.GetTime("timestamp"),
		"service":    "user-service",
	}

	c.JSON(status, response)
}

func (h *UserHandler) respondError(c *gin.Context, err error) {
	lang := middleware.GetLanguageFromGinContext(c)

	// Handle domain errors with localization
	if domainErr, ok := err.(*domain.UserError); ok {
		statusCode := h.getHTTPStatusFromErrorCode(domainErr.Code)
		message := h.localizer.LocalizeError(lang, domainErr.Code, domainErr.TemplateData)

		response := gin.H{
			"success": false,
			"error": gin.H{
				"code":    domainErr.Code,
				"message": message,
			},
			"language":   lang,
			"request_id": c.GetString("request_id"),
			"timestamp":  c.GetTime("timestamp"),
			"service":    "user-service",
		}

		if domainErr.Field != "" {
			response["error"].(gin.H)["field"] = domainErr.Field
		}

		c.JSON(statusCode, response)
		return
	}

	// Handle generic errors
	statusCode := http.StatusInternalServerError
	message := h.localizer.LocalizeError(lang, "USER_033", nil)

	response := gin.H{
		"success": false,
		"error": gin.H{
			"code":    "USER_033",
			"message": message,
		},
		"language":   lang,
		"request_id": c.GetString("request_id"),
		"timestamp":  c.GetTime("timestamp"),
		"service":    "user-service",
	}

	c.JSON(statusCode, response)
}

func (h *UserHandler) respondValidationError(c *gin.Context, validationErrors map[string]string) {
	response := middleware.CreateValidationErrorResponse(c, h.localizer, validationErrors)
	c.JSON(http.StatusBadRequest, response)
}

func (h *UserHandler) getHTTPStatusFromErrorCode(code string) int {
	switch {
	case strings.HasPrefix(code, "USER_001"), strings.HasPrefix(code, "USER_002"),
		strings.HasPrefix(code, "USER_003"), strings.HasPrefix(code, "USER_004"),
		strings.HasPrefix(code, "USER_005"), strings.HasPrefix(code, "USER_011"),
		strings.HasPrefix(code, "USER_012"), strings.HasPrefix(code, "USER_017"):
		return http.StatusBadRequest
	case code == domain.USER_006, code == domain.USER_007, code == domain.USER_008,
		code == domain.USER_010, code == domain.USER_020:
		return http.StatusConflict
	case code == domain.USER_030, code == domain.USER_031, code == domain.USER_014:
		return http.StatusNotFound
	case code == domain.USER_032:
		return http.StatusForbidden
	case code == domain.USER_033:
		return http.StatusTooManyRequests
	case strings.HasPrefix(code, "USER_026"), strings.HasPrefix(code, "USER_027"),
		strings.HasPrefix(code, "USER_028"), strings.HasPrefix(code, "USER_029"),
		strings.HasPrefix(code, "USER_034"), strings.HasPrefix(code, "USER_035"):
		return http.StatusInternalServerError
	case strings.HasPrefix(code, "USER_021"), strings.HasPrefix(code, "USER_025"):
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
