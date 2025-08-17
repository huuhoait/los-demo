package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"underwriting-worker/domain"
)

// CreditService provides credit-related business logic
type CreditService struct {
	logger              *zap.Logger
	creditReportRepo    domain.CreditReportRepository
	creditBureauService domain.CreditBureauService
	auditLogger         domain.AuditLogger
}

// NewCreditService creates a new credit service
func NewCreditService(
	logger *zap.Logger,
	creditReportRepo domain.CreditReportRepository,
	creditBureauService domain.CreditBureauService,
	auditLogger domain.AuditLogger,
) *CreditService {
	return &CreditService{
		logger:              logger,
		creditReportRepo:    creditReportRepo,
		creditBureauService: creditBureauService,
		auditLogger:         auditLogger,
	}
}

// GetCreditReport gets or retrieves a credit report for an application
func (cs *CreditService) GetCreditReport(ctx context.Context, applicationID, userID string) (*domain.CreditReport, error) {
	logger := cs.logger.With(
		zap.String("application_id", applicationID),
		zap.String("user_id", userID),
		zap.String("operation", "get_credit_report"),
	)

	logger.Info("Getting credit report")

	// Check for existing recent credit report
	existingReport, err := cs.creditReportRepo.GetByApplicationID(ctx, applicationID)
	if err == nil {
		// Use existing report if it's less than 30 days old
		if time.Since(existingReport.ReportDate) < 30*24*time.Hour {
			logger.Info("Using existing credit report",
				zap.String("report_id", existingReport.ID),
				zap.Time("report_date", existingReport.ReportDate),
				zap.Int("credit_score", existingReport.CreditScore))
			return existingReport, nil
		}
	}

	// Request new credit report
	request := &domain.CreditReportRequest{
		UserID:        userID,
		ApplicationID: applicationID,
		ReportType:    "full",
		Permissible:   "loan_application",
	}

	creditReport, err := cs.creditBureauService.GetCreditReport(ctx, request)
	if err != nil {
		logger.Error("Failed to get credit report from bureau", zap.Error(err))
		return nil, fmt.Errorf("failed to get credit report: %w", err)
	}

	// Enrich credit report with additional analysis
	cs.enrichCreditReport(creditReport)

	// Save the credit report
	if err := cs.creditReportRepo.Create(ctx, creditReport); err != nil {
		logger.Error("Failed to save credit report", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	// Log audit event
	cs.logCreditReportEvent(ctx, applicationID, userID, creditReport)

	logger.Info("Credit report retrieved successfully",
		zap.String("report_id", creditReport.ID),
		zap.Int("credit_score", creditReport.CreditScore),
		zap.String("score_range", string(creditReport.CreditScoreRange)))

	return creditReport, nil
}

// AnalyzeCreditRisk analyzes credit risk factors from a credit report
func (cs *CreditService) AnalyzeCreditRisk(ctx context.Context, creditReport *domain.CreditReport) (*CreditRiskAnalysis, error) {
	logger := cs.logger.With(
		zap.String("report_id", creditReport.ID),
		zap.String("operation", "analyze_credit_risk"),
	)

	logger.Info("Analyzing credit risk")

	analysis := &CreditRiskAnalysis{
		CreditScore:         creditReport.CreditScore,
		CreditScoreRange:    creditReport.CreditScoreRange,
		CreditUtilization:   creditReport.CreditUtilization,
		TotalCreditLimit:    creditReport.TotalCreditLimit,
		TotalCurrentBalance: creditReport.TotalCurrentBalance,
		RiskFactors:         []CreditRiskFactor{},
		PositiveFactors:     []CreditRiskFactor{},
		RiskLevel:           domain.RiskLow,
		RiskScore:           0,
	}

	// Analyze credit score risk
	cs.analyzeCreditScoreRisk(analysis)

	// Analyze credit utilization risk
	cs.analyzeCreditUtilizationRisk(analysis)

	// Analyze payment history risk
	cs.analyzePaymentHistoryRisk(analysis, creditReport)

	// Analyze credit mix and account age
	cs.analyzeCreditMixRisk(analysis, creditReport)

	// Analyze derogatory items
	cs.analyzeDerogatoriesRisk(analysis, creditReport)

	// Analyze recent inquiries
	cs.analyzeInquiriesRisk(analysis, creditReport)

	// Calculate overall risk score
	cs.calculateCreditRiskScore(analysis)

	logger.Info("Credit risk analysis completed",
		zap.String("risk_level", string(analysis.RiskLevel)),
		zap.Float64("risk_score", analysis.RiskScore),
		zap.Int("risk_factors", len(analysis.RiskFactors)))

	return analysis, nil
}

// RefreshCreditReport refreshes a credit report if it's outdated
func (cs *CreditService) RefreshCreditReport(ctx context.Context, applicationID string) (*domain.CreditReport, error) {
	logger := cs.logger.With(
		zap.String("application_id", applicationID),
		zap.String("operation", "refresh_credit_report"),
	)

	logger.Info("Refreshing credit report")

	return cs.creditBureauService.RefreshCreditReport(ctx, applicationID)
}

// enrichCreditReport adds additional calculated fields to the credit report
func (cs *CreditService) enrichCreditReport(report *domain.CreditReport) {
	// Set credit score range
	report.CreditScoreRange = domain.GetCreditScoreRange(report.CreditScore)

	// Calculate credit utilization if not provided
	if report.CreditUtilization == 0 && report.TotalCreditLimit > 0 {
		report.CreditUtilization = report.TotalCurrentBalance / report.TotalCreditLimit
	}

	// Identify risk factors
	riskFactors := []string{}

	if report.CreditScore < 600 {
		riskFactors = append(riskFactors, "low_credit_score")
	}

	if report.CreditUtilization > 0.7 {
		riskFactors = append(riskFactors, "high_credit_utilization")
	}

	if report.PaymentHistory.LatePayments30 > 3 {
		riskFactors = append(riskFactors, "recent_late_payments")
	}

	if report.GetDerogatoriesCount() > 0 {
		riskFactors = append(riskFactors, "derogatory_items")
	}

	report.RiskFactors = riskFactors

	// Analyze credit mix
	creditTypes := make(map[string]bool)
	for _, account := range report.CreditAccounts {
		creditTypes[account.AccountType] = true
	}

	creditMix := []string{}
	for accountType := range creditTypes {
		creditMix = append(creditMix, accountType)
	}
	report.CreditMix = creditMix
}

// analyzeCreditScoreRisk analyzes risk based on credit score
func (cs *CreditService) analyzeCreditScoreRisk(analysis *CreditRiskAnalysis) {
	score := analysis.CreditScore

	switch {
	case score >= 800:
		analysis.PositiveFactors = append(analysis.PositiveFactors, CreditRiskFactor{
			Factor:      "excellent_credit_score",
			Description: "Excellent credit score (800+)",
			Impact:      "positive",
			Score:       10,
		})
	case score >= 740:
		analysis.PositiveFactors = append(analysis.PositiveFactors, CreditRiskFactor{
			Factor:      "very_good_credit_score",
			Description: "Very good credit score (740-799)",
			Impact:      "positive",
			Score:       5,
		})
	case score >= 670:
		// Neutral range, no factor added
	case score >= 580:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "fair_credit_score",
			Description: "Fair credit score (580-669)",
			Impact:      "medium",
			Score:       20,
		})
	default:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "poor_credit_score",
			Description: "Poor credit score (<580)",
			Impact:      "high",
			Score:       40,
		})
	}
}

// analyzeCreditUtilizationRisk analyzes risk based on credit utilization
func (cs *CreditService) analyzeCreditUtilizationRisk(analysis *CreditRiskAnalysis) {
	utilization := analysis.CreditUtilization

	switch {
	case utilization <= 0.1:
		analysis.PositiveFactors = append(analysis.PositiveFactors, CreditRiskFactor{
			Factor:      "low_credit_utilization",
			Description: "Very low credit utilization (<10%)",
			Impact:      "positive",
			Score:       5,
		})
	case utilization <= 0.3:
		// Good range, no factor added
	case utilization <= 0.7:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "moderate_credit_utilization",
			Description: "Moderate credit utilization (30-70%)",
			Impact:      "medium",
			Score:       15,
		})
	default:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "high_credit_utilization",
			Description: "High credit utilization (>70%)",
			Impact:      "high",
			Score:       25,
		})
	}
}

// analyzePaymentHistoryRisk analyzes risk based on payment history
func (cs *CreditService) analyzePaymentHistoryRisk(analysis *CreditRiskAnalysis, report *domain.CreditReport) {
	paymentHistory := report.PaymentHistory

	// Late payments risk
	totalLatePayments := paymentHistory.LatePayments30 + paymentHistory.LatePayments60 +
		paymentHistory.LatePayments90 + paymentHistory.LatePayments120

	switch {
	case totalLatePayments == 0:
		analysis.PositiveFactors = append(analysis.PositiveFactors, CreditRiskFactor{
			Factor:      "perfect_payment_history",
			Description: "No late payments on record",
			Impact:      "positive",
			Score:       10,
		})
	case totalLatePayments <= 2:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "minor_late_payments",
			Description: "Few late payments on record",
			Impact:      "low",
			Score:       5,
		})
	case totalLatePayments <= 5:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "moderate_late_payments",
			Description: "Moderate late payments on record",
			Impact:      "medium",
			Score:       15,
		})
	default:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "frequent_late_payments",
			Description: "Frequent late payments on record",
			Impact:      "high",
			Score:       30,
		})
	}

	// Severe delinquencies
	if paymentHistory.ChargeOffs > 0 {
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "charge_offs",
			Description: fmt.Sprintf("%d charge-off(s) on record", paymentHistory.ChargeOffs),
			Impact:      "high",
			Score:       35,
		})
	}

	if paymentHistory.Collections > 0 {
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "collections",
			Description: fmt.Sprintf("%d collection(s) on record", paymentHistory.Collections),
			Impact:      "high",
			Score:       30,
		})
	}
}

// analyzeCreditMixRisk analyzes risk based on credit mix and account age
func (cs *CreditService) analyzeCreditMixRisk(analysis *CreditRiskAnalysis, report *domain.CreditReport) {
	creditMix := report.CreditMix
	accounts := report.CreditAccounts

	// Credit mix diversity
	if len(creditMix) >= 3 {
		analysis.PositiveFactors = append(analysis.PositiveFactors, CreditRiskFactor{
			Factor:      "diverse_credit_mix",
			Description: "Diverse mix of credit types",
			Impact:      "positive",
			Score:       5,
		})
	}

	// Account age analysis
	if len(accounts) > 0 {
		oldestAccount := time.Now()
		avgAge := time.Duration(0)

		for _, account := range accounts {
			if account.OpenDate.Before(oldestAccount) {
				oldestAccount = account.OpenDate
			}
			avgAge += time.Since(account.OpenDate)
		}
		avgAge = avgAge / time.Duration(len(accounts))

		oldestAccountAge := time.Since(oldestAccount)

		if oldestAccountAge >= 7*365*24*time.Hour { // 7 years
			analysis.PositiveFactors = append(analysis.PositiveFactors, CreditRiskFactor{
				Factor:      "long_credit_history",
				Description: "Long credit history (7+ years)",
				Impact:      "positive",
				Score:       8,
			})
		} else if oldestAccountAge < 2*365*24*time.Hour { // 2 years
			analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
				Factor:      "short_credit_history",
				Description: "Short credit history (<2 years)",
				Impact:      "medium",
				Score:       15,
			})
		}
	}
}

// analyzeDerogatoriesRisk analyzes risk based on derogatory items
func (cs *CreditService) analyzeDerogatoriesRisk(analysis *CreditRiskAnalysis, report *domain.CreditReport) {
	derogatories := report.DerogatoryCounts

	if derogatories.Bankruptcies > 0 {
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "bankruptcy",
			Description: fmt.Sprintf("%d bankruptcy(ies) on record", derogatories.Bankruptcies),
			Impact:      "high",
			Score:       50,
		})
	}

	if derogatories.Liens > 0 {
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "tax_liens",
			Description: fmt.Sprintf("%d tax lien(s) on record", derogatories.Liens),
			Impact:      "high",
			Score:       35,
		})
	}

	if derogatories.Judgments > 0 {
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "judgments",
			Description: fmt.Sprintf("%d judgment(s) on record", derogatories.Judgments),
			Impact:      "high",
			Score:       30,
		})
	}
}

// analyzeInquiriesRisk analyzes risk based on recent credit inquiries
func (cs *CreditService) analyzeInquiriesRisk(analysis *CreditRiskAnalysis, report *domain.CreditReport) {
	recentInquiries := 0
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	for _, inquiry := range report.CreditInquiries {
		if inquiry.InquiryType == "hard" && inquiry.InquiryDate.After(sixMonthsAgo) {
			recentInquiries++
		}
	}

	switch {
	case recentInquiries == 0:
		// No factor for no inquiries
	case recentInquiries <= 2:
		// Normal range, no factor
	case recentInquiries <= 5:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "moderate_credit_seeking",
			Description: "Moderate recent credit seeking activity",
			Impact:      "medium",
			Score:       10,
		})
	default:
		analysis.RiskFactors = append(analysis.RiskFactors, CreditRiskFactor{
			Factor:      "high_credit_seeking",
			Description: "High recent credit seeking activity",
			Impact:      "high",
			Score:       20,
		})
	}
}

// calculateCreditRiskScore calculates the overall credit risk score
func (cs *CreditService) calculateCreditRiskScore(analysis *CreditRiskAnalysis) {
	riskScore := 0.0
	positiveScore := 0.0

	// Sum up risk factors
	for _, factor := range analysis.RiskFactors {
		riskScore += factor.Score
	}

	// Sum up positive factors
	for _, factor := range analysis.PositiveFactors {
		positiveScore += factor.Score
	}

	// Calculate net risk score (0-100 scale)
	netScore := riskScore - positiveScore
	if netScore < 0 {
		netScore = 0
	}
	if netScore > 100 {
		netScore = 100
	}

	analysis.RiskScore = netScore

	// Determine risk level
	switch {
	case netScore >= 60:
		analysis.RiskLevel = domain.RiskCritical
	case netScore >= 40:
		analysis.RiskLevel = domain.RiskHigh
	case netScore >= 20:
		analysis.RiskLevel = domain.RiskMedium
	default:
		analysis.RiskLevel = domain.RiskLow
	}
}

// logCreditReportEvent logs credit report access for audit
func (cs *CreditService) logCreditReportEvent(ctx context.Context, applicationID, userID string, report *domain.CreditReport) {
	event := &domain.AccessEvent{
		EventID:       report.ID + "_access",
		UserID:        userID,
		ApplicationID: applicationID,
		Action:        "credit_report_accessed",
		Resource:      "credit_report",
		Timestamp:     time.Now(),
		Success:       true,
	}

	if err := cs.auditLogger.LogAccessEvent(ctx, event); err != nil {
		cs.logger.Error("Failed to log credit report access event", zap.Error(err))
	}
}

// CreditRiskAnalysis represents the result of credit risk analysis
type CreditRiskAnalysis struct {
	CreditScore         int                     `json:"credit_score"`
	CreditScoreRange    domain.CreditScoreRange `json:"credit_score_range"`
	CreditUtilization   float64                 `json:"credit_utilization"`
	TotalCreditLimit    float64                 `json:"total_credit_limit"`
	TotalCurrentBalance float64                 `json:"total_current_balance"`
	RiskFactors         []CreditRiskFactor      `json:"risk_factors"`
	PositiveFactors     []CreditRiskFactor      `json:"positive_factors"`
	RiskLevel           domain.RiskLevel        `json:"risk_level"`
	RiskScore           float64                 `json:"risk_score"`
}

// CreditRiskFactor represents a specific credit risk factor
type CreditRiskFactor struct {
	Factor      string  `json:"factor"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // positive, low, medium, high
	Score       float64 `json:"score"`
}
