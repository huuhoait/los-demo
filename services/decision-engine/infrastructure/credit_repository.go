package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huuhoait/los-demo/services/decision-engine/domain"
	"go.uber.org/zap"
)

// CreditBureauRepository handles external credit bureau integrations
type CreditBureauRepository struct {
	logger *zap.Logger
	config CreditBureauConfig
}

// CreditBureauConfig holds configuration for credit bureau services
type CreditBureauConfig struct {
	ExperianEndpoint   string
	EquifaxEndpoint    string
	TransUnionEndpoint string
	APITimeout         time.Duration
	RetryAttempts      int
}

// NewCreditBureauRepository creates a new credit bureau repository
func NewCreditBureauRepository(logger *zap.Logger, config CreditBureauConfig) *CreditBureauRepository {
	return &CreditBureauRepository{
		logger: logger,
		config: config,
	}
}

// GetCreditScore retrieves credit score from primary bureau
func (r *CreditBureauRepository) GetCreditScore(ctx context.Context, request *domain.CreditScoreRequest) (*domain.CreditScoreResponse, error) {
	logger := r.logger.With(
		zap.String("ssn", maskSSN(request.SSN)),
		zap.String("operation", "get_credit_score"),
	)

	logger.Info("Retrieving credit score from bureau")

	// In production, this would make actual API calls to credit bureaus
	// For now, we'll simulate the response based on provided data
	response := r.simulateCreditBureauResponse(request)

	logger.Info("Credit score retrieved",
		zap.Int("credit_score", response.CreditScore),
		zap.String("report_date", response.ReportDate.Format("2006-01-02")),
	)

	return response, nil
}

// GetDetailedCreditReport retrieves full credit report
func (r *CreditBureauRepository) GetDetailedCreditReport(ctx context.Context, request *domain.CreditReportRequest) (*domain.CreditReport, error) {
	logger := r.logger.With(
		zap.String("ssn", maskSSN(request.SSN)),
		zap.String("operation", "get_credit_report"),
	)

	logger.Info("Retrieving detailed credit report")

	// Simulate detailed credit report
	report := r.simulateDetailedCreditReport(request)

	logger.Info("Credit report retrieved",
		zap.Int("account_count", len(report.Accounts)),
		zap.Int("inquiry_count", len(report.Inquiries)),
	)

	return report, nil
}

// simulateCreditBureauResponse simulates credit bureau API response
func (r *CreditBureauRepository) simulateCreditBureauResponse(request *domain.CreditScoreRequest) *domain.CreditScoreResponse {
	// This is simulation logic - in production, replace with actual API calls

	// Use last 4 digits of SSN to generate consistent but varied scores
	lastFour := request.SSN[len(request.SSN)-4:]
	seed := 0
	for _, char := range lastFour {
		seed += int(char)
	}

	// Generate score between 300-850 based on seed
	baseScore := 300 + (seed % 551)

	// Adjust based on name length (arbitrary simulation factor)
	nameAdjustment := len(request.FirstName) + len(request.LastName)
	adjustedScore := baseScore + (nameAdjustment % 100)

	// Ensure score is in valid range
	if adjustedScore > 850 {
		adjustedScore = 850
	}
	if adjustedScore < 300 {
		adjustedScore = 300
	}

	return &domain.CreditScoreResponse{
		CreditScore: adjustedScore,
		ScoreType:   "FICO",
		ScoreRange:  "300-850",
		Bureau:      "EXPERIAN",
		ReportDate:  time.Now(),
		FactorCodes: r.generateFactorCodes(adjustedScore),
		RiskLevel:   r.categorizeRiskLevel(adjustedScore),
	}
}

// simulateDetailedCreditReport simulates detailed credit report
func (r *CreditBureauRepository) simulateDetailedCreditReport(request *domain.CreditReportRequest) *domain.CreditReport {
	creditScore := r.simulateCreditBureauResponse(&domain.CreditScoreRequest{
		SSN:       request.SSN,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	})

	report := &domain.CreditReport{
		PersonalInfo: domain.PersonalInfo{
			FirstName:   request.FirstName,
			LastName:    request.LastName,
			SSN:         maskSSN(request.SSN),
			DateOfBirth: request.DateOfBirth,
			Address:     request.Address,
		},
		CreditScore: creditScore.CreditScore,
		ReportDate:  time.Now(),
		Bureau:      "EXPERIAN",
	}

	// Generate sample accounts based on credit score
	report.Accounts = r.generateSampleAccounts(creditScore.CreditScore)

	// Generate sample inquiries
	report.Inquiries = r.generateSampleInquiries(creditScore.CreditScore)

	// Generate payment history
	report.PaymentHistory = r.generatePaymentHistory(creditScore.CreditScore)

	return report
}

// generateFactorCodes generates factor codes based on credit score
func (r *CreditBureauRepository) generateFactorCodes(score int) []string {
	var codes []string

	switch {
	case score >= 750:
		codes = []string{"01", "02"} // Excellent credit factors
	case score >= 700:
		codes = []string{"03", "04", "05"} // Good credit factors
	case score >= 650:
		codes = []string{"06", "07", "08", "09"} // Fair credit factors
	case score >= 600:
		codes = []string{"10", "11", "12", "13", "14"} // Poor credit factors
	default:
		codes = []string{"15", "16", "17", "18", "19", "20"} // Very poor credit factors
	}

	return codes
}

// categorizeRiskLevel categorizes risk based on credit score
func (r *CreditBureauRepository) categorizeRiskLevel(score int) string {
	switch {
	case score >= 750:
		return "EXCELLENT"
	case score >= 700:
		return "GOOD"
	case score >= 650:
		return "FAIR"
	case score >= 600:
		return "POOR"
	default:
		return "VERY_POOR"
	}
}

// generateSampleAccounts generates sample credit accounts
func (r *CreditBureauRepository) generateSampleAccounts(creditScore int) []domain.CreditAccount {
	var accounts []domain.CreditAccount

	// Number of accounts varies by credit score
	accountCount := 3
	if creditScore >= 700 {
		accountCount = 5
	} else if creditScore < 600 {
		accountCount = 2
	}

	for i := 0; i < accountCount; i++ {
		account := domain.CreditAccount{
			AccountID:      fmt.Sprintf("ACC-%d-%d", creditScore, i+1),
			AccountType:    r.getAccountType(i),
			Creditor:       r.getCreditorName(i),
			OpenDate:       time.Now().AddDate(-2-i, 0, 0),
			Balance:        r.generateBalance(creditScore, i),
			CreditLimit:    r.generateCreditLimit(creditScore, i),
			PaymentStatus:  r.getPaymentStatus(creditScore),
			MonthsReviewed: 24,
		}

		// Calculate utilization
		if account.CreditLimit > 0 {
			account.Utilization = (account.Balance / account.CreditLimit) * 100
		}

		accounts = append(accounts, account)
	}

	return accounts
}

// Helper functions for account generation
func (r *CreditBureauRepository) getAccountType(index int) string {
	types := []string{"CREDIT_CARD", "AUTO_LOAN", "MORTGAGE", "PERSONAL_LOAN", "STUDENT_LOAN"}
	return types[index%len(types)]
}

func (r *CreditBureauRepository) getCreditorName(index int) string {
	creditors := []string{"Chase Bank", "Bank of America", "Wells Fargo", "Citibank", "Capital One"}
	return creditors[index%len(creditors)]
}

func (r *CreditBureauRepository) generateBalance(creditScore, index int) float64 {
	baseFactor := float64(creditScore) / 100.0
	return float64(index+1) * 1000.0 * baseFactor
}

func (r *CreditBureauRepository) generateCreditLimit(creditScore, index int) float64 {
	if index == 0 { // Credit card
		switch {
		case creditScore >= 750:
			return 15000.0
		case creditScore >= 700:
			return 10000.0
		case creditScore >= 650:
			return 5000.0
		default:
			return 2000.0
		}
	}
	return 0.0 // Non-revolving accounts
}

func (r *CreditBureauRepository) getPaymentStatus(creditScore int) string {
	if creditScore >= 700 {
		return "CURRENT"
	} else if creditScore >= 600 {
		return "30_DAYS_LATE"
	} else {
		return "60_DAYS_LATE"
	}
}

// generateSampleInquiries generates sample credit inquiries
func (r *CreditBureauRepository) generateSampleInquiries(creditScore int) []domain.CreditInquiry {
	var inquiries []domain.CreditInquiry

	// Fewer inquiries for higher credit scores
	inquiryCount := 3
	if creditScore >= 750 {
		inquiryCount = 1
	} else if creditScore < 600 {
		inquiryCount = 5
	}

	for i := 0; i < inquiryCount; i++ {
		inquiry := domain.CreditInquiry{
			InquiryDate: time.Now().AddDate(0, -(i+1)*2, 0),
			Creditor:    r.getCreditorName(i),
			InquiryType: "HARD",
			Purpose:     r.getInquiryPurpose(i),
		}
		inquiries = append(inquiries, inquiry)
	}

	return inquiries
}

func (r *CreditBureauRepository) getInquiryPurpose(index int) string {
	purposes := []string{"CREDIT_CARD", "AUTO_LOAN", "PERSONAL_LOAN", "MORTGAGE", "BUSINESS_LOAN"}
	return purposes[index%len(purposes)]
}

// generatePaymentHistory generates payment history data
func (r *CreditBureauRepository) generatePaymentHistory(creditScore int) domain.PaymentHistory {
	var history domain.PaymentHistory

	// Payment history varies by credit score
	switch {
	case creditScore >= 750:
		history = domain.PaymentHistory{
			OnTimePayments: 95,
			LatePayments:   2,
			Defaults:       0,
			Bankruptcies:   0,
			CreditAge:      120, // 10 years
			PaymentScore:   0.95,
		}
	case creditScore >= 700:
		history = domain.PaymentHistory{
			OnTimePayments: 85,
			LatePayments:   8,
			Defaults:       0,
			Bankruptcies:   0,
			CreditAge:      84, // 7 years
			PaymentScore:   0.85,
		}
	case creditScore >= 650:
		history = domain.PaymentHistory{
			OnTimePayments: 75,
			LatePayments:   15,
			Defaults:       1,
			Bankruptcies:   0,
			CreditAge:      60, // 5 years
			PaymentScore:   0.70,
		}
	case creditScore >= 600:
		history = domain.PaymentHistory{
			OnTimePayments: 65,
			LatePayments:   20,
			Defaults:       2,
			Bankruptcies:   0,
			CreditAge:      48, // 4 years
			PaymentScore:   0.60,
		}
	default:
		history = domain.PaymentHistory{
			OnTimePayments: 50,
			LatePayments:   30,
			Defaults:       5,
			Bankruptcies:   1,
			CreditAge:      36, // 3 years
			PaymentScore:   0.40,
		}
	}

	return history
}

// maskSSN masks SSN for logging (shows only last 4 digits)
func maskSSN(ssn string) string {
	if len(ssn) < 4 {
		return "***"
	}
	return "***-**-" + ssn[len(ssn)-4:]
}

// ValidateSSN validates SSN format
func ValidateSSN(ssn string) error {
	// Remove any non-digit characters
	cleaned := ""
	for _, char := range ssn {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}

	if len(cleaned) != 9 {
		return fmt.Errorf("SSN must be exactly 9 digits")
	}

	// Check for invalid SSN patterns
	invalidPatterns := []string{
		"000000000", "111111111", "222222222", "333333333",
		"444444444", "555555555", "666666666", "777777777",
		"888888888", "999999999", "123456789",
	}

	for _, pattern := range invalidPatterns {
		if cleaned == pattern {
			return fmt.Errorf("invalid SSN pattern")
		}
	}

	return nil
}
