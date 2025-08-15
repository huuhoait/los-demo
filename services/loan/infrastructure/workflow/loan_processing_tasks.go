package workflow

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"loan-service/pkg/i18n"
)

// LoanProcessingTaskHandler handles loan processing workflow tasks
type LoanProcessingTaskHandler struct {
	logger    *zap.Logger
	localizer *i18n.Localizer
}

// Execute implements the TaskHandler interface
func (h *LoanProcessingTaskHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract the task type from the input to determine which method to call
	taskType, _ := input["taskType"].(string)
	referenceTaskName, _ := input["referenceTaskName"].(string)

	logger := h.logger.With(
		zap.String("task_type", taskType),
		zap.String("reference_task_name", referenceTaskName),
		zap.String("operation", "execute_loan_processing_task"),
	)

	logger.Info("Executing loan processing task")

	// Determine which method to call based on the task type
	switch taskType {
	case "validate_application":
		return h.ValidateApplication(ctx, input)
	case "update_application_state":
		return h.UpdateApplicationState(ctx, input)
	case "document_collection":
		return h.DocumentCollection(ctx, input)
	case "identity_verification":
		return h.IdentityVerification(ctx, input)
	case "finalize_loan_decision":
		return h.FinalizeLoanDecision(ctx, input)
	default:
		// Fallback: try to determine task type from input structure
		logger.Warn("Unknown task type, attempting to determine from input structure", zap.String("task_type", taskType))

		if _, hasLoanAmount := input["loanAmount"]; hasLoanAmount {
			return h.ValidateApplication(ctx, input)
		}

		if _, hasFromState := input["fromState"]; hasFromState {
			return h.UpdateApplicationState(ctx, input)
		}

		if _, hasRequiredDocuments := input["requiredDocuments"]; hasRequiredDocuments {
			return h.DocumentCollection(ctx, input)
		}

		if _, hasIdentificationDocument := input["identificationDocument"]; hasIdentificationDocument {
			return h.IdentityVerification(ctx, input)
		}

		if _, hasDecision := input["decision"]; hasDecision {
			return h.FinalizeLoanDecision(ctx, input)
		}

		return nil, fmt.Errorf("unknown task type: %s", taskType)
	}
}

// NewLoanProcessingTaskHandler creates a new loan processing task handler
func NewLoanProcessingTaskHandler(logger *zap.Logger, localizer *i18n.Localizer) *LoanProcessingTaskHandler {
	return &LoanProcessingTaskHandler{
		logger:    logger,
		localizer: localizer,
	}
}

// ValidateApplication validates loan application data and business rules
func (h *LoanProcessingTaskHandler) ValidateApplication(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "validate_application"))

	logger.Info("Starting application validation",
		zap.Any("input_data", input),
		zap.Int("input_keys", len(input)))

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	userID, _ := input["userId"].(string)
	loanAmount, _ := input["loanAmount"].(float64)
	loanPurpose, _ := input["loanPurpose"].(string)
	annualIncome, _ := input["annualIncome"].(float64)
	monthlyIncome, _ := input["monthlyIncome"].(float64)
	requestedTerm, _ := input["requestedTerm"].(float64)

	logger.Info("Extracted input parameters",
		zap.String("application_id", applicationID),
		zap.String("user_id", userID),
		zap.Float64("loan_amount", loanAmount),
		zap.String("loan_purpose", loanPurpose),
		zap.Float64("annual_income", annualIncome),
		zap.Float64("monthly_income", monthlyIncome),
		zap.Float64("requested_term", requestedTerm))

	// Validate required fields
	errors := make(map[string]string)
	normalizedData := make(map[string]interface{})

	if applicationID == "" {
		errors["applicationId"] = "Application ID is required"
	}

	if userID == "" {
		errors["userId"] = "User ID is required"
	}

	if loanAmount <= 0 {
		errors["loanAmount"] = "Loan amount must be greater than 0"
	} else if loanAmount < 5000 {
		errors["loanAmount"] = "Loan amount must be at least $5,000"
	} else if loanAmount > 50000 {
		errors["loanAmount"] = "Loan amount cannot exceed $50,000"
	}

	if loanPurpose == "" {
		errors["loanPurpose"] = "Loan purpose is required"
	} else {
		// Validate loan purpose enum
		validPurposes := []string{"debt_consolidation", "home_improvement", "major_purchase", "medical_expenses", "education", "other"}
		valid := false
		for _, purpose := range validPurposes {
			if purpose == loanPurpose {
				valid = true
				break
			}
		}
		if !valid {
			errors["loanPurpose"] = "Invalid loan purpose"
		}
	}

	if annualIncome <= 0 {
		errors["annualIncome"] = "Annual income must be greater than 0"
	} else if annualIncome < 25000 {
		errors["annualIncome"] = "Annual income must be at least $25,000"
	}

	if monthlyIncome <= 0 {
		errors["monthlyIncome"] = "Monthly income must be greater than 0"
	}

	if requestedTerm <= 0 {
		errors["requestedTerm"] = "Requested term must be greater than 0"
	} else if requestedTerm < 12 {
		errors["requestedTerm"] = "Requested term must be at least 12 months"
	} else if requestedTerm > 84 {
		errors["requestedTerm"] = "Requested term cannot exceed 84 months"
	}

	// Business rule validation
	if len(errors) == 0 {
		// Check debt-to-income ratio (assuming monthly debt is available)
		monthlyDebt, hasMonthlyDebt := input["monthlyDebt"].(float64)
		if hasMonthlyDebt && monthlyIncome > 0 {
			dtiRatio := monthlyDebt / monthlyIncome
			if dtiRatio > 0.4 {
				errors["monthlyDebt"] = "Debt-to-income ratio exceeds 40%"
			}
		}

		// Check if loan amount is reasonable for income
		if annualIncome > 0 {
			maxLoanAmount := annualIncome * 2.0 // Simple rule: max 2x annual income
			if loanAmount > maxLoanAmount {
				errors["loanAmount"] = fmt.Sprintf("Loan amount exceeds maximum allowed (%.2f)", maxLoanAmount)
			}
		}
	}

	// Normalize data for output
	normalizedData["applicationId"] = applicationID
	normalizedData["userId"] = userID
	normalizedData["loanAmount"] = loanAmount
	normalizedData["loanPurpose"] = loanPurpose
	normalizedData["annualIncome"] = annualIncome
	normalizedData["monthlyIncome"] = monthlyIncome
	normalizedData["requestedTerm"] = int(requestedTerm)
	normalizedData["validatedAt"] = time.Now().Format(time.RFC3339)

	// Check if validation passed
	valid := len(errors) == 0

	logger.Info("Application validation completed",
		zap.Bool("valid", valid),
		zap.Int("error_count", len(errors)),
		zap.String("application_id", applicationID),
		zap.Any("validation_errors", errors),
		zap.Any("normalized_data", normalizedData))

	// Prepare output
	output := map[string]interface{}{
		"valid":            valid,
		"validationErrors": errors,
		"normalizedData":   normalizedData,
	}

	logger.Info("Returning validation output",
		zap.Any("output", output),
		zap.Int("output_keys", len(output)))

	return output, nil
}

// UpdateApplicationState updates loan application state in database
func (h *LoanProcessingTaskHandler) UpdateApplicationState(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "update_application_state"))

	logger.Info("Updating application state")

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	fromState, _ := input["fromState"].(string)
	toState, _ := input["toState"].(string)
	reason, _ := input["reason"].(string)

	// Validate required fields
	errors := make(map[string]string)

	if applicationID == "" {
		errors["applicationId"] = "Application ID is required"
	}

	if fromState == "" {
		errors["fromState"] = "From state is required"
	}

	if toState == "" {
		errors["toState"] = "To state is required"
	}

	if reason == "" {
		errors["reason"] = "Reason for state change is required"
	}

	if len(errors) > 0 {
		return map[string]interface{}{
			"success": false,
			"errors":  errors,
		}, nil
	}

	// Simulate state update (in real implementation, this would update the database)
	updatedAt := time.Now()
	newState := toState

	logger.Info("Application state updated successfully",
		zap.String("application_id", applicationID),
		zap.String("from_state", fromState),
		zap.String("to_state", toState),
		zap.String("reason", reason),
	)

	return map[string]interface{}{
		"success":   true,
		"updatedAt": updatedAt.Format(time.RFC3339),
		"newState":  newState,
	}, nil
}

// DocumentCollection handles document collection task (human task)
func (h *LoanProcessingTaskHandler) DocumentCollection(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "document_collection"))

	logger.Info("Document collection task initiated")

	// This is a human task, so we just return the task status
	// In a real implementation, this would create a human task in the workflow engine
	return map[string]interface{}{
		"documentsCollected":     false,
		"incomeVerification":     false,
		"employmentVerification": false,
		"bankStatements":         false,
		"identificationDocument": false,
		"collectionCompletedAt":  nil,
		"status":                 "pending_human_action",
	}, nil
}

// IdentityVerification verifies customer identity using provided documents
func (h *LoanProcessingTaskHandler) IdentityVerification(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "identity_verification"))

	logger.Info("Performing identity verification")

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	userID, _ := input["userId"].(string)

	// Validate required fields
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Simulate identity verification process
	// In real implementation, this would integrate with identity verification services
	verificationScore := 85.0 // Simulated score
	verified := verificationScore >= 80.0

	logger.Info("Identity verification completed",
		zap.String("application_id", applicationID),
		zap.Bool("verified", verified),
		zap.Float64("verification_score", verificationScore),
	)

	return map[string]interface{}{
		"verified":          verified,
		"verificationScore": verificationScore,
		"personalInfo":      "verified",
		"ssn":               "verified",
		"verificationDetails": map[string]interface{}{
			"method": "document_verification",
			"score":  verificationScore,
			"status": "completed",
		},
	}, nil
}

// FinalizeLoanDecision finalizes loan processing decision and updates application
func (h *LoanProcessingTaskHandler) FinalizeLoanDecision(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "finalize_loan_decision"))

	logger.Info("Finalizing loan decision")

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	finalState, _ := input["finalState"].(string)
	decision, _ := input["decision"].(string)
	interestRate, _ := input["interestRate"].(float64)

	// Validate required fields
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	if finalState == "" {
		return nil, fmt.Errorf("final state is required")
	}

	if decision == "" {
		return nil, fmt.Errorf("decision is required")
	}

	// Simulate finalization process
	completedAt := time.Now()
	nextSteps := []string{}

	switch decision {
	case "approved":
		nextSteps = []string{"sign_loan_agreement", "provide_banking_info", "fund_disbursement"}
	case "denied":
		nextSteps = []string{"review_denial_reason", "consider_appeal", "alternative_products"}
	case "manual_review":
		nextSteps = []string{"underwriter_review", "additional_documentation", "follow_up"}
	default:
		nextSteps = []string{"contact_customer_service"}
	}

	logger.Info("Loan decision finalized",
		zap.String("application_id", applicationID),
		zap.String("decision", decision),
		zap.String("final_state", finalState),
		zap.Float64("interest_rate", interestRate),
	)

	return map[string]interface{}{
		"finalState":   finalState,
		"decision":     decision,
		"interestRate": interestRate,
		"completedAt":  completedAt.Format(time.RFC3339),
		"nextSteps":    nextSteps,
	}, nil
}
