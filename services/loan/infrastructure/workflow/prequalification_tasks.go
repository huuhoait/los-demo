package workflow

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"loan-service/pkg/i18n"
)

// PreQualificationTaskHandler handles pre-qualification workflow tasks
type PreQualificationTaskHandler struct {
	logger    *zap.Logger
	localizer *i18n.Localizer
}

// Execute implements the TaskHandler interface
func (h *PreQualificationTaskHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract the task name from the input to determine which method to call
	taskName, _ := input["taskName"].(string)

	switch taskName {
	case "validate_prequalify_input":
		return h.ValidatePreQualifyInput(ctx, input)
	case "calculate_dti_ratio":
		return h.CalculateDTIRatio(ctx, input)
	case "assess_prequalify_risk":
		return h.AssessPreQualifyRisk(ctx, input)
	case "generate_prequalify_terms":
		return h.GeneratePreQualifyTerms(ctx, input)
	case "finalize_prequalification":
		return h.FinalizePreQualification(ctx, input)
	case "update_application_state":
		return h.UpdateApplicationState(ctx, input)
	default:
		return nil, fmt.Errorf("unknown task type: %s", taskName)
	}
}

// NewPreQualificationTaskHandler creates a new pre-qualification task handler
func NewPreQualificationTaskHandler(logger *zap.Logger, localizer *i18n.Localizer) *PreQualificationTaskHandler {
	return &PreQualificationTaskHandler{
		logger:    logger,
		localizer: localizer,
	}
}

// ValidatePreQualifyInput validates the pre-qualification input parameters
func (h *PreQualificationTaskHandler) ValidatePreQualifyInput(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "validate_prequalify_input"))

	logger.Info("Validating pre-qualification input parameters")

	// Extract input parameters
	userID, _ := input["userId"].(string)
	loanAmount, _ := input["loanAmount"].(float64)
	annualIncome, _ := input["annualIncome"].(float64)
	monthlyDebt, _ := input["monthlyDebt"].(float64)
	employmentStatus, _ := input["employmentStatus"].(string)

	// Validate required fields
	errors := make(map[string]string)

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

	if annualIncome <= 0 {
		errors["annualIncome"] = "Annual income must be greater than 0"
	} else if annualIncome < 25000 {
		errors["annualIncome"] = "Annual income must be at least $25,000"
	}

	if monthlyDebt < 0 {
		errors["monthlyDebt"] = "Monthly debt cannot be negative"
	}

	if employmentStatus == "" {
		errors["employmentStatus"] = "Employment status is required"
	}

	// Check if validation passed
	valid := len(errors) == 0

	logger.Info("Input validation completed",
		zap.Bool("valid", valid),
		zap.Int("error_count", len(errors)),
	)

	return map[string]interface{}{
		"valid":            valid,
		"validationErrors": errors,
	}, nil
}

// CalculateDTIRatio calculates the debt-to-income ratio
func (h *PreQualificationTaskHandler) CalculateDTIRatio(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "calculate_dti_ratio"))

	logger.Info("Calculating debt-to-income ratio")

	// Extract input parameters
	annualIncome, _ := input["annualIncome"].(float64)
	monthlyDebt, _ := input["monthlyDebt"].(float64)

	// Calculate monthly income
	monthlyIncome := annualIncome / 12

	// Calculate DTI ratio
	var dtiRatio float64
	if monthlyIncome > 0 {
		dtiRatio = monthlyDebt / monthlyIncome
	} else {
		dtiRatio = 0
	}

	// Round to 4 decimal places
	dtiRatio = math.Round(dtiRatio*10000) / 10000

	logger.Info("DTI ratio calculated",
		zap.Float64("annual_income", annualIncome),
		zap.Float64("monthly_income", monthlyIncome),
		zap.Float64("monthly_debt", monthlyDebt),
		zap.Float64("dti_ratio", dtiRatio),
	)

	return map[string]interface{}{
		"dtiRatio":      dtiRatio,
		"monthlyIncome": monthlyIncome,
	}, nil
}

// AssessPreQualifyRisk performs initial risk assessment for pre-qualification
func (h *PreQualificationTaskHandler) AssessPreQualifyRisk(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "assess_prequalify_risk"))

	logger.Info("Performing pre-qualification risk assessment")

	// Extract input parameters
	loanAmount, _ := input["loanAmount"].(float64)
	annualIncome, _ := input["annualIncome"].(float64)
	employmentStatus, _ := input["employmentStatus"].(string)
	dtiRatio, _ := input["dtiRatio"].(float64)

	// Risk assessment logic
	riskLevel := "LOW"
	riskFactors := []string{}
	baseInterestRate := 8.0 // Base rate of 8%

	// Assess DTI ratio risk
	if dtiRatio > 0.43 {
		riskLevel = "HIGH"
		riskFactors = append(riskFactors, "High debt-to-income ratio")
		baseInterestRate += 3.0
	} else if dtiRatio > 0.36 {
		riskLevel = "MEDIUM"
		riskFactors = append(riskFactors, "Moderate debt-to-income ratio")
		baseInterestRate += 1.5
	}

	// Assess income risk
	if annualIncome < 35000 {
		riskLevel = "HIGH"
		riskFactors = append(riskFactors, "Low annual income")
		baseInterestRate += 2.0
	} else if annualIncome < 50000 {
		riskLevel = "MEDIUM"
		riskFactors = append(riskFactors, "Moderate annual income")
		baseInterestRate += 1.0
	}

	// Assess employment risk
	switch employmentStatus {
	case "unemployed":
		riskLevel = "HIGH"
		riskFactors = append(riskFactors, "Unemployed")
		baseInterestRate += 4.0
	case "part_time":
		riskLevel = "MEDIUM"
		riskFactors = append(riskFactors, "Part-time employment")
		baseInterestRate += 1.5
	case "self_employed":
		riskLevel = "MEDIUM"
		riskFactors = append(riskFactors, "Self-employed")
		baseInterestRate += 1.0
	}

	// Assess loan amount risk
	if loanAmount > annualIncome*0.8 {
		riskLevel = "HIGH"
		riskFactors = append(riskFactors, "High loan amount relative to income")
		baseInterestRate += 2.0
	}

	// Cap the interest rate
	if baseInterestRate > 25.0 {
		baseInterestRate = 25.0
	}

	logger.Info("Risk assessment completed",
		zap.String("risk_level", riskLevel),
		zap.Float64("base_interest_rate", baseInterestRate),
		zap.Int("risk_factor_count", len(riskFactors)),
	)

	return map[string]interface{}{
		"riskLevel":        riskLevel,
		"riskFactors":      riskFactors,
		"baseInterestRate": baseInterestRate,
	}, nil
}

// GeneratePreQualifyTerms generates pre-qualification terms and loan offers
func (h *PreQualificationTaskHandler) GeneratePreQualifyTerms(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "generate_prequalify_terms"))

	logger.Info("Generating pre-qualification terms")

	// Extract input parameters
	annualIncome, _ := input["annualIncome"].(float64)
	employmentStatus, _ := input["employmentStatus"].(string)
	dtiRatio, _ := input["dtiRatio"].(float64)
	riskAssessment, _ := input["riskAssessment"].(map[string]interface{})

	// Extract risk assessment details
	riskLevel, _ := riskAssessment["riskLevel"].(string)
	baseInterestRate, _ := riskAssessment["baseInterestRate"].(float64)

	// Determine qualification
	qualified := h.determineQualification(dtiRatio, annualIncome, employmentStatus, riskLevel)

	var maxLoanAmount float64
	var interestRateRange map[string]float64
	var recommendedTerms []int
	var message string

	if qualified {
		// Calculate max loan amount based on income and DTI
		maxLoanAmount = h.calculateMaxLoanAmount(annualIncome, dtiRatio)

		// Calculate interest rate range
		interestRateRange = h.calculateInterestRateRange(baseInterestRate, dtiRatio, annualIncome)

		// Determine recommended terms
		recommendedTerms = h.determineRecommendedTerms(annualIncome, dtiRatio)

		message = "You are pre-qualified for a loan"
	} else {
		maxLoanAmount = 0
		interestRateRange = map[string]float64{"min": 0, "max": 0}
		recommendedTerms = []int{}

		// Generate specific message based on rejection reason
		message = h.generateRejectionMessage(dtiRatio, annualIncome, employmentStatus)
	}

	logger.Info("Pre-qualification terms generated",
		zap.Bool("qualified", qualified),
		zap.Float64("max_loan_amount", maxLoanAmount),
		zap.String("message", message),
	)

	return map[string]interface{}{
		"qualified":         qualified,
		"maxLoanAmount":     maxLoanAmount,
		"interestRateRange": interestRateRange,
		"recommendedTerms":  recommendedTerms,
		"message":           message,
	}, nil
}

// FinalizePreQualification finalizes the pre-qualification result
func (h *PreQualificationTaskHandler) FinalizePreQualification(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "finalize_prequalification"))

	logger.Info("Finalizing pre-qualification result")

	// Extract input parameters
	userID, _ := input["userId"].(string)
	qualified, _ := input["qualified"].(bool)
	maxLoanAmount, _ := input["maxLoanAmount"].(float64)
	interestRateRange, _ := input["interestRateRange"].(map[string]interface{})
	recommendedTerms, _ := input["recommendedTerms"].([]interface{})
	dtiRatio, _ := input["dtiRatio"].(float64)
	message, _ := input["message"].(string)

	// Generate pre-qualification ID
	prequalificationID := h.generatePreQualificationID(userID)

	// Extract interest rate range values
	minInterestRate, _ := interestRateRange["min"].(float64)
	maxInterestRate, _ := interestRateRange["max"].(float64)

	// Convert recommended terms to int slice
	terms := make([]int, len(recommendedTerms))
	for i, term := range recommendedTerms {
		if termInt, ok := term.(int); ok {
			terms[i] = termInt
		}
	}

	logger.Info("Pre-qualification finalized",
		zap.String("prequalification_id", prequalificationID),
		zap.Bool("qualified", qualified),
		zap.Float64("max_loan_amount", maxLoanAmount),
	)

	return map[string]interface{}{
		"qualified":          qualified,
		"maxLoanAmount":      maxLoanAmount,
		"minInterestRate":    minInterestRate,
		"maxInterestRate":    maxInterestRate,
		"recommendedTerms":   terms,
		"dtiRatio":           dtiRatio,
		"message":            message,
		"prequalificationId": prequalificationID,
	}, nil
}

// Helper methods

func (h *PreQualificationTaskHandler) determineQualification(
	dtiRatio, annualIncome float64,
	employmentStatus, riskLevel string,
) bool {
	// Basic qualification rules
	if dtiRatio > 0.43 {
		return false
	}

	if annualIncome < 25000 {
		return false
	}

	if employmentStatus == "unemployed" {
		return false
	}

	if riskLevel == "HIGH" && dtiRatio > 0.40 {
		return false
	}

	return true
}

func (h *PreQualificationTaskHandler) calculateMaxLoanAmount(annualIncome, dtiRatio float64) float64 {
	// Calculate max monthly payment (25% of monthly income)
	monthlyIncome := annualIncome / 12
	maxMonthlyPayment := monthlyIncome * 0.25

	// Adjust for DTI ratio
	if dtiRatio > 0.30 {
		maxMonthlyPayment *= 0.8 // Reduce by 20% for higher DTI
	}

	// Calculate max loan amount assuming 60-month term and 10% interest
	monthlyRate := 0.10 / 12
	termMonths := 60.0

	maxAmount := maxMonthlyPayment * (math.Pow(1+monthlyRate, termMonths) - 1) / (monthlyRate * math.Pow(1+monthlyRate, termMonths))

	// Cap at $50,000
	if maxAmount > 50000 {
		maxAmount = 50000
	}

	return math.Round(maxAmount*100) / 100
}

func (h *PreQualificationTaskHandler) calculateInterestRateRange(
	baseRate, dtiRatio, annualIncome float64,
) map[string]float64 {
	minRate := baseRate
	maxRate := baseRate + 2.0

	// Adjust based on DTI ratio
	if dtiRatio > 0.40 {
		minRate += 1.0
		maxRate += 1.5
	} else if dtiRatio > 0.30 {
		minRate += 0.5
		maxRate += 1.0
	}

	// Adjust based on income
	if annualIncome < 40000 {
		minRate += 0.5
		maxRate += 0.5
	}

	// Ensure rates are within reasonable bounds
	if minRate < 5.0 {
		minRate = 5.0
	}
	if maxRate > 25.0 {
		maxRate = 25.0
	}

	return map[string]float64{
		"min": math.Round(minRate*100) / 100,
		"max": math.Round(maxRate*100) / 100,
	}
}

func (h *PreQualificationTaskHandler) determineRecommendedTerms(annualIncome, dtiRatio float64) []int {
	var terms []int

	// Base terms
	if annualIncome >= 50000 && dtiRatio <= 0.30 {
		terms = []int{36, 48, 60, 72}
	} else if annualIncome >= 35000 && dtiRatio <= 0.35 {
		terms = []int{36, 48, 60}
	} else {
		terms = []int{36, 48}
	}

	return terms
}

func (h *PreQualificationTaskHandler) generateRejectionMessage(
	dtiRatio, annualIncome float64,
	employmentStatus string,
) string {
	if dtiRatio > 0.43 {
		return "Your debt-to-income ratio is too high for loan approval"
	}

	if annualIncome < 25000 {
		return "Your annual income is below the minimum requirement"
	}

	if employmentStatus == "unemployed" {
		return "Employment verification required for loan approval"
	}

	return "You do not currently qualify for a loan based on the provided information"
}

func (h *PreQualificationTaskHandler) generatePreQualificationID(userID string) string {
	timestamp := time.Now().UnixNano()
	return "PQ" + userID[:8] + "_" + string(rune(timestamp%1000000))
}

// UpdateApplicationState updates the state of a loan application
func (h *PreQualificationTaskHandler) UpdateApplicationState(
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
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}
	if fromState == "" {
		return nil, fmt.Errorf("from state is required")
	}
	if toState == "" {
		return nil, fmt.Errorf("to state is required")
	}
	if reason == "" {
		return nil, fmt.Errorf("reason is required")
	}

	logger.Info("Application state transition",
		zap.String("application_id", applicationID),
		zap.String("from_state", fromState),
		zap.String("to_state", toState),
		zap.String("reason", reason),
	)

	// TODO: In a real implementation, this would:
	// 1. Update the application state in the database
	// 2. Log the state transition
	// 3. Trigger any necessary notifications
	// 4. Update audit trail

	// For now, simulate successful state update
	output := map[string]interface{}{
		"success":   true,
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"newState":  toState,
		"message":   fmt.Sprintf("Application state updated from %s to %s: %s", fromState, toState, reason),
	}

	logger.Info("Application state updated successfully",
		zap.String("application_id", applicationID),
		zap.String("new_state", toState),
	)

	return output, nil
}
