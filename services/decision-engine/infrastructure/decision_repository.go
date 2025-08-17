package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huuhoait/los-demo/services/decision-engine/domain"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

// DecisionRepository implements decision data persistence
type DecisionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDecisionRepository creates a new decision repository
func NewDecisionRepository(db *sql.DB, logger *zap.Logger) *DecisionRepository {
	return &DecisionRepository{
		db:     db,
		logger: logger,
	}
}

// SaveDecision saves a decision to the database
func (r *DecisionRepository) SaveDecision(ctx context.Context, decision *domain.DecisionResponse) error {
	logger := r.logger.With(
		zap.String("application_id", decision.ApplicationID),
		zap.String("operation", "save_decision"),
	)

	logger.Info("Saving decision to database")

	// Serialize complex objects to JSON
	riskAssessmentJSON, err := json.Marshal(decision.RiskAssessment)
	if err != nil {
		logger.Error("Failed to marshal risk assessment", zap.Error(err))
		return fmt.Errorf("failed to marshal risk assessment: %w", err)
	}

	appliedRulesJSON, err := json.Marshal(decision.AppliedRules)
	if err != nil {
		logger.Error("Failed to marshal applied rules", zap.Error(err))
		return fmt.Errorf("failed to marshal applied rules: %w", err)
	}

	recommendationsJSON, err := json.Marshal(decision.Recommendations)
	if err != nil {
		logger.Error("Failed to marshal recommendations", zap.Error(err))
		return fmt.Errorf("failed to marshal recommendations: %w", err)
	}

	// Insert decision record
	query := `
		INSERT INTO decisions (
			application_id, decision, confidence_score, interest_rate, 
			max_amount, reason, risk_assessment, applied_rules, 
			recommendations, decision_date, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id`

	var decisionID int64
	err = r.db.QueryRowContext(ctx, query,
		decision.ApplicationID,
		decision.Decision,
		decision.ConfidenceScore,
		decision.InterestRate,
		decision.MaxAmount,
		decision.Reason,
		riskAssessmentJSON,
		appliedRulesJSON,
		recommendationsJSON,
		decision.DecisionDate,
		time.Now(),
	).Scan(&decisionID)

	if err != nil {
		logger.Error("Failed to save decision", zap.Error(err))
		return fmt.Errorf("failed to save decision: %w", err)
	}

	logger.Info("Decision saved successfully", zap.Int64("decision_id", decisionID))
	return nil
}

// GetDecisionByApplicationID retrieves a decision by application ID
func (r *DecisionRepository) GetDecisionByApplicationID(ctx context.Context, applicationID string) (*domain.DecisionResponse, error) {
	logger := r.logger.With(
		zap.String("application_id", applicationID),
		zap.String("operation", "get_decision"),
	)

	logger.Info("Retrieving decision from database")

	query := `
		SELECT application_id, decision, confidence_score, interest_rate,
			   max_amount, reason, risk_assessment, applied_rules,
			   recommendations, decision_date, created_at
		FROM decisions 
		WHERE application_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1`

	var decision domain.DecisionResponse
	var riskAssessmentJSON, appliedRulesJSON, recommendationsJSON []byte
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, applicationID).Scan(
		&decision.ApplicationID,
		&decision.Decision,
		&decision.ConfidenceScore,
		&decision.InterestRate,
		&decision.MaxAmount,
		&decision.Reason,
		&riskAssessmentJSON,
		&appliedRulesJSON,
		&recommendationsJSON,
		&decision.DecisionDate,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("No decision found for application")
			return nil, fmt.Errorf("decision not found for application %s", applicationID)
		}
		logger.Error("Failed to retrieve decision", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve decision: %w", err)
	}

	// Deserialize JSON fields
	if err := json.Unmarshal(riskAssessmentJSON, &decision.RiskAssessment); err != nil {
		logger.Error("Failed to unmarshal risk assessment", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal risk assessment: %w", err)
	}

	if err := json.Unmarshal(appliedRulesJSON, &decision.AppliedRules); err != nil {
		logger.Error("Failed to unmarshal applied rules", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal applied rules: %w", err)
	}

	if err := json.Unmarshal(recommendationsJSON, &decision.Recommendations); err != nil {
		logger.Error("Failed to unmarshal recommendations", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal recommendations: %w", err)
	}

	logger.Info("Decision retrieved successfully")
	return &decision, nil
}

// SaveDecisionRequest saves the original decision request
func (r *DecisionRepository) SaveDecisionRequest(ctx context.Context, request *domain.DecisionRequest) error {
	logger := r.logger.With(
		zap.String("application_id", request.ApplicationID),
		zap.String("operation", "save_decision_request"),
	)

	logger.Info("Saving decision request to database")

	query := `
		INSERT INTO decision_requests (
			application_id, customer_id, loan_amount, loan_purpose, 
			loan_term_months, annual_income, monthly_income, credit_score,
			employment_type, requested_amount, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id`

	var requestID int64
	err := r.db.QueryRowContext(ctx, query,
		request.ApplicationID,
		request.CustomerID,
		request.LoanAmount,
		request.LoanPurpose,
		request.LoanTermMonths,
		request.AnnualIncome,
		request.MonthlyIncome,
		request.CreditScore,
		request.EmploymentType,
		request.LoanAmount, // requested_amount same as loan_amount for now
		time.Now(),
	).Scan(&requestID)

	if err != nil {
		logger.Error("Failed to save decision request", zap.Error(err))
		return fmt.Errorf("failed to save decision request: %w", err)
	}

	logger.Info("Decision request saved successfully", zap.Int64("request_id", requestID))
	return nil
}

// GetDecisionHistory retrieves decision history for a customer
func (r *DecisionRepository) GetDecisionHistory(ctx context.Context, customerID string, limit int) ([]*domain.DecisionResponse, error) {
	logger := r.logger.With(
		zap.String("customer_id", customerID),
		zap.Int("limit", limit),
		zap.String("operation", "get_decision_history"),
	)

	logger.Info("Retrieving decision history")

	query := `
		SELECT d.application_id, d.decision, d.confidence_score, d.interest_rate,
			   d.max_amount, d.reason, d.risk_assessment, d.applied_rules,
			   d.recommendations, d.decision_date, d.created_at
		FROM decisions d
		JOIN decision_requests dr ON d.application_id = dr.application_id
		WHERE dr.customer_id = $1
		ORDER BY d.created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, customerID, limit)
	if err != nil {
		logger.Error("Failed to query decision history", zap.Error(err))
		return nil, fmt.Errorf("failed to query decision history: %w", err)
	}
	defer rows.Close()

	var decisions []*domain.DecisionResponse

	for rows.Next() {
		var decision domain.DecisionResponse
		var riskAssessmentJSON, appliedRulesJSON, recommendationsJSON []byte
		var createdAt time.Time

		err := rows.Scan(
			&decision.ApplicationID,
			&decision.Decision,
			&decision.ConfidenceScore,
			&decision.InterestRate,
			&decision.MaxAmount,
			&decision.Reason,
			&riskAssessmentJSON,
			&appliedRulesJSON,
			&recommendationsJSON,
			&decision.DecisionDate,
			&createdAt,
		)

		if err != nil {
			logger.Error("Failed to scan decision row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan decision row: %w", err)
		}

		// Deserialize JSON fields
		if err := json.Unmarshal(riskAssessmentJSON, &decision.RiskAssessment); err != nil {
			logger.Error("Failed to unmarshal risk assessment", zap.Error(err))
			continue // Skip this record but continue with others
		}

		if err := json.Unmarshal(appliedRulesJSON, &decision.AppliedRules); err != nil {
			logger.Error("Failed to unmarshal applied rules", zap.Error(err))
			continue
		}

		if err := json.Unmarshal(recommendationsJSON, &decision.Recommendations); err != nil {
			logger.Error("Failed to unmarshal recommendations", zap.Error(err))
			continue
		}

		decisions = append(decisions, &decision)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating over decision rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating over decision rows: %w", err)
	}

	logger.Info("Decision history retrieved successfully", zap.Int("count", len(decisions)))
	return decisions, nil
}

// GetDecisionStatistics retrieves statistics about decisions
func (r *DecisionRepository) GetDecisionStatistics(ctx context.Context, dateFrom, dateTo time.Time) (*domain.DecisionStatistics, error) {
	logger := r.logger.With(
		zap.Time("date_from", dateFrom),
		zap.Time("date_to", dateTo),
		zap.String("operation", "get_decision_statistics"),
	)

	logger.Info("Retrieving decision statistics")

	query := `
		SELECT 
			COUNT(*) as total_decisions,
			COUNT(CASE WHEN decision = 'APPROVED' THEN 1 END) as approved_count,
			COUNT(CASE WHEN decision = 'DECLINED' THEN 1 END) as declined_count,
			COUNT(CASE WHEN decision = 'CONDITIONAL' THEN 1 END) as conditional_count,
			AVG(CASE WHEN decision = 'APPROVED' THEN confidence_score END) as avg_approval_confidence,
			AVG(CASE WHEN decision = 'APPROVED' THEN interest_rate END) as avg_interest_rate,
			AVG(CASE WHEN decision = 'APPROVED' THEN max_amount END) as avg_approved_amount
		FROM decisions 
		WHERE decision_date BETWEEN $1 AND $2`

	var stats domain.DecisionStatistics
	var avgApprovalConfidence, avgInterestRate, avgApprovedAmount sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, dateFrom, dateTo).Scan(
		&stats.TotalDecisions,
		&stats.ApprovedCount,
		&stats.DeclinedCount,
		&stats.ConditionalCount,
		&avgApprovalConfidence,
		&avgInterestRate,
		&avgApprovedAmount,
	)

	if err != nil {
		logger.Error("Failed to retrieve decision statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve decision statistics: %w", err)
	}

	// Handle null values
	if avgApprovalConfidence.Valid {
		stats.AvgApprovalConfidence = avgApprovalConfidence.Float64
	}
	if avgInterestRate.Valid {
		stats.AvgInterestRate = avgInterestRate.Float64
	}
	if avgApprovedAmount.Valid {
		stats.AvgApprovedAmount = avgApprovedAmount.Float64
	}

	// Calculate approval rate
	if stats.TotalDecisions > 0 {
		stats.ApprovalRate = float64(stats.ApprovedCount) / float64(stats.TotalDecisions) * 100
	}

	logger.Info("Decision statistics retrieved successfully",
		zap.Int("total_decisions", stats.TotalDecisions),
		zap.Float64("approval_rate", stats.ApprovalRate),
	)

	return &stats, nil
}

// InitializeDatabase creates necessary tables
func (r *DecisionRepository) InitializeDatabase(ctx context.Context) error {
	logger := r.logger.With(zap.String("operation", "initialize_database"))

	logger.Info("Initializing database tables")

	// Create decision_requests table
	requestsQuery := `
		CREATE TABLE IF NOT EXISTS decision_requests (
			id BIGSERIAL PRIMARY KEY,
			application_id VARCHAR(255) UNIQUE NOT NULL,
			customer_id VARCHAR(255) NOT NULL,
			loan_amount DECIMAL(15,2) NOT NULL,
			loan_purpose VARCHAR(100) NOT NULL,
			loan_term_months INTEGER NOT NULL,
			annual_income DECIMAL(15,2) NOT NULL,
			monthly_income DECIMAL(15,2) NOT NULL,
			credit_score INTEGER NOT NULL,
			employment_type VARCHAR(50) NOT NULL,
			requested_amount DECIMAL(15,2) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			INDEX idx_customer_id (customer_id),
			INDEX idx_application_id (application_id),
			INDEX idx_created_at (created_at)
		)`

	// Create decisions table
	decisionsQuery := `
		CREATE TABLE IF NOT EXISTS decisions (
			id BIGSERIAL PRIMARY KEY,
			application_id VARCHAR(255) NOT NULL,
			decision VARCHAR(20) NOT NULL,
			confidence_score DECIMAL(5,4) NOT NULL,
			interest_rate DECIMAL(5,4),
			max_amount DECIMAL(15,2),
			reason TEXT,
			risk_assessment JSONB,
			applied_rules JSONB,
			recommendations JSONB,
			decision_date TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			INDEX idx_application_id (application_id),
			INDEX idx_decision (decision),
			INDEX idx_decision_date (decision_date),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (application_id) REFERENCES decision_requests(application_id)
		)`

	// Execute table creation
	if _, err := r.db.ExecContext(ctx, requestsQuery); err != nil {
		logger.Error("Failed to create decision_requests table", zap.Error(err))
		return fmt.Errorf("failed to create decision_requests table: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, decisionsQuery); err != nil {
		logger.Error("Failed to create decisions table", zap.Error(err))
		return fmt.Errorf("failed to create decisions table: %w", err)
	}

	logger.Info("Database tables initialized successfully")
	return nil
}

// GetDecision implements the domain.DecisionRepository interface
func (r *DecisionRepository) GetDecision(applicationID string) (*domain.DecisionResponse, error) {
	return r.GetDecisionByApplicationID(context.Background(), applicationID)
}

// GetDecisionHistoryByUser implements the domain.DecisionRepository interface
func (r *DecisionRepository) GetDecisionHistoryByUser(userID string) ([]domain.DecisionResponse, error) {
	decisions, err := r.GetDecisionHistory(context.Background(), userID, 100)
	if err != nil {
		return nil, err
	}

	result := make([]domain.DecisionResponse, len(decisions))
	for i, decision := range decisions {
		if decision != nil {
			result[i] = *decision
		}
	}

	return result, nil
}

// UpdateDecision implements the domain.DecisionRepository interface
func (r *DecisionRepository) UpdateDecision(response *domain.DecisionResponse) error {
	// Implementation would update an existing decision
	// For now, just save it as new
	return r.SaveDecision(context.Background(), response)
}
