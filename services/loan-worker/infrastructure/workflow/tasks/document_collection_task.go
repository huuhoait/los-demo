package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DocumentCollectionTaskHandler handles document collection tasks
type DocumentCollectionTaskHandler struct {
	logger *zap.Logger
}

// NewDocumentCollectionTaskHandler creates a new document collection task handler
func NewDocumentCollectionTaskHandler(logger *zap.Logger) *DocumentCollectionTaskHandler {
	return &DocumentCollectionTaskHandler{
		logger: logger,
	}
}

// DocumentResult represents the result of document processing
type DocumentResult struct {
	Collected    bool                   `json:"collected"`
	Validated    bool                   `json:"validated"`
	DocumentType string                 `json:"documentType"`
	FileName     string                 `json:"fileName,omitempty"`
	FileSize     int64                  `json:"fileSize,omitempty"`
	UploadedAt   time.Time              `json:"uploadedAt,omitempty"`
	ValidatedAt  time.Time              `json:"validatedAt,omitempty"`
	Errors       []string               `json:"errors,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Execute handles the collection and validation of required loan documents
func (h *DocumentCollectionTaskHandler) Execute(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "document_collection"))

	logger.Info("Document collection task initiated",
		zap.Any("input", input))

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	userID, _ := input["userId"].(string)
	requiredDocuments, _ := input["requiredDocuments"].([]interface{})

	// Validate required fields
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if len(requiredDocuments) == 0 {
		return nil, fmt.Errorf("required documents list is empty")
	}

	logger.Info("Processing document collection",
		zap.String("application_id", applicationID),
		zap.String("user_id", userID),
		zap.Int("required_documents_count", len(requiredDocuments)))

	// Convert required documents to string slice
	requiredDocs := make([]string, len(requiredDocuments))
	for i, doc := range requiredDocuments {
		if docStr, ok := doc.(string); ok {
			requiredDocs[i] = docStr
		}
	}

	// Process document collection
	documentResults := h.processDocumentCollection(ctx, applicationID, userID, requiredDocs)

	// Determine overall collection status
	allDocumentsCollected := true
	collectionCompletedAt := time.Now()

	for _, result := range documentResults {
		if !result.Collected {
			allDocumentsCollected = false
			collectionCompletedAt = time.Time{} // Reset if not all collected
			break
		}
	}

	// Prepare output
	output := map[string]interface{}{
		"documentsCollected":     allDocumentsCollected,
		"incomeVerification":     documentResults["income_verification"].Collected,
		"employmentVerification": documentResults["employment_verification"].Collected,
		"bankStatements":         documentResults["bank_statements"].Collected,
		"identificationDocument": documentResults["identification"].Collected,
		"collectionCompletedAt":  collectionCompletedAt,
		"status":                 h.getCollectionStatus(allDocumentsCollected),
		"documentDetails":        documentResults,
		"validationSummary": map[string]interface{}{
			"totalRequired":    len(requiredDocs),
			"collected":        h.countCollectedDocuments(documentResults),
			"pending":          h.countPendingDocuments(documentResults),
			"validationErrors": h.getValidationErrors(documentResults),
		},
	}

	logger.Info("Document collection completed",
		zap.Bool("all_documents_collected", allDocumentsCollected),
		zap.Int("collected_count", h.countCollectedDocuments(documentResults)),
		zap.Int("pending_count", h.countPendingDocuments(documentResults)))

	return output, nil
}

// ExecuteHumanTask handles the document collection as a HUMAN task
// This method is specifically for human intervention scenarios
func (h *DocumentCollectionTaskHandler) ExecuteHumanTask(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "document_collection_human"))

	logger.Info("Document collection HUMAN task initiated",
		zap.Any("input", input))

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	userID, _ := input["userId"].(string)
	requiredDocuments, _ := input["requiredDocuments"].([]interface{})

	// Validate required fields
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if len(requiredDocuments) == 0 {
		return nil, fmt.Errorf("required documents list is empty")
	}

	logger.Info("Document collection HUMAN task requires manual intervention",
		zap.String("application_id", applicationID),
		zap.String("user_id", userID),
		zap.Int("required_documents_count", len(requiredDocuments)))

	// For HUMAN tasks, we don't process documents automatically
	// Instead, we set up the task for human operators to handle
	logger.Info("Setting up document collection task for human processing")

	// Create initial status indicating human intervention is needed
	output := map[string]interface{}{
		"taskType":                  "document_collection",
		"taskStatus":                "PENDING_HUMAN_ACTION",
		"humanInterventionRequired": true,
		"applicationId":             applicationID,
		"userId":                    userID,
		"requiredDocuments":         requiredDocuments,
		"message":                   "Document collection requires manual processing by loan officer",
		"nextSteps": []string{
			"1. Review required documents list",
			"2. Contact applicant for document submission",
			"3. Verify document authenticity",
			"4. Update task status when complete",
		},
		"estimatedProcessingTime": "24 hours",
		"createdAt":               time.Now().UTC().Format(time.RFC3339),
		"status":                  "pending_human_action",
	}

	logger.Info("Document collection HUMAN task setup completed",
		zap.String("application_id", applicationID),
		zap.String("status", "PENDING_HUMAN_ACTION"))

	return output, nil
}

// processDocumentCollection processes the collection of required documents
func (h *DocumentCollectionTaskHandler) processDocumentCollection(
	ctx context.Context,
	applicationID string,
	userID string,
	requiredDocs []string,
) map[string]DocumentResult {
	logger := h.logger.With(
		zap.String("application_id", applicationID),
		zap.String("user_id", userID),
		zap.String("operation", "process_document_collection"))

	results := make(map[string]DocumentResult)

	// Initialize results for all required documents
	for _, docType := range requiredDocs {
		results[docType] = DocumentResult{
			Collected:    false,
			Validated:    false,
			DocumentType: docType,
			Errors:       []string{},
			Metadata:     make(map[string]interface{}),
		}
	}

	// Simulate document collection process
	// In a real implementation, this would:
	// 1. Check if documents are uploaded to storage
	// 2. Validate document formats and content
	// 3. Verify document authenticity
	// 4. Store document metadata

	for docType := range results {
		// Simulate document processing with realistic delays
		time.Sleep(100 * time.Millisecond)

		// Simulate different collection scenarios
		switch docType {
		case "income_verification":
			results[docType] = h.processIncomeVerification(ctx, applicationID, userID)
		case "employment_verification":
			results[docType] = h.processEmploymentVerification(ctx, applicationID, userID)
		case "bank_statements":
			results[docType] = h.processBankStatements(ctx, applicationID, userID)
		case "identification":
			results[docType] = h.processIdentificationDocument(ctx, applicationID, userID)
		default:
			result := results[docType]
			result.Errors = append(result.Errors,
				fmt.Sprintf("Unknown document type: %s", docType))
			results[docType] = result
		}
	}

	logger.Info("Document processing completed",
		zap.Int("total_documents", len(results)),
		zap.Int("collected", h.countCollectedDocuments(results)),
		zap.Int("validated", h.countValidatedDocuments(results)))

	return results
}

// processIncomeVerification processes income verification documents
func (h *DocumentCollectionTaskHandler) processIncomeVerification(
	ctx context.Context,
	applicationID string,
	userID string,
) DocumentResult {
	// Simulate income verification document processing
	// In real implementation, this would:
	// - Check for pay stubs, W2 forms, tax returns
	// - Validate income amounts against application data
	// - Verify document authenticity and recency

	result := DocumentResult{
		Collected:    true, // Simulate successful collection
		Validated:    true,
		DocumentType: "income_verification",
		FileName:     fmt.Sprintf("income_verification_%s.pdf", applicationID),
		FileSize:     2048576, // 2MB
		UploadedAt:   time.Now().Add(-2 * time.Hour),
		ValidatedAt:  time.Now().Add(-1 * time.Hour),
		Metadata: map[string]interface{}{
			"documentType": "pay_stub",
			"employer":     "Tech Corp Inc",
			"period":       "2025-08-01 to 2025-08-15",
			"grossIncome":  5000.00,
			"netIncome":    3800.00,
		},
	}

	// Simulate validation logic
	if result.Metadata["grossIncome"].(float64) < 1000 {
		result.Validated = false
		result.Errors = append(result.Errors, "Income below minimum threshold")
	}

	return result
}

// processEmploymentVerification processes employment verification documents
func (h *DocumentCollectionTaskHandler) processEmploymentVerification(
	ctx context.Context,
	applicationID string,
	userID string,
) DocumentResult {
	// Simulate employment verification document processing
	result := DocumentResult{
		Collected:    true,
		Validated:    true,
		DocumentType: "employment_verification",
		FileName:     fmt.Sprintf("employment_verification_%s.pdf", applicationID),
		FileSize:     1536000, // 1.5MB
		UploadedAt:   time.Now().Add(-3 * time.Hour),
		ValidatedAt:  time.Now().Add(-2 * time.Hour),
		Metadata: map[string]interface{}{
			"documentType":   "employment_letter",
			"employer":       "Tech Corp Inc",
			"position":       "Software Engineer",
			"startDate":      "2023-01-15",
			"employmentType": "full_time",
			"verifiedBy":     "HR Department",
		},
	}

	// Simulate validation logic
	startDate, _ := time.Parse("2006-01-02", result.Metadata["startDate"].(string))
	if time.Since(startDate) < 6*30*24*time.Hour { // Less than 6 months
		result.Validated = false
		result.Errors = append(result.Errors, "Employment duration below minimum requirement")
	}

	return result
}

// processBankStatements processes bank statement documents
func (h *DocumentCollectionTaskHandler) processBankStatements(
	ctx context.Context,
	applicationID string,
	userID string,
) DocumentResult {
	// Simulate bank statement processing
	result := DocumentResult{
		Collected:    true,
		Validated:    true,
		DocumentType: "bank_statements",
		FileName:     fmt.Sprintf("bank_statements_%s.pdf", applicationID),
		FileSize:     3072000, // 3MB
		UploadedAt:   time.Now().Add(-4 * time.Hour),
		ValidatedAt:  time.Now().Add(-3 * time.Hour),
		Metadata: map[string]interface{}{
			"documentType":     "bank_statement",
			"bankName":         "Chase Bank",
			"accountType":      "checking",
			"statementPeriod":  "2025-07-01 to 2025-07-31",
			"averageBalance":   8500.00,
			"minimumBalance":   3200.00,
			"transactionCount": 45,
		},
	}

	// Simulate validation logic
	avgBalance := result.Metadata["averageBalance"].(float64)
	if avgBalance < 1000 {
		result.Validated = false
		result.Errors = append(result.Errors, "Insufficient average balance")
	}

	return result
}

// processIdentificationDocument processes identification documents
func (h *DocumentCollectionTaskHandler) processIdentificationDocument(
	ctx context.Context,
	applicationID string,
	userID string,
) DocumentResult {
	// Simulate identification document processing
	result := DocumentResult{
		Collected:    true,
		Validated:    true,
		DocumentType: "identification",
		FileName:     fmt.Sprintf("identification_%s.jpg", applicationID),
		FileSize:     512000, // 512KB
		UploadedAt:   time.Now().Add(-1 * time.Hour),
		ValidatedAt:  time.Now().Add(-30 * time.Minute),
		Metadata: map[string]interface{}{
			"documentType":       "drivers_license",
			"state":              "CA",
			"licenseNumber":      "A123456789",
			"expiryDate":         "2028-05-15",
			"dateOfBirth":        "1990-03-15",
			"verificationMethod": "OCR_verification",
		},
	}

	// Simulate validation logic
	expiryDate, _ := time.Parse("2006-01-02", result.Metadata["expiryDate"].(string))
	if expiryDate.Before(time.Now().Add(6 * 30 * 24 * time.Hour)) { // Expires within 6 months
		result.Validated = false
		result.Errors = append(result.Errors, "Identification document expires soon")
	}

	return result
}

// Helper methods for document processing
func (h *DocumentCollectionTaskHandler) getCollectionStatus(allCollected bool) string {
	if allCollected {
		return "completed"
	}
	return "pending_human_action"
}

func (h *DocumentCollectionTaskHandler) countCollectedDocuments(results map[string]DocumentResult) int {
	count := 0
	for _, result := range results {
		if result.Collected {
			count++
		}
	}
	return count
}

func (h *DocumentCollectionTaskHandler) countValidatedDocuments(results map[string]DocumentResult) int {
	count := 0
	for _, result := range results {
		if result.Validated {
			count++
		}
	}
	return count
}

func (h *DocumentCollectionTaskHandler) countPendingDocuments(results map[string]DocumentResult) int {
	count := 0
	for _, result := range results {
		if !result.Collected {
			count++
		}
	}
	return count
}

func (h *DocumentCollectionTaskHandler) getValidationErrors(results map[string]DocumentResult) map[string][]string {
	errors := make(map[string][]string)
	for docType, result := range results {
		if len(result.Errors) > 0 {
			errors[docType] = result.Errors
		}
	}
	return errors
}

// GetHumanTaskInstructions returns instructions for human operators
func (h *DocumentCollectionTaskHandler) GetHumanTaskInstructions() map[string]interface{} {
	return map[string]interface{}{
		"taskType": "document_collection",
		"instructions": []string{
			"Review the loan application and identify required documents",
			"Contact the applicant to request missing documents",
			"Verify document authenticity and completeness",
			"Upload documents to the system",
			"Mark the task as complete when all documents are collected",
		},
		"requiredDocuments": []string{
			"income_verification",
			"employment_verification",
			"bank_statements",
			"identification",
		},
		"estimatedTime": "2-4 hours",
		"priority":      "High",
		"assignedTo":    "Loan Officer",
	}
}

// GetEstimatedProcessingTime returns the estimated processing time for human tasks
func (h *DocumentCollectionTaskHandler) GetEstimatedProcessingTime() string {
	return "24 hours"
}
