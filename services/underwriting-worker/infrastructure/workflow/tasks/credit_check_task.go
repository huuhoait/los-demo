package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"underwriting_worker/application/services"
	"underwriting_worker/application/usecases"
	"underwriting_worker/domain"
)

// CreditCheckTaskHandler handles credit check tasks
type CreditCheckTaskHandler struct {
	logger              *zap.Logger
	creditService       *services.CreditService
	underwritingUseCase *usecases.UnderwritingUseCase
	loanApplicationRepo domain.LoanApplicationRepository
	creditReportRepo    domain.CreditReportRepository
}

// NewCreditCheckTaskHandler creates a new credit check task handler
func NewCreditCheckTaskHandler(
	logger *zap.Logger,
	creditService *services.CreditService,
	underwritingUseCase *usecases.UnderwritingUseCase,
	loanApplicationRepo domain.LoanApplicationRepository,
	creditReportRepo domain.CreditReportRepository,
) *CreditCheckTaskHandler {
	return &CreditCheckTaskHandler{
		logger:              logger,
		creditService:       creditService,
		underwritingUseCase: underwritingUseCase,
		loanApplicationRepo: loanApplicationRepo,
		creditReportRepo:    creditReportRepo,
	}
}

// Execute performs credit check for a loan application
func (h *CreditCheckTaskHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	logger := h.logger.With(zap.String("operation", "credit_check"))

	logger.Info("Starting credit check task")

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

	logger.Info("Retrieved loan application",
		zap.String("application_id", applicationID),
		zap.String("user_id", userID),
		zap.Float64("loan_amount", application.LoanAmount))

	// Perform credit check
	creditReport, err := h.creditService.GetCreditReport(ctx, applicationID, userID)
	if err != nil {
		logger.Error("Credit check failed", zap.Error(err))
		return h.createFailureResponse(applicationID, err), nil
	}

	// Analyze credit risk
	riskAnalysis, err := h.creditService.AnalyzeCreditRisk(ctx, creditReport)
	if err != nil {
		logger.Error("Credit risk analysis failed", zap.Error(err))
		return h.createFailureResponse(applicationID, err), nil
	}

	// Determine if credit check passes basic requirements
	creditDecision := h.evaluateCreditDecision(creditReport, riskAnalysis, application)

	processingTime := time.Since(startTime)

	logger.Info("Credit check completed",
		zap.String("application_id", applicationID),
		zap.Int("credit_score", creditReport.CreditScore),
		zap.String("credit_range", string(creditReport.CreditScoreRange)),
		zap.String("risk_level", string(riskAnalysis.RiskLevel)),
		zap.Bool("credit_approved", creditDecision.Approved),
		zap.Duration("processing_time", processingTime))

	// Create response
	return map[string]interface{}{
		"success":           true,
		"applicationId":     applicationID,
		"userId":            userID,
		"creditScore":       creditReport.CreditScore,
		"creditScoreRange":  string(creditReport.CreditScoreRange),
		"creditUtilization": creditReport.CreditUtilization,
		"totalCreditLimit":  creditReport.TotalCreditLimit,
		"paymentHistory": map[string]interface{}{
			"onTimePayments": creditReport.PaymentHistory.OnTimePayments,
			"latePayments30": creditReport.PaymentHistory.LatePayments30,
			"latePayments60": creditReport.PaymentHistory.LatePayments60,
			"latePayments90": creditReport.PaymentHistory.LatePayments90,
			"paymentScore":   creditReport.PaymentHistory.PaymentScore,
		},
		"derogatoryCounts": map[string]interface{}{
			"bankruptcies": creditReport.DerogatoryCounts.Bankruptcies,
			"liens":        creditReport.DerogatoryCounts.Liens,
			"judgments":    creditReport.DerogatoryCounts.Judgments,
			"chargeOffs":   creditReport.DerogatoryCounts.ChargeOffs,
			"collections":  creditReport.DerogatoryCounts.Collections,
		},
		"riskAnalysis": map[string]interface{}{
			"riskLevel":       string(riskAnalysis.RiskLevel),
			"riskScore":       riskAnalysis.RiskScore,
			"riskFactors":     h.formatRiskFactors(riskAnalysis.RiskFactors),
			"positiveFactors": h.formatRiskFactors(riskAnalysis.PositiveFactors),
		},
		"creditDecision": map[string]interface{}{
			"approved":        creditDecision.Approved,
			"reason":          creditDecision.Reason,
			"recommendations": creditDecision.Recommendations,
			"manualReview":    creditDecision.ManualReview,
		},
		"reportDetails": map[string]interface{}{
			"reportId":       creditReport.ID,
			"reportProvider": creditReport.ReportProvider,
			"reportDate":     creditReport.ReportDate.Format(time.RFC3339),
			"riskFactors":    creditReport.RiskFactors,
			"creditMix":      creditReport.CreditMix,
		},
		"processingTime": processingTime.String(),
		"completedAt":    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// evaluateCreditDecision evaluates whether the credit check passes basic requirements
func (h *CreditCheckTaskHandler) evaluateCreditDecision(
	creditReport *domain.CreditReport,
	riskAnalysis *services.CreditRiskAnalysis,
	application *domain.LoanApplication,
) CreditDecision {
	decision := CreditDecision{
		Approved:        false,
		Reason:          "",
		Recommendations: []string{},
		ManualReview:    false,
	}

	// Basic credit score requirements
	minCreditScore := 580 // Minimum credit score for consideration
	if creditReport.CreditScore < minCreditScore {
		decision.Reason = fmt.Sprintf("Credit score %d is below minimum requirement of %d",
			creditReport.CreditScore, minCreditScore)
		decision.Recommendations = append(decision.Recommendations,
			"Improve credit score by paying down existing debts and ensuring on-time payments")
		return decision
	}

	// Check for recent bankruptcies
	if creditReport.DerogatoryCounts.Bankruptcies > 0 {
		decision.Reason = "Recent bankruptcy on credit report"
		decision.ManualReview = true
		decision.Recommendations = append(decision.Recommendations,
			"Manual review required due to bankruptcy history")
		return decision
	}

	// Check credit utilization
	maxUtilization := 0.8 // 80% max utilization
	if creditReport.CreditUtilization > maxUtilization {
		decision.Reason = fmt.Sprintf("Credit utilization %.1f%% exceeds maximum of %.1f%%",
			creditReport.CreditUtilization*100, maxUtilization*100)
		decision.Recommendations = append(decision.Recommendations,
			"Pay down existing credit card balances to reduce utilization")
		return decision
	}

	// Check for excessive recent late payments
	totalRecentLates := creditReport.PaymentHistory.LatePayments30 +
		creditReport.PaymentHistory.LatePayments60 +
		creditReport.PaymentHistory.LatePayments90

	if totalRecentLates > 5 {
		decision.Reason = fmt.Sprintf("Too many recent late payments: %d", totalRecentLates)
		decision.Recommendations = append(decision.Recommendations,
			"Establish consistent on-time payment history")
		return decision
	}

	// Check overall risk level
	if riskAnalysis.RiskLevel == domain.RiskCritical {
		decision.Reason = "Credit risk level is too high"
		decision.ManualReview = true
		decision.Recommendations = append(decision.Recommendations,
			"Manual underwriter review required due to high risk factors")
		return decision
	}

	// If we get here, credit check passes
	decision.Approved = true
	decision.Reason = "Credit check passed basic requirements"

	// Add recommendations based on risk level
	switch riskAnalysis.RiskLevel {
	case domain.RiskHigh:
		decision.Recommendations = append(decision.Recommendations,
			"Consider additional income verification due to elevated credit risk")
	case domain.RiskMedium:
		decision.Recommendations = append(decision.Recommendations,
			"Standard underwriting process with additional documentation")
	default:
		decision.Recommendations = append(decision.Recommendations,
			"Credit profile supports standard underwriting")
	}

	return decision
}

// formatRiskFactors formats risk factors for output
func (h *CreditCheckTaskHandler) formatRiskFactors(factors []services.CreditRiskFactor) []map[string]interface{} {
	formatted := make([]map[string]interface{}, len(factors))
	for i, factor := range factors {
		formatted[i] = map[string]interface{}{
			"factor":      factor.Factor,
			"description": factor.Description,
			"impact":      factor.Impact,
			"score":       factor.Score,
		}
	}
	return formatted
}

// createFailureResponse creates a failure response
func (h *CreditCheckTaskHandler) createFailureResponse(applicationID string, err error) map[string]interface{} {
	return map[string]interface{}{
		"success":       false,
		"applicationId": applicationID,
		"error":         err.Error(),
		"creditDecision": map[string]interface{}{
			"approved":     false,
			"reason":       "Credit check failed due to system error",
			"manualReview": true,
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}
}

// CreditDecision represents the result of credit evaluation
type CreditDecision struct {
	Approved        bool     `json:"approved"`
	Reason          string   `json:"reason"`
	Recommendations []string `json:"recommendations"`
	ManualReview    bool     `json:"manual_review"`
}
