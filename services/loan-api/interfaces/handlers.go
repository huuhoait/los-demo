package interfaces

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/loan-api/application"
	"github.com/huuhoait/los-demo/services/loan-api/domain"
	"github.com/huuhoait/los-demo/services/loan-api/interfaces/middleware"
	"github.com/huuhoait/los-demo/services/loan-api/pkg/i18n"
)

// LoanHandler handles HTTP requests for loan operations
type LoanHandler struct {
	loanService *application.LoanService
	logger      *zap.Logger
	localizer   *i18n.Localizer
	validate    *validator.Validate
}

// NewLoanHandler creates a new loan handler
func NewLoanHandler(loanService *application.LoanService, logger *zap.Logger, localizer *i18n.Localizer) *LoanHandler {
	return &LoanHandler{
		loanService: loanService,
		logger:      logger,
		localizer:   localizer,
		validate:    validator.New(),
	}
}

// CreateApplication creates a new loan application
// @Summary Create a new loan application
// @Description Create a new loan application with the provided details
// @Tags Applications
// @Accept json
// @Produce json
// @Param application body domain.CreateApplicationRequest true "Loan application details"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=domain.LoanApplication} "Application created successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid request data"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/applications [post]
func (h *LoanHandler) CreateApplication(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "create_application"),
		zap.String("ip_address", c.ClientIP()),
	)

	// Get user ID from context (assuming auth middleware sets this)
	// User ID is no longer needed as it's extracted from the request body
	// The service will create or find the user based on the email in the request

	var req domain.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request format", zap.Error(err))

		// Create detailed error information
		errorDetails := map[string]interface{}{
			"validation_error": err.Error(),
			"field_errors":     getFieldErrors(err),
			"request_body":     getRequestBody(c),
		}

		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, errorDetails)
		return
	}

	application, err := h.loanService.CreateApplication(c.Request.Context(), &req)
	if err != nil {
		if loanErr, ok := err.(*domain.LoanError); ok {
			logger.Warn("Failed to create application",
				zap.String("error_code", loanErr.Code),
				zap.Error(err))
			middleware.CreateErrorResponse(c, loanErr.HTTPStatus, loanErr.Code, nil)
			return
		}

		logger.Error("Unexpected error creating application", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusInternalServerError, domain.LOAN_023, nil)
		return
	}

	logger.Info("Application created successfully",
		zap.String("application_id", application.ID))

	middleware.CreateSuccessResponse(c, application, "APPLICATION_CREATED", nil)
}

// GetApplication retrieves a loan application by ID
// @Summary Get a loan application by ID
// @Description Retrieve a specific loan application by its ID
// @Tags Applications
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=domain.LoanApplication} "Application retrieved successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid application ID"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 404 {object} middleware.ErrorResponse "Application not found"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/applications/{id} [get]
func (h *LoanHandler) GetApplication(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "get_application"),
	)

	applicationID := c.Param("id")
	if applicationID == "" {
		logger.Warn("Missing application ID")
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	application, err := h.loanService.GetApplication(c.Request.Context(), applicationID)
	if err != nil {
		if loanErr, ok := err.(*domain.LoanError); ok {
			logger.Warn("Failed to get application",
				zap.String("error_code", loanErr.Code),
				zap.String("application_id", applicationID),
				zap.Error(err))
			middleware.CreateErrorResponse(c, loanErr.HTTPStatus, loanErr.Code, nil)
			return
		}

		logger.Error("Unexpected error getting application", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusInternalServerError, domain.LOAN_023, nil)
		return
	}

	middleware.CreateSuccessResponse(c, application, "", nil)
}

// UpdateApplication updates a loan application
// PUT /v1/loans/applications/:id
func (h *LoanHandler) UpdateApplication(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "update_application"),
	)

	applicationID := c.Param("id")
	if applicationID == "" {
		logger.Warn("Missing application ID")
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	var req domain.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request format", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	application, err := h.loanService.UpdateApplication(c.Request.Context(), applicationID, &req)
	if err != nil {
		if loanErr, ok := err.(*domain.LoanError); ok {
			logger.Warn("Failed to update application",
				zap.String("error_code", loanErr.Code),
				zap.String("application_id", applicationID),
				zap.Error(err))
			middleware.CreateErrorResponse(c, loanErr.HTTPStatus, loanErr.Code, nil)
			return
		}

		logger.Error("Unexpected error updating application", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusInternalServerError, domain.LOAN_023, nil)
		return
	}

	logger.Info("Application updated successfully",
		zap.String("application_id", applicationID))

	middleware.CreateSuccessResponse(c, application, "APPLICATION_UPDATED", nil)
}

// SubmitApplication submits a draft application for processing
// POST /v1/loans/applications/:id/submit
func (h *LoanHandler) SubmitApplication(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "submit_application"),
	)

	applicationID := c.Param("id")
	if applicationID == "" {
		logger.Warn("Missing application ID")
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	application, err := h.loanService.SubmitApplication(c.Request.Context(), applicationID)
	if err != nil {
		if loanErr, ok := err.(*domain.LoanError); ok {
			logger.Warn("Failed to submit application",
				zap.String("error_code", loanErr.Code),
				zap.String("application_id", applicationID),
				zap.Error(err))
			middleware.CreateErrorResponse(c, loanErr.HTTPStatus, loanErr.Code, nil)
			return
		}

		logger.Error("Unexpected error submitting application", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusInternalServerError, domain.LOAN_023, nil)
		return
	}

	logger.Info("Application submitted successfully",
		zap.String("application_id", applicationID))

	middleware.CreateSuccessResponse(c, application, "APPLICATION_SUBMITTED", nil)
}

// GetApplicationsByUser retrieves all applications for the current user
// @Summary Get all loan applications for the current user
// @Description Retrieve all loan applications associated with the authenticated user
// @Tags Applications
// @Accept json
// @Produce json
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=[]domain.LoanApplication} "Applications retrieved successfully"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/applications [get]
func (h *LoanHandler) GetApplicationsByUser(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "get_applications_by_user"),
	)

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		middleware.CreateErrorResponse(c, http.StatusUnauthorized, domain.LOAN_022, nil)
		return
	}

	applications, err := h.loanService.GetApplicationsByUser(c.Request.Context(), userID.(string))
	if err != nil {
		if loanErr, ok := err.(*domain.LoanError); ok {
			logger.Warn("Failed to get applications",
				zap.String("error_code", loanErr.Code),
				zap.String("user_id", userID.(string)),
				zap.Error(err))
			middleware.CreateErrorResponse(c, loanErr.HTTPStatus, loanErr.Code, nil)
			return
		}

		logger.Error("Unexpected error getting applications", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusInternalServerError, domain.LOAN_023, nil)
		return
	}

	middleware.CreateSuccessResponse(c, applications, "", nil)
}

// PreQualify performs pre-qualification check
// @Summary Perform loan pre-qualification
// @Description Check if a user qualifies for a loan based on income, debt, and other factors
// @Tags Pre-qualification
// @Accept json
// @Produce json
// @Param request body domain.PreQualifyRequest true "Pre-qualification request"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=domain.PreQualifyResult} "Pre-qualification completed"
// @Failure 400 {object} middleware.ErrorResponse "Invalid request data"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/prequalify [post]
func (h *LoanHandler) PreQualify(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "pre_qualify"),
	)

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		middleware.CreateErrorResponse(c, http.StatusUnauthorized, domain.LOAN_022, nil)
		return
	}

	var req domain.PreQualifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request format", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	// TODO: Implement pre-qualification workflow initiation
	// For now, return a placeholder response
	logger.Info("Pre-qualification workflow initiated",
		zap.String("user_id", userID.(string)))

	middleware.CreateSuccessResponse(c, gin.H{
		"message": "Pre-qualification workflow initiated",
		"status":  "pending",
	}, "PRE_QUALIFICATION_SUCCESS", nil)
}

// GenerateOffer generates a loan offer for an application
// POST /v1/loans/applications/:id/offer
func (h *LoanHandler) GenerateOffer(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "generate_offer"),
	)

	applicationID := c.Param("id")
	if applicationID == "" {
		logger.Warn("Missing application ID")
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	// TODO: Implement offer generation workflow initiation
	// For now, return a placeholder response
	logger.Info("Offer generation workflow initiated",
		zap.String("application_id", applicationID))

	middleware.CreateSuccessResponse(c, gin.H{
		"message":        "Offer generation workflow initiated",
		"status":         "pending",
		"application_id": applicationID,
	}, "OFFER_GENERATED", nil)
}

// AcceptOffer accepts a loan offer
// POST /v1/loans/applications/:id/accept-offer
func (h *LoanHandler) AcceptOffer(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "accept_offer"),
	)

	applicationID := c.Param("id")
	if applicationID == "" {
		logger.Warn("Missing application ID")
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	var req domain.AcceptOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request format", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	// TODO: Implement offer acceptance workflow initiation
	// For now, return a placeholder response
	logger.Info("Offer acceptance workflow initiated",
		zap.String("application_id", applicationID),
		zap.String("offer_id", req.OfferID))

	middleware.CreateSuccessResponse(c, gin.H{
		"message":        "Offer acceptance workflow initiated",
		"status":         "pending",
		"application_id": applicationID,
		"offer_id":       req.OfferID,
	}, "OFFER_ACCEPTED", nil)
}

// TransitionState transitions an application state (admin endpoint)
// POST /v1/loans/applications/:id/transition
func (h *LoanHandler) TransitionState(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "transition_state"),
	)

	applicationID := c.Param("id")
	if applicationID == "" {
		logger.Warn("Missing application ID")
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	var req struct {
		FromState string `json:"from_state" binding:"required"`
		ToState   string `json:"to_state" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request format", zap.Error(err))
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	// Validate state values
	_ = domain.ApplicationState(req.FromState)
	_ = domain.ApplicationState(req.ToState)

	// TODO: Implement state transition workflow initiation
	// For now, return a placeholder response
	logger.Info("State transition workflow initiated",
		zap.String("application_id", applicationID),
		zap.String("from_state", req.FromState),
		zap.String("to_state", req.ToState))

	middleware.CreateSuccessResponse(c, gin.H{
		"message":        "State transition workflow initiated",
		"status":         "pending",
		"application_id": applicationID,
		"from_state":     req.FromState,
		"to_state":       req.ToState,
	}, "STATE_TRANSITION_SUCCESS", nil)
}

// GetApplicationStats gets application statistics (admin endpoint)
// @Summary Get application statistics
// @Description Retrieve statistics about loan applications (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param state query string false "Filter by state"
// @Param days query int false "Number of days to look back (default: 30)"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=map[string]interface{}} "Statistics retrieved successfully"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 403 {object} middleware.ErrorResponse "Forbidden"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/stats [get]
func (h *LoanHandler) GetApplicationStats(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "get_application_stats"),
	)

	// Get query parameters for filtering
	status := c.Query("status")
	state := c.Query("state")
	daysStr := c.Query("days")

	days := 30 // Default to 30 days
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	// TODO: Implement real database query for application statistics
	// This would aggregate data from the applications table
	middleware.CreateErrorResponse(c, http.StatusNotImplemented, domain.LOAN_023, map[string]interface{}{
		"message": "Application statistics not implemented - database repository required",
	})

	logger.Info("Application statistics requested but not implemented",
		zap.Int("days", days),
		zap.String("status_filter", status),
		zap.String("state_filter", state))
}

// Health check endpoint
// @Summary Health check
// @Description Check the health status of the loan service
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} middleware.SuccessResponse{data=map[string]interface{}} "Service is healthy"
// @Router /health [get]
func (h *LoanHandler) Health(c *gin.Context) {
	health := gin.H{
		"status":  "healthy",
		"service": "loan-service",
		"version": "v1.0.0",
		"timestamp": gin.H{
			"unix": gin.H{
				"seconds": gin.H{
					"value": "1692032400",
				},
			},
		},
	}

	middleware.CreateSuccessResponse(c, health, "", nil)
}

// GetWorkflowStatus gets the status of a workflow
// @Summary Get workflow status
// @Description Retrieve the current status of a workflow execution
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=map[string]interface{}} "Workflow status retrieved successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid workflow ID"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 404 {object} middleware.ErrorResponse "Workflow not found"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /workflows/{id}/status [get]
func (h *LoanHandler) GetWorkflowStatus(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	// TODO: Implement real workflow status retrieval
	// This would call the workflow service to get actual status
	middleware.CreateErrorResponse(c, http.StatusNotImplemented, domain.LOAN_023, map[string]interface{}{
		"message": "Workflow status retrieval not implemented - workflow service required",
	})

}

// PauseWorkflow pauses a running workflow
// @Summary Pause workflow
// @Description Pause a running workflow execution
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=map[string]interface{}} "Workflow paused successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid workflow ID"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 404 {object} middleware.ErrorResponse "Workflow not found"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /workflows/{id}/pause [post]
func (h *LoanHandler) PauseWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	logger := h.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "pause_workflow"),
	)

	logger.Info("Workflow pause requested")

	// For now, just return success
	// In a real implementation, this would call the workflow service
	middleware.CreateSuccessResponse(c, gin.H{"status": "paused"}, "WORKFLOW_PAUSED", nil)
}

// ResumeWorkflow resumes a paused workflow
// @Summary Resume workflow
// @Description Resume a paused workflow execution
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=map[string]interface{}} "Workflow resumed successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid workflow ID"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 404 {object} middleware.ErrorResponse "Workflow not found"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /workflows/{id}/resume [post]
func (h *LoanHandler) ResumeWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	logger := h.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "resume_workflow"),
	)

	logger.Info("Workflow resume requested")

	// For now, just return success
	// In a real implementation, this would call the workflow service
	middleware.CreateSuccessResponse(c, gin.H{"status": "resumed"}, "WORKFLOW_RESUMED", nil)
}

// TerminateWorkflow terminates a running workflow
// @Summary Terminate workflow
// @Description Terminate a running workflow execution
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID"
// @Param reason body map[string]string true "Termination reason"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} middleware.SuccessResponse{data=map[string]interface{}} "Workflow terminated successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid workflow ID"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 404 {object} middleware.ErrorResponse "Workflow not found"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /workflows/{id}/terminate [post]
func (h *LoanHandler) TerminateWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.CreateErrorResponse(c, http.StatusBadRequest, domain.LOAN_020, nil)
		return
	}

	reason := req["reason"]
	if reason == "" {
		reason = "User requested termination"
	}

	logger := h.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("reason", reason),
		zap.String("operation", "terminate_workflow"),
	)

	logger.Info("Workflow termination requested")

	// For now, just return success
	// In a real implementation, this would call the workflow service
	middleware.CreateSuccessResponse(c, gin.H{"status": "terminated"}, "WORKFLOW_TERMINATED", nil)
}

// DocumentUploadRequest represents a document upload request
type DocumentUploadRequest struct {
	ApplicationID string                 `json:"applicationId" validate:"required"`
	UserID        string                 `json:"userId" validate:"required"`
	DocumentType  string                 `json:"documentType" validate:"required"`
	FileName      string                 `json:"fileName" validate:"required"`
	FileSize      int64                  `json:"fileSize" validate:"required,min=1"`
	ContentType   string                 `json:"contentType" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentUploadResponse represents a document upload response
type DocumentUploadResponse struct {
	Success          bool                   `json:"success"`
	DocumentID       string                 `json:"documentId,omitempty"`
	UploadedAt       time.Time              `json:"uploadedAt,omitempty"`
	ValidationStatus string                 `json:"validationStatus,omitempty"`
	Errors           []string               `json:"errors,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentCollectionStatus represents the status of document collection
type DocumentCollectionStatus struct {
	ApplicationID         string                  `json:"applicationId"`
	UserID                string                  `json:"userId"`
	Status                string                  `json:"status"`
	TotalRequired         int                     `json:"totalRequired"`
	Collected             int                     `json:"collected"`
	Pending               int                     `json:"pending"`
	Documents             map[string]DocumentInfo `json:"documents"`
	CollectionStarted     time.Time               `json:"collectionStarted"`
	CollectionCompletedAt *time.Time              `json:"collectionCompletedAt,omitempty"`
	ValidationErrors      map[string][]string     `json:"validationErrors,omitempty"`
}

// DocumentInfo represents information about a specific document
type DocumentInfo struct {
	DocumentType string                 `json:"documentType"`
	Collected    bool                   `json:"collected"`
	Validated    bool                   `json:"validated"`
	FileName     string                 `json:"fileName,omitempty"`
	FileSize     int64                  `json:"fileSize,omitempty"`
	UploadedAt   *time.Time             `json:"uploadedAt,omitempty"`
	ValidatedAt  *time.Time             `json:"validatedAt,omitempty"`
	Errors       []string               `json:"errors,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UploadDocument handles document upload for loan applications
// @Summary Upload a document for a loan application
// @Description Upload a document file for a specific loan application
// @Tags Documents
// @Accept json
// @Produce json
// @Param document body DocumentUploadRequest true "Document upload details"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} DocumentUploadResponse "Document uploaded successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid request data"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/documents/upload [post]
func (h *LoanHandler) UploadDocument(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "upload_document"))

	// TODO: Implement document upload workflow initiation
	// For now, return a placeholder response
	logger.Info("Document upload workflow initiated")

	middleware.CreateSuccessResponse(c, gin.H{
		"message": "Document upload workflow initiated",
		"status":  "pending",
	}, "DOCUMENT_UPLOAD_SUCCESS", nil)
}

// GetDocumentCollectionStatus retrieves the status of document collection for an application
// @Summary Get document collection status
// @Description Retrieve the current status of document collection for a loan application
// @Tags Documents
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param userId query string true "User ID"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} DocumentCollectionStatus "Document collection status retrieved successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid request data"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 404 {object} middleware.ErrorResponse "Application not found"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/applications/{id}/documents/status [get]
func (h *LoanHandler) GetDocumentCollectionStatus(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "get_document_collection_status"))

	applicationID := c.Param("id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	logger.Info("Retrieving document collection status",
		zap.String("application_id", applicationID),
		zap.String("user_id", userID))

	// TODO: Implement document collection status retrieval
	// For now, return a placeholder response
	logger.Info("Document collection status requested",
		zap.String("application_id", applicationID),
		zap.String("user_id", userID))

	// Return placeholder status
	placeholderStatus := gin.H{
		"application_id": applicationID,
		"user_id":        userID,
		"status":         "pending",
		"total_required": 3,
		"collected":      0,
		"pending":        3,
	}

	middleware.CreateSuccessResponse(c, placeholderStatus, "DOCUMENT_STATUS_RETRIEVED", nil)
}

// CompleteDocumentCollection marks document collection as completed
// @Summary Complete document collection
// @Description Mark the document collection process as completed for a loan application
// @Tags Documents
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param request body object{userId=string,force=bool} true "Completion request"
// @Param X-Language header string false "Language preference (en, vi)"
// @Success 200 {object} object{success=bool,message=string} "Document collection completed successfully"
// @Failure 400 {object} middleware.ErrorResponse "Invalid request data"
// @Failure 401 {object} middleware.ErrorResponse "Unauthorized"
// @Failure 404 {object} middleware.ErrorResponse "Application not found"
// @Failure 500 {object} middleware.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /loans/applications/{id}/documents/complete [post]
func (h *LoanHandler) CompleteDocumentCollection(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "complete_document_collection"))

	applicationID := c.Param("id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	var req struct {
		UserID string `json:"userId" validate:"required"`
		Force  bool   `json:"force,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Completing document collection",
		zap.String("application_id", applicationID),
		zap.String("user_id", req.UserID),
		zap.Bool("force", req.Force))

	// TODO: Implement document collection completion workflow initiation
	// For now, return a placeholder response
	logger.Info("Document collection completion workflow initiated",
		zap.String("application_id", applicationID),
		zap.String("user_id", req.UserID),
		zap.Bool("force", req.Force))

	middleware.CreateSuccessResponse(c, gin.H{
		"message":        "Document collection completion workflow initiated",
		"status":         "pending",
		"application_id": applicationID,
		"user_id":        req.UserID,
	}, "DOCUMENT_COLLECTION_COMPLETED", nil)
}

// getFieldErrors extracts field-specific errors from validation errors
func getFieldErrors(err error) map[string]string {
	fieldErrors := make(map[string]string)

	// Parse common validation error patterns
	if err != nil {
		errStr := err.Error()

		// Handle common Gin validation error patterns
		if strings.Contains(errStr, "parsing time") {
			fieldErrors["date_format"] = "Date must be in ISO 8601 format (e.g., 1990-01-01T00:00:00Z)"
		}
		if strings.Contains(errStr, "required") {
			fieldErrors["missing_fields"] = "Required fields are missing"
		}
		if strings.Contains(errStr, "email") {
			fieldErrors["email"] = "Invalid email format"
		}
		if strings.Contains(errStr, "len") {
			fieldErrors["field_length"] = "Field length validation failed"
		}
		if strings.Contains(errStr, "min") {
			fieldErrors["min_value"] = "Value is below minimum requirement"
		}
		if strings.Contains(errStr, "max") {
			fieldErrors["max_value"] = "Value exceeds maximum limit"
		}
	}

	return fieldErrors
}

// getRequestBody safely extracts the request body for error reporting
func getRequestBody(c *gin.Context) string {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "Unable to read request body"
	}

	// Restore the body for potential re-reading
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Return a safe version of the body (truncated if too long)
	if len(body) > 500 {
		return string(body[:500]) + "... (truncated)"
	}
	return string(body)
}

// RegisterRoutes registers all loan service routes
func (h *LoanHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes
	router.GET("/health", h.Health)

	// Protected routes (require authentication)
	loans := router.Group("/loans")
	{
		// Application management
		loans.POST("/applications", h.CreateApplication)
		loans.GET("/applications", h.GetApplicationsByUser)
		loans.GET("/applications/:id", h.GetApplication)
		loans.PUT("/applications/:id", h.UpdateApplication)
		loans.POST("/applications/:id/submit", h.SubmitApplication)

		// Pre-qualification
		loans.POST("/prequalify", h.PreQualify)

		// Offers
		loans.POST("/applications/:id/offer", h.GenerateOffer)
		loans.POST("/applications/:id/accept-offer", h.AcceptOffer)

		// Admin endpoints (would typically require admin role)
		loans.POST("/applications/:id/transition", h.TransitionState)
		loans.GET("/stats", h.GetApplicationStats)

		// Document management
		loans.POST("/documents/upload", h.UploadDocument)
		loans.GET("/applications/:id/documents/status", h.GetDocumentCollectionStatus)
		loans.POST("/applications/:id/documents/complete", h.CompleteDocumentCollection)
	}

	// Workflow management routes
	workflows := router.Group("/workflows")
	{
		workflows.GET("/:id/status", h.GetWorkflowStatus)
		workflows.POST("/:id/pause", h.PauseWorkflow)
		workflows.POST("/:id/resume", h.ResumeWorkflow)
		workflows.POST("/:id/terminate", h.TerminateWorkflow)
	}
}
