package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"underwriting_worker/application/usecases"
	"underwriting_worker/domain"
)

// RiskAssessmentTaskHandler handles risk assessment tasks
type RiskAssessmentTaskHandler struct {
	logger              *zap.Logger
	underwritingUseCase *usecases.UnderwritingUseCase
	loanApplicationRepo domain.LoanApplicationRepository
	creditReportRepo    domain.CreditReportRepository
	riskAssessmentRepo  domain.RiskAssessmentRepository
	riskScoringService  domain.RiskScoringService
}

// NewRiskAssessmentTaskHandler creates a new risk assessment task handler
func NewRiskAssessmentTaskHandler(
	logger *zap.Logger,
	underwritingUseCase *usecases.UnderwritingUseCase,
	loanApplicationRepo domain.LoanApplicationRepository,
	creditReportRepo domain.CreditReportRepository,
	riskAssessmentRepo domain.RiskAssessmentRepository,
	riskScoringService domain.RiskScoringService,
) *RiskAssessmentTaskHandler {
	return &RiskAssessmentTaskHandler{
		logger:              logger,
		underwritingUseCase: underwritingUseCase,
		loanApplicationRepo: loanApplicationRepo,
		creditReportRepo:    creditReportRepo,
		riskAssessmentRepo:  riskAssessmentRepo,
		riskScoringService:  riskScoringService,
	}
}

// Execute performs risk assessment for a loan application
func (h *RiskAssessmentTaskHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	logger := h.logger.With(zap.String("operation", "risk_assessment"))

	logger.Info("Starting risk assessment task")

	// Extract input parameters
	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	userID, ok := input["userId"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get loan application
	application, err := h.loanApplicationRepo.GetByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get loan application", zap.Error(err))
		return nil, fmt.Errorf("failed to get loan application: %w", err)
	}

	// Get credit report
	creditReport, err := h.creditReportRepo.GetByApplicationID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get credit report", zap.Error(err))
		return nil, fmt.Errorf("failed to get credit report: %w", err)
	}

	logger.Info("Starting risk assessment",
		zap.String("application_id", applicationID),
		zap.String("user_id", userID),
		zap.Float64("loan_amount", application.LoanAmount),
		zap.Int("credit_score", creditReport.CreditScore))

	// Check for existing recent risk assessment
	existingAssessment, err := h.riskAssessmentRepo.GetByApplicationID(ctx, applicationID)
	if err == nil && time.Since(existingAssessment.CreatedAt) < 24*time.Hour {
		logger.Info("Using existing risk assessment",
			zap.String("assessment_id", existingAssessment.ID),
			zap.String("risk_level", string(existingAssessment.OverallRiskLevel)))

		return h.createResponseFromExisting(existingAssessment, time.Since(startTime)), nil
	}

	// Perform risk assessment
	riskAssessment, err := h.performRiskAssessment(ctx, application, creditReport)
	if err != nil {
		logger.Error("Risk assessment failed", zap.Error(err))
		return h.createFailureResponse(applicationID, err), nil
	}

	// Perform additional risk analysis
	riskAnalysis := h.performDetailedRiskAnalysis(riskAssessment, application, creditReport)

	// Save risk assessment
	if err := h.riskAssessmentRepo.Create(ctx, riskAssessment); err != nil {
		logger.Error("Failed to save risk assessment", zap.Error(err))
		// Don't fail the process, just log the error
	}

	processingTime := time.Since(startTime)

	logger.Info("Risk assessment completed",
		zap.String("application_id", applicationID),
		zap.String("overall_risk_level", string(riskAssessment.OverallRiskLevel)),
		zap.Float64("risk_score", riskAssessment.RiskScore),
		zap.Float64("probability_of_default", riskAssessment.ProbabilityOfDefault),
		zap.Duration("processing_time", processingTime))

	// Create response
	return map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"userId":        userID,
		"riskAssessment": map[string]interface{}{
			"assessmentId":         riskAssessment.ID,
			"overallRiskLevel":     string(riskAssessment.OverallRiskLevel),
			"riskScore":            riskAssessment.RiskScore,
			"creditRiskScore":      riskAssessment.CreditRiskScore,
			"incomeRiskScore":      riskAssessment.IncomeRiskScore,
			"debtRiskScore":        riskAssessment.DebtRiskScore,
			"fraudRiskScore":       riskAssessment.FraudRiskScore,
			"probabilityOfDefault": riskAssessment.ProbabilityOfDefault,
			"recommendedAction":    riskAssessment.RecommendedAction,
			"confidenceLevel":      riskAssessment.ConfidenceLevel,
			"modelVersion":         riskAssessment.ModelVersion,
		},
		"riskFactors":       h.formatRiskFactors(riskAssessment.RiskFactors),
		"mitigatingFactors": h.formatMitigatingFactors(riskAssessment.MitigatingFactors),
		"riskAnalysis": map[string]interface{}{
			"riskCategory":           riskAnalysis.RiskCategory,
			"primaryRiskDrivers":     riskAnalysis.PrimaryRiskDrivers,
			"riskMitigators":         riskAnalysis.RiskMitigators,
			"riskRecommendations":    riskAnalysis.RiskRecommendations,
			"additionalRequirements": riskAnalysis.AdditionalRequirements,
			"monitoringRequired":     riskAnalysis.MonitoringRequired,
		},
		"scoreBreakdown": map[string]interface{}{
			"creditScore":     riskAssessment.CreditRiskScore,
			"incomeScore":     riskAssessment.IncomeRiskScore,
			"debtScore":       riskAssessment.DebtRiskScore,
			"fraudScore":      riskAssessment.FraudRiskScore,
			"totalScore":      riskAssessment.RiskScore,
			"scoreComponents": riskAnalysis.ScoreComponents,
		},
		"processingTime": processingTime.String(),
		"completedAt":    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// performRiskAssessment performs the actual risk assessment
func (h *RiskAssessmentTaskHandler) performRiskAssessment(
	ctx context.Context,
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
) (*domain.RiskAssessment, error) {
	// Use risk scoring service
	riskAssessment, err := h.riskScoringService.CalculateRiskScore(ctx, application, creditReport)
	if err != nil {
		// If service fails, create a basic risk assessment
		riskAssessment = h.createBasicRiskAssessment(application, creditReport)
	}

	// Enhance assessment with additional calculations
	h.enhanceRiskAssessment(riskAssessment, application, creditReport)

	return riskAssessment, nil
}

// createBasicRiskAssessment creates a basic risk assessment when service is unavailable
func (h *RiskAssessmentTaskHandler) createBasicRiskAssessment(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
) *domain.RiskAssessment {
	// Basic risk scoring algorithm
	creditRisk := h.calculateCreditRiskScore(creditReport)
	incomeRisk := h.calculateIncomeRiskScore(application)
	debtRisk := h.calculateDebtRiskScore(application)
	fraudRisk := h.calculateFraudRiskScore(application)

	// Weighted overall risk score
	overallRisk := (creditRisk*0.4 + incomeRisk*0.3 + debtRisk*0.2 + fraudRisk*0.1)

	riskLevel := domain.GetRiskLevel(overallRisk)

	return &domain.RiskAssessment{
		ID:                   application.ID + "_risk_assessment",
		ApplicationID:        application.ID,
		UserID:               application.UserID,
		OverallRiskLevel:     riskLevel,
		RiskScore:            overallRisk,
		CreditRiskScore:      creditRisk,
		IncomeRiskScore:      incomeRisk,
		DebtRiskScore:        debtRisk,
		FraudRiskScore:       fraudRisk,
		RiskFactors:          []domain.RiskFactor{},
		MitigatingFactors:    []domain.MitigatingFactor{},
		ProbabilityOfDefault: h.calculateProbabilityOfDefault(overallRisk),
		RecommendedAction:    h.getRecommendedAction(riskLevel),
		ConfidenceLevel:      0.85, // 85% confidence for basic assessment
		ModelVersion:         "basic_v1.0",
		AssessmentData:       make(map[string]interface{}),
		CreatedAt:            time.Now(),
	}
}

// enhanceRiskAssessment adds additional risk factors and mitigating factors
func (h *RiskAssessmentTaskHandler) enhanceRiskAssessment(
	assessment *domain.RiskAssessment,
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
) {
	// Add detailed risk factors
	riskFactors := h.identifyRiskFactors(application, creditReport, assessment)
	assessment.RiskFactors = riskFactors

	// Add mitigating factors
	mitigatingFactors := h.identifyMitigatingFactors(application, creditReport, assessment)
	assessment.MitigatingFactors = mitigatingFactors

	// Update assessment data with detailed analysis
	assessment.AssessmentData = map[string]interface{}{
		"credit_utilization":   creditReport.CreditUtilization,
		"dti_ratio":            application.CalculateDTI(),
		"loan_to_income":       application.LoanAmount / application.AnnualIncome,
		"credit_age_months":    h.calculateCreditAge(creditReport),
		"derogatory_count":     creditReport.GetDerogatoriesCount(),
		"inquiry_count_6m":     h.countRecentInquiries(creditReport, 6),
		"employment_verified":  application.IncomeVerificationStatus == domain.IncomeVerified,
		"assessment_timestamp": time.Now(),
	}
}

// performDetailedRiskAnalysis performs additional detailed risk analysis
func (h *RiskAssessmentTaskHandler) performDetailedRiskAnalysis(
	assessment *domain.RiskAssessment,
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
) RiskAnalysis {
	analysis := RiskAnalysis{
		RiskCategory:           h.categorizeRisk(assessment.OverallRiskLevel, assessment.RiskScore),
		PrimaryRiskDrivers:     h.identifyPrimaryRiskDrivers(assessment),
		RiskMitigators:         h.identifyRiskMitigators(assessment),
		RiskRecommendations:    []string{},
		AdditionalRequirements: []string{},
		MonitoringRequired:     assessment.OverallRiskLevel >= domain.RiskMedium,
		ScoreComponents:        h.createScoreComponents(assessment),
	}

	// Generate recommendations based on risk level
	analysis.RiskRecommendations = h.generateRiskRecommendations(assessment, application)

	// Determine additional requirements
	analysis.AdditionalRequirements = h.determineAdditionalRequirements(assessment, application)

	return analysis
}

// Risk calculation methods
func (h *RiskAssessmentTaskHandler) calculateCreditRiskScore(creditReport *domain.CreditReport) float64 {
	score := 0.0

	// Credit score impact (0-40 points)
	switch {
	case creditReport.CreditScore >= 800:
		score += 5
	case creditReport.CreditScore >= 740:
		score += 10
	case creditReport.CreditScore >= 670:
		score += 20
	case creditReport.CreditScore >= 580:
		score += 35
	default:
		score += 50
	}

	// Credit utilization impact (0-20 points)
	if creditReport.CreditUtilization > 0.8 {
		score += 20
	} else if creditReport.CreditUtilization > 0.5 {
		score += 15
	} else if creditReport.CreditUtilization > 0.3 {
		score += 10
	}

	// Payment history impact (0-30 points)
	totalLates := creditReport.PaymentHistory.LatePayments30 +
		creditReport.PaymentHistory.LatePayments60 +
		creditReport.PaymentHistory.LatePayments90
	score += float64(totalLates) * 3

	// Derogatory items impact (0-30 points)
	derogCount := creditReport.GetDerogatoriesCount()
	score += float64(derogCount) * 10

	return min(score, 100)
}

func (h *RiskAssessmentTaskHandler) calculateIncomeRiskScore(application *domain.LoanApplication) float64 {
	score := 0.0

	// Income adequacy
	if application.AnnualIncome < 25000 {
		score += 40
	} else if application.AnnualIncome < 40000 {
		score += 20
	} else if application.AnnualIncome < 60000 {
		score += 10
	}

	// Employment status
	switch application.EmploymentStatus {
	case "unemployed":
		score += 50
	case "part_time":
		score += 30
	case "self_employed":
		score += 20
	case "retired":
		score += 15
	}

	// Income verification status
	if application.IncomeVerificationStatus == domain.IncomeFailed {
		score += 30
	} else if application.IncomeVerificationStatus == domain.IncomeUnverified {
		score += 20
	}

	return min(score, 100)
}

func (h *RiskAssessmentTaskHandler) calculateDebtRiskScore(application *domain.LoanApplication) float64 {
	dti := application.CalculateDTI()

	switch {
	case dti > 0.5:
		return 80
	case dti > 0.43:
		return 60
	case dti > 0.36:
		return 40
	case dti > 0.28:
		return 20
	default:
		return 10
	}
}

func (h *RiskAssessmentTaskHandler) calculateFraudRiskScore(application *domain.LoanApplication) float64 {
	// Basic fraud indicators (in real implementation, this would be more sophisticated)
	score := 0.0

	// Loan amount vs income ratio
	if application.AnnualIncome > 0 {
		loanToIncome := application.LoanAmount / application.AnnualIncome
		if loanToIncome > 2 {
			score += 20
		} else if loanToIncome > 1 {
			score += 10
		}
	}

	return min(score, 100)
}

func (h *RiskAssessmentTaskHandler) calculateProbabilityOfDefault(riskScore float64) float64 {
	// Convert risk score to probability of default
	// This is a simplified mapping - real models would be more complex
	return min(riskScore/100*0.3, 0.3) // Max 30% PD
}

func (h *RiskAssessmentTaskHandler) getRecommendedAction(riskLevel domain.RiskLevel) string {
	switch riskLevel {
	case domain.RiskLow:
		return "approve_standard_terms"
	case domain.RiskMedium:
		return "approve_with_conditions"
	case domain.RiskHigh:
		return "manual_review_required"
	case domain.RiskCritical:
		return "decline"
	default:
		return "manual_review_required"
	}
}

// Risk factor identification methods
func (h *RiskAssessmentTaskHandler) identifyRiskFactors(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	assessment *domain.RiskAssessment,
) []domain.RiskFactor {
	factors := []domain.RiskFactor{}

	// Credit-related risk factors
	if creditReport.CreditScore < 620 {
		factors = append(factors, domain.RiskFactor{
			FactorID:    "low_credit_score",
			FactorType:  "credit",
			Description: fmt.Sprintf("Credit score %d below preferred range", creditReport.CreditScore),
			Impact:      "high",
			Score:       assessment.CreditRiskScore,
			Weight:      0.4,
		})
	}

	if creditReport.CreditUtilization > 0.7 {
		factors = append(factors, domain.RiskFactor{
			FactorID:    "high_credit_utilization",
			FactorType:  "credit",
			Description: fmt.Sprintf("Credit utilization %.1f%% is high", creditReport.CreditUtilization*100),
			Impact:      "medium",
			Score:       20,
			Weight:      0.2,
		})
	}

	// Income-related risk factors
	if application.AnnualIncome < 35000 {
		factors = append(factors, domain.RiskFactor{
			FactorID:    "low_income",
			FactorType:  "income",
			Description: fmt.Sprintf("Annual income $%.0f below preferred range", application.AnnualIncome),
			Impact:      "medium",
			Score:       assessment.IncomeRiskScore,
			Weight:      0.3,
		})
	}

	// Debt-related risk factors
	dti := application.CalculateDTI()
	if dti > 0.4 {
		factors = append(factors, domain.RiskFactor{
			FactorID:    "high_dti_ratio",
			FactorType:  "debt",
			Description: fmt.Sprintf("Debt-to-income ratio %.1f%% is high", dti*100),
			Impact:      "high",
			Score:       assessment.DebtRiskScore,
			Weight:      0.2,
		})
	}

	return factors
}

func (h *RiskAssessmentTaskHandler) identifyMitigatingFactors(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	assessment *domain.RiskAssessment,
) []domain.MitigatingFactor {
	factors := []domain.MitigatingFactor{}

	// Positive credit factors
	if creditReport.CreditScore >= 750 {
		factors = append(factors, domain.MitigatingFactor{
			FactorID:    "excellent_credit_score",
			FactorType:  "credit",
			Description: fmt.Sprintf("Excellent credit score of %d", creditReport.CreditScore),
			Impact:      "high",
			Score:       10,
			Weight:      0.4,
		})
	}

	if creditReport.PaymentHistory.OnTimePayments > 20 {
		factors = append(factors, domain.MitigatingFactor{
			FactorID:    "strong_payment_history",
			FactorType:  "credit",
			Description: "Strong payment history with minimal late payments",
			Impact:      "medium",
			Score:       5,
			Weight:      0.3,
		})
	}

	// Income stability factors
	if application.IncomeVerificationStatus == domain.IncomeVerified {
		factors = append(factors, domain.MitigatingFactor{
			FactorID:    "verified_income",
			FactorType:  "income",
			Description: "Income successfully verified through documentation",
			Impact:      "medium",
			Score:       5,
			Weight:      0.3,
		})
	}

	return factors
}

// Analysis helper methods
func (h *RiskAssessmentTaskHandler) categorizeRisk(riskLevel domain.RiskLevel, riskScore float64) string {
	categories := map[domain.RiskLevel]string{
		domain.RiskLow:      "low_risk_prime",
		domain.RiskMedium:   "medium_risk_near_prime",
		domain.RiskHigh:     "high_risk_subprime",
		domain.RiskCritical: "critical_risk_deep_subprime",
	}

	if category, exists := categories[riskLevel]; exists {
		return category
	}
	return "unknown_risk_category"
}

func (h *RiskAssessmentTaskHandler) identifyPrimaryRiskDrivers(assessment *domain.RiskAssessment) []string {
	drivers := []string{}

	if assessment.CreditRiskScore > 40 {
		drivers = append(drivers, "credit_profile")
	}
	if assessment.IncomeRiskScore > 30 {
		drivers = append(drivers, "income_stability")
	}
	if assessment.DebtRiskScore > 40 {
		drivers = append(drivers, "debt_burden")
	}
	if assessment.FraudRiskScore > 20 {
		drivers = append(drivers, "fraud_indicators")
	}

	return drivers
}

func (h *RiskAssessmentTaskHandler) identifyRiskMitigators(assessment *domain.RiskAssessment) []string {
	mitigators := []string{}

	for _, factor := range assessment.MitigatingFactors {
		mitigators = append(mitigators, factor.FactorType)
	}

	return mitigators
}

func (h *RiskAssessmentTaskHandler) generateRiskRecommendations(
	assessment *domain.RiskAssessment,
	application *domain.LoanApplication,
) []string {
	recommendations := []string{}

	switch assessment.OverallRiskLevel {
	case domain.RiskLow:
		recommendations = append(recommendations, "Standard underwriting process")
		recommendations = append(recommendations, "Consider premium pricing tier")
	case domain.RiskMedium:
		recommendations = append(recommendations, "Additional income verification recommended")
		recommendations = append(recommendations, "Consider standard pricing with monitoring")
	case domain.RiskHigh:
		recommendations = append(recommendations, "Manual underwriter review required")
		recommendations = append(recommendations, "Consider conditional approval with additional requirements")
	case domain.RiskCritical:
		recommendations = append(recommendations, "Decline recommendation")
		recommendations = append(recommendations, "Consider alternative products if available")
	}

	return recommendations
}

func (h *RiskAssessmentTaskHandler) determineAdditionalRequirements(
	assessment *domain.RiskAssessment,
	application *domain.LoanApplication,
) []string {
	requirements := []string{}

	if assessment.IncomeRiskScore > 30 {
		requirements = append(requirements, "Additional income documentation")
	}

	if assessment.CreditRiskScore > 50 {
		requirements = append(requirements, "Explanation of credit issues")
	}

	if assessment.DebtRiskScore > 40 {
		requirements = append(requirements, "Debt consolidation plan")
	}

	return requirements
}

func (h *RiskAssessmentTaskHandler) createScoreComponents(assessment *domain.RiskAssessment) map[string]interface{} {
	return map[string]interface{}{
		"credit_score_component": map[string]interface{}{
			"score":  assessment.CreditRiskScore,
			"weight": 0.4,
			"impact": assessment.CreditRiskScore * 0.4,
		},
		"income_score_component": map[string]interface{}{
			"score":  assessment.IncomeRiskScore,
			"weight": 0.3,
			"impact": assessment.IncomeRiskScore * 0.3,
		},
		"debt_score_component": map[string]interface{}{
			"score":  assessment.DebtRiskScore,
			"weight": 0.2,
			"impact": assessment.DebtRiskScore * 0.2,
		},
		"fraud_score_component": map[string]interface{}{
			"score":  assessment.FraudRiskScore,
			"weight": 0.1,
			"impact": assessment.FraudRiskScore * 0.1,
		},
	}
}

// Helper calculation methods
func (h *RiskAssessmentTaskHandler) calculateCreditAge(creditReport *domain.CreditReport) int {
	if len(creditReport.CreditAccounts) == 0 {
		return 0
	}

	oldestDate := time.Now()
	for _, account := range creditReport.CreditAccounts {
		if account.OpenDate.Before(oldestDate) {
			oldestDate = account.OpenDate
		}
	}

	return int(time.Since(oldestDate).Hours() / 24 / 30) // months
}

func (h *RiskAssessmentTaskHandler) countRecentInquiries(creditReport *domain.CreditReport, months int) int {
	cutoffDate := time.Now().AddDate(0, -months, 0)
	count := 0

	for _, inquiry := range creditReport.CreditInquiries {
		if inquiry.InquiryType == "hard" && inquiry.InquiryDate.After(cutoffDate) {
			count++
		}
	}

	return count
}

// Response creation methods
func (h *RiskAssessmentTaskHandler) createResponseFromExisting(
	assessment *domain.RiskAssessment,
	processingTime time.Duration,
) map[string]interface{} {
	return map[string]interface{}{
		"success":       true,
		"applicationId": assessment.ApplicationID,
		"userId":        assessment.UserID,
		"riskAssessment": map[string]interface{}{
			"assessmentId":         assessment.ID,
			"overallRiskLevel":     string(assessment.OverallRiskLevel),
			"riskScore":            assessment.RiskScore,
			"probabilityOfDefault": assessment.ProbabilityOfDefault,
			"recommendedAction":    assessment.RecommendedAction,
			"confidenceLevel":      assessment.ConfidenceLevel,
			"modelVersion":         assessment.ModelVersion,
		},
		"cached":         true,
		"processingTime": processingTime.String(),
		"completedAt":    time.Now().UTC().Format(time.RFC3339),
	}
}

func (h *RiskAssessmentTaskHandler) createFailureResponse(applicationID string, err error) map[string]interface{} {
	return map[string]interface{}{
		"success":       false,
		"applicationId": applicationID,
		"error":         err.Error(),
		"riskAssessment": map[string]interface{}{
			"overallRiskLevel":  string(domain.RiskCritical),
			"recommendedAction": "manual_review_required",
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}
}

// Format methods for response
func (h *RiskAssessmentTaskHandler) formatRiskFactors(factors []domain.RiskFactor) []map[string]interface{} {
	formatted := make([]map[string]interface{}, len(factors))
	for i, factor := range factors {
		formatted[i] = map[string]interface{}{
			"factorId":    factor.FactorID,
			"factorType":  factor.FactorType,
			"description": factor.Description,
			"impact":      factor.Impact,
			"score":       factor.Score,
			"weight":      factor.Weight,
		}
	}
	return formatted
}

func (h *RiskAssessmentTaskHandler) formatMitigatingFactors(factors []domain.MitigatingFactor) []map[string]interface{} {
	formatted := make([]map[string]interface{}, len(factors))
	for i, factor := range factors {
		formatted[i] = map[string]interface{}{
			"factorId":    factor.FactorID,
			"factorType":  factor.FactorType,
			"description": factor.Description,
			"impact":      factor.Impact,
			"score":       factor.Score,
			"weight":      factor.Weight,
		}
	}
	return formatted
}

// Utility functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RiskAnalysis represents detailed risk analysis results
type RiskAnalysis struct {
	RiskCategory           string                 `json:"risk_category"`
	PrimaryRiskDrivers     []string               `json:"primary_risk_drivers"`
	RiskMitigators         []string               `json:"risk_mitigators"`
	RiskRecommendations    []string               `json:"risk_recommendations"`
	AdditionalRequirements []string               `json:"additional_requirements"`
	MonitoringRequired     bool                   `json:"monitoring_required"`
	ScoreComponents        map[string]interface{} `json:"score_components"`
}
