package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ValidateApplicationTaskHandler handles application validation tasks
type ValidateApplicationTaskHandler struct {
	logger *zap.Logger
}

// NewValidateApplicationTaskHandler creates a new validate application task handler
func NewValidateApplicationTaskHandler(logger *zap.Logger) *ValidateApplicationTaskHandler {
	return &ValidateApplicationTaskHandler{
		logger: logger,
	}
}

// Execute validates loan application data and business rules
func (h *ValidateApplicationTaskHandler) Execute(
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
		isValidPurpose := false
		for _, purpose := range validPurposes {
			if loanPurpose == purpose {
				isValidPurpose = true
				break
			}
		}
		if !isValidPurpose {
			errors["loanPurpose"] = fmt.Sprintf("Invalid loan purpose. Must be one of: %v", validPurposes)
		}
	}

	if annualIncome <= 0 {
		errors["annualIncome"] = "Annual income must be greater than 0"
	} else if annualIncome < 20000 {
		errors["annualIncome"] = "Annual income must be at least $20,000"
	}

	if monthlyIncome <= 0 {
		errors["monthlyIncome"] = "Monthly income must be greater than 0"
	} else if monthlyIncome < 1667 {
		errors["monthlyIncome"] = "Monthly income must be at least $1,667"
	}

	if requestedTerm <= 0 {
		errors["requestedTerm"] = "Requested term must be greater than 0"
	} else if requestedTerm < 12 {
		errors["requestedTerm"] = "Requested term must be at least 12 months"
	} else if requestedTerm > 84 {
		errors["requestedTerm"] = "Requested term cannot exceed 84 months"
	}

	// Calculate debt-to-income ratio
	var dtiRatio float64
	if monthlyIncome > 0 {
		// Assume monthly debt payments are 10% of monthly income for demo
		monthlyDebt := monthlyIncome * 0.1
		dtiRatio = (monthlyDebt / monthlyIncome) * 100
	}

	// Business rule: DTI ratio should be less than 43%
	if dtiRatio > 43 {
		errors["debtToIncomeRatio"] = fmt.Sprintf("Debt-to-income ratio (%.1f%%) exceeds maximum allowed (43%%)", dtiRatio)
	}

	// Prepare normalized data
	normalizedData = map[string]interface{}{
		"applicationId": applicationID,
		"userId":        userID,
		"loanAmount":    loanAmount,
		"loanPurpose":   loanPurpose,
		"annualIncome":  annualIncome,
		"monthlyIncome": monthlyIncome,
		"requestedTerm": int(requestedTerm),
		"dtiRatio":      dtiRatio,
		"validatedAt":   time.Now(),
	}

	// Determine overall validation result
	isValid := len(errors) == 0

	logger.Info("Application validation completed",
		zap.Bool("is_valid", isValid),
		zap.Int("error_count", len(errors)),
		zap.Float64("dti_ratio", dtiRatio))

	return map[string]interface{}{
		"valid":            isValid,
		"validationErrors": errors,
		"normalizedData":   normalizedData,
	}, nil
}
