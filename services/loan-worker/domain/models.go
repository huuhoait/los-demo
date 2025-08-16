package domain

import (
	"time"
)

// Error codes for loan service
const (
	// Loan management errors
	LOAN_001 = "LOAN_001" // Invalid loan amount
	LOAN_002 = "LOAN_002" // Invalid loan purpose
	LOAN_003 = "LOAN_003" // Invalid loan term
	LOAN_004 = "LOAN_004" // Invalid income information
	LOAN_005 = "LOAN_005" // Loan amount below minimum
	LOAN_006 = "LOAN_006" // Loan amount above maximum
	LOAN_007 = "LOAN_007" // Insufficient income
	LOAN_008 = "LOAN_008" // Invalid state transition
	LOAN_009 = "LOAN_009" // Offer expired
	LOAN_010 = "LOAN_010" // Application not found
	LOAN_011 = "LOAN_011" // Workflow start failed
	LOAN_012 = "LOAN_012" // Workflow execution failed
	LOAN_013 = "LOAN_013" // State conflict
	LOAN_014 = "LOAN_014" // Conductor unavailable
	LOAN_015 = "LOAN_015" // Decision engine error
	LOAN_016 = "LOAN_016" // State machine error
	LOAN_017 = "LOAN_017" // Offer calculation error
	LOAN_018 = "LOAN_018" // Application validation failed
	LOAN_019 = "LOAN_019" // Invalid application status
	LOAN_020 = "LOAN_020" // Invalid request format
	LOAN_021 = "LOAN_021" // User not found
	LOAN_022 = "LOAN_022" // Unauthorized access
	LOAN_023 = "LOAN_023" // Database error
	LOAN_024 = "LOAN_024" // External service error
	LOAN_025 = "LOAN_025" // Document verification required
	LOAN_026 = "LOAN_026" // Credit check failed
	LOAN_027 = "LOAN_027" // KYC verification pending
	LOAN_028 = "LOAN_028" // Manual review required
	LOAN_029 = "LOAN_029" // Application already exists
	LOAN_030 = "LOAN_030" // Invalid offer terms
)

// ApplicationState represents the state of a loan application
type ApplicationState string

const (
	StateInitiated          ApplicationState = "initiated"
	StatePreQualified       ApplicationState = "pre_qualified"
	StateDocumentsSubmitted ApplicationState = "documents_submitted"
	StateIdentityVerified   ApplicationState = "identity_verified"
	StateUnderwriting       ApplicationState = "underwriting"
	StateManualReview       ApplicationState = "manual_review"
	StateApproved           ApplicationState = "approved"
	StateDenied             ApplicationState = "denied"
	StateDocumentsSigned    ApplicationState = "documents_signed"
	StateFunded             ApplicationState = "funded"
	StateActive             ApplicationState = "active"
	StateClosed             ApplicationState = "closed"
)

// ApplicationStatus represents the status of a loan application
type ApplicationStatus string

const (
	StatusDraft       ApplicationStatus = "draft"
	StatusSubmitted   ApplicationStatus = "submitted"
	StatusUnderReview ApplicationStatus = "under_review"
	StatusApproved    ApplicationStatus = "approved"
	StatusDenied      ApplicationStatus = "denied"
	StatusFunded      ApplicationStatus = "funded"
	StatusActive      ApplicationStatus = "active"
	StatusClosed      ApplicationStatus = "closed"
)

// LoanPurpose represents the purpose of the loan
type LoanPurpose string

const (
	PurposeDebtConsolidation LoanPurpose = "debt_consolidation"
	PurposeHomeImprovement   LoanPurpose = "home_improvement"
	PurposeMedical           LoanPurpose = "medical"
	PurposeVacation          LoanPurpose = "vacation"
	PurposeWedding           LoanPurpose = "wedding"
	PurposeMajorPurchase     LoanPurpose = "major_purchase"
	PurposeOther             LoanPurpose = "other"
)

// EmploymentStatus represents employment status
type EmploymentStatus string

const (
	EmploymentFullTime     EmploymentStatus = "full_time"
	EmploymentPartTime     EmploymentStatus = "part_time"
	EmploymentSelfEmployed EmploymentStatus = "self_employed"
	EmploymentRetired      EmploymentStatus = "retired"
	EmploymentUnemployed   EmploymentStatus = "unemployed"
	EmploymentStudent      EmploymentStatus = "student"
)

// ResidenceType represents the type of residence
type ResidenceType string

const (
	ResidenceOwn    ResidenceType = "own"
	ResidenceRent   ResidenceType = "rent"
	ResidenceFamily ResidenceType = "family"
	ResidenceOther  ResidenceType = "other"
)

// AccountType represents the type of bank account
type AccountType string

const (
	AccountChecking AccountType = "checking"
	AccountSavings  AccountType = "savings"
)

// LoanApplication represents a loan application
type LoanApplication struct {
	ID                string            `json:"id" db:"id"`
	UserID            string            `json:"user_id" db:"user_id"`
	ApplicationNumber string            `json:"application_number" db:"application_number"`
	LoanAmount        float64           `json:"loan_amount" db:"loan_amount"`
	LoanPurpose       LoanPurpose       `json:"loan_purpose" db:"loan_purpose"`
	RequestedTerm     int               `json:"requested_term_months" db:"requested_term_months"`
	AnnualIncome      float64           `json:"annual_income" db:"annual_income"`
	MonthlyIncome     float64           `json:"monthly_income" db:"monthly_income"`
	EmploymentStatus  EmploymentStatus  `json:"employment_status" db:"employment_status"`
	MonthlyDebt       float64           `json:"monthly_debt_payments" db:"monthly_debt_payments"`
	CurrentState      ApplicationState  `json:"current_state" db:"current_state"`
	Status            ApplicationStatus `json:"status" db:"status"`
	RiskScore         *int              `json:"risk_score" db:"risk_score"`
	WorkflowID        *string           `json:"workflow_id" db:"workflow_id"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

// LoanOffer represents a loan offer
type LoanOffer struct {
	ID             string    `json:"id" db:"id"`
	ApplicationID  string    `json:"application_id" db:"application_id"`
	OfferAmount    float64   `json:"offer_amount" db:"offer_amount"`
	InterestRate   float64   `json:"interest_rate" db:"interest_rate"`
	TermMonths     int       `json:"term_months" db:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment" db:"monthly_payment"`
	TotalInterest  float64   `json:"total_interest" db:"total_interest"`
	APR            float64   `json:"apr" db:"apr"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// StateTransition represents a state transition in the application workflow
type StateTransition struct {
	ID               string                 `json:"id" db:"id"`
	ApplicationID    string                 `json:"application_id" db:"application_id"`
	FromState        *ApplicationState      `json:"from_state" db:"from_state"`
	ToState          ApplicationState       `json:"to_state" db:"to_state"`
	TransitionReason string                 `json:"transition_reason" db:"transition_reason"`
	Automated        bool                   `json:"automated" db:"automated"`
	UserID           *string                `json:"user_id" db:"user_id"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// WorkflowExecution represents a workflow execution
type WorkflowExecution struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	ApplicationID string                 `json:"application_id"`
	Status        string                 `json:"status"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// User represents user information for loan applications
type User struct {
	ID             string         `json:"id,omitempty" db:"id"`
	FirstName      string         `json:"first_name" binding:"required" example:"John"`
	LastName       string         `json:"last_name" binding:"required" example:"Doe"`
	Email          string         `json:"email" binding:"required,email" example:"john.doe@example.com"`
	PhoneNumber    string         `json:"phone_number" binding:"required" example:"+1234567890"`
	DateOfBirth    time.Time      `json:"date_of_birth" binding:"required" example:"1990-01-01"`
	SSN            string         `json:"ssn" binding:"required,len=9" example:"123456789"`
	Address        Address        `json:"address" binding:"required"`
	EmploymentInfo EmploymentInfo `json:"employment_info" binding:"required"`
	BankingInfo    BankingInfo    `json:"banking_info" binding:"required"`
	CreatedAt      time.Time      `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at,omitempty" db:"updated_at"`
}

// Address represents user's address information
type Address struct {
	StreetAddress string        `json:"street_address" binding:"required" example:"123 Main St"`
	City          string        `json:"city" binding:"required" example:"New York"`
	State         string        `json:"state" binding:"required" example:"NY"`
	ZipCode       string        `json:"zip_code" binding:"required" example:"10001"`
	Country       string        `json:"country" binding:"required" example:"USA"`
	ResidenceType ResidenceType `json:"residence_type" binding:"required" example:"own"`
	TimeAtAddress int           `json:"time_at_address_months" binding:"required,min=0" example:"24"` // months
}

// EmploymentInfo represents user's employment information
type EmploymentInfo struct {
	EmployerName string `json:"employer_name" binding:"required" example:"ABC Company"`
	JobTitle     string `json:"job_title" binding:"required" example:"Software Engineer"`
	TimeEmployed int    `json:"time_employed_months" binding:"required,min=0" example:"36"` // months
	WorkPhone    string `json:"work_phone" binding:"required" example:"+1234567890"`
	WorkEmail    string `json:"work_email" binding:"omitempty,email" example:"john.doe@abccompany.com"`
}

// BankingInfo represents user's banking information
type BankingInfo struct {
	BankName      string      `json:"bank_name" binding:"required" example:"Chase Bank"`
	AccountType   AccountType `json:"account_type" binding:"required" example:"checking"`
	AccountNumber string      `json:"account_number" binding:"required" example:"1234567890"`
	RoutingNumber string      `json:"routing_number" binding:"required" example:"021000021"`
}

// CreateApplicationRequest represents a request to create a loan application
// @Description Request to create a new loan application with user information
type CreateApplicationRequest struct {
	// User information for application
	User User `json:"user" binding:"required"`

	// Loan application details
	LoanAmount       float64          `json:"loan_amount" binding:"required,min=5000,max=50000" example:"25000" minimum:"5000" maximum:"50000"`
	LoanPurpose      LoanPurpose      `json:"loan_purpose" binding:"required" example:"debt_consolidation"`
	RequestedTerm    int              `json:"requested_term_months" binding:"required,min=12,max=84" example:"60" minimum:"12" maximum:"84"`
	AnnualIncome     float64          `json:"annual_income" binding:"required,min=0" example:"75000" minimum:"0"`
	MonthlyIncome    float64          `json:"monthly_income" binding:"required,min=0" example:"6250" minimum:"0"`
	EmploymentStatus EmploymentStatus `json:"employment_status" binding:"required" example:"full_time"`
	MonthlyDebt      float64          `json:"monthly_debt_payments" binding:"min=0" example:"1500" minimum:"0"`
}

// UpdateApplicationRequest represents a request to update a loan application
type UpdateApplicationRequest struct {
	LoanAmount       *float64          `json:"loan_amount,omitempty" binding:"omitempty,min=5000,max=50000"`
	LoanPurpose      *LoanPurpose      `json:"loan_purpose,omitempty"`
	RequestedTerm    *int              `json:"requested_term_months,omitempty" binding:"omitempty,min=12,max=84"`
	AnnualIncome     *float64          `json:"annual_income,omitempty" binding:"omitempty,min=0"`
	MonthlyIncome    *float64          `json:"monthly_income,omitempty" binding:"omitempty,min=0"`
	EmploymentStatus *EmploymentStatus `json:"employment_status,omitempty"`
	MonthlyDebt      *float64          `json:"monthly_debt_payments,omitempty" binding:"omitempty,min=0"`
}

// PreQualifyRequest represents a pre-qualification request
// @Description Request to perform loan pre-qualification
type PreQualifyRequest struct {
	LoanAmount       float64          `json:"loan_amount" binding:"required,min=5000,max=50000" example:"25000" minimum:"5000" maximum:"50000"`
	AnnualIncome     float64          `json:"annual_income" binding:"required,min=0" example:"75000" minimum:"0"`
	MonthlyDebt      float64          `json:"monthly_debt_payments" binding:"min=0" example:"1500" minimum:"0"`
	EmploymentStatus EmploymentStatus `json:"employment_status" binding:"required" example:"full_time"`
}

// PreQualifyResult represents a pre-qualification result
// @Description Result of loan pre-qualification
type PreQualifyResult struct {
	Qualified        bool    `json:"qualified" example:"true"`
	MaxLoanAmount    float64 `json:"max_loan_amount" example:"50000"`
	MinInterestRate  float64 `json:"min_interest_rate" example:"8.5"`
	MaxInterestRate  float64 `json:"max_interest_rate" example:"12.0"`
	RecommendedTerms []int   `json:"recommended_terms"`
	DTIRatio         float64 `json:"dti_ratio" example:"0.24"`
	Message          string  `json:"message" example:"You are pre-qualified for a loan"`
}

// AcceptOfferRequest represents a request to accept a loan offer
type AcceptOfferRequest struct {
	OfferID string `json:"offer_id" binding:"required"`
}

// LoanError represents a domain-specific error
type LoanError struct {
	Code        string
	Message     string
	Description string
	HTTPStatus  int
}

func (e *LoanError) Error() string {
	return e.Message
}

// NewLoanError creates a new loan error
func NewLoanError(code, message, description string, httpStatus int) *LoanError {
	return &LoanError{
		Code:        code,
		Message:     message,
		Description: description,
		HTTPStatus:  httpStatus,
	}
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors map[string]string `json:"errors,omitempty"`
}

// Validate validates a create application request
func (req *CreateApplicationRequest) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string]string),
	}

	// Validate user information
	if req.User.FirstName == "" {
		result.Valid = false
		result.Errors["user.first_name"] = LOAN_020
	}
	if req.User.LastName == "" {
		result.Valid = false
		result.Errors["user.last_name"] = LOAN_020
	}
	if req.User.Email == "" {
		result.Valid = false
		result.Errors["user.email"] = LOAN_020
	}
	if req.User.PhoneNumber == "" {
		result.Valid = false
		result.Errors["user.phone_number"] = LOAN_020
	}
	if req.User.SSN == "" || len(req.User.SSN) != 9 {
		result.Valid = false
		result.Errors["user.ssn"] = LOAN_020
	}

	// Validate address
	if req.User.Address.StreetAddress == "" {
		result.Valid = false
		result.Errors["user.address.street_address"] = LOAN_020
	}
	if req.User.Address.City == "" {
		result.Valid = false
		result.Errors["user.address.city"] = LOAN_020
	}
	if req.User.Address.State == "" {
		result.Valid = false
		result.Errors["user.address.state"] = LOAN_020
	}
	if req.User.Address.ZipCode == "" {
		result.Valid = false
		result.Errors["user.address.zip_code"] = LOAN_020
	}
	if req.User.Address.TimeAtAddress < 0 {
		result.Valid = false
		result.Errors["user.address.time_at_address_months"] = LOAN_020
	}

	// Validate employment info
	if req.User.EmploymentInfo.EmployerName == "" {
		result.Valid = false
		result.Errors["user.employment_info.employer_name"] = LOAN_020
	}
	if req.User.EmploymentInfo.JobTitle == "" {
		result.Valid = false
		result.Errors["user.employment_info.job_title"] = LOAN_020
	}
	if req.User.EmploymentInfo.TimeEmployed < 0 {
		result.Valid = false
		result.Errors["user.employment_info.time_employed_months"] = LOAN_020
	}
	if req.User.EmploymentInfo.WorkPhone == "" {
		result.Valid = false
		result.Errors["user.employment_info.work_phone"] = LOAN_020
	}

	// Validate banking info
	if req.User.BankingInfo.BankName == "" {
		result.Valid = false
		result.Errors["user.banking_info.bank_name"] = LOAN_020
	}
	if req.User.BankingInfo.AccountNumber == "" {
		result.Valid = false
		result.Errors["user.banking_info.account_number"] = LOAN_020
	}
	if req.User.BankingInfo.RoutingNumber == "" {
		result.Valid = false
		result.Errors["user.banking_info.routing_number"] = LOAN_020
	}

	// Validate loan amount
	if req.LoanAmount < 5000 {
		result.Valid = false
		result.Errors["loan_amount"] = LOAN_005
	} else if req.LoanAmount > 50000 {
		result.Valid = false
		result.Errors["loan_amount"] = LOAN_006
	}

	// Validate term
	if req.RequestedTerm < 12 || req.RequestedTerm > 84 {
		result.Valid = false
		result.Errors["requested_term_months"] = LOAN_003
	}

	// Validate income
	if req.AnnualIncome <= 0 {
		result.Valid = false
		result.Errors["annual_income"] = LOAN_004
	}

	if req.MonthlyIncome <= 0 {
		result.Valid = false
		result.Errors["monthly_income"] = LOAN_004
	}

	// Validate DTI ratio (monthly debt should not exceed 40% of monthly income)
	if req.MonthlyIncome > 0 {
		dtiRatio := req.MonthlyDebt / req.MonthlyIncome
		if dtiRatio > 0.4 {
			result.Valid = false
			result.Errors["monthly_debt_payments"] = LOAN_007
		}
	}

	return result
}

// ValidateUser validates user information
func (u *User) ValidateUser() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string]string),
	}

	// Basic validation
	if u.FirstName == "" {
		result.Valid = false
		result.Errors["first_name"] = LOAN_020
	}
	if u.LastName == "" {
		result.Valid = false
		result.Errors["last_name"] = LOAN_020
	}
	if u.Email == "" {
		result.Valid = false
		result.Errors["email"] = LOAN_020
	}
	if u.PhoneNumber == "" {
		result.Valid = false
		result.Errors["phone_number"] = LOAN_020
	}
	if u.SSN == "" || len(u.SSN) != 9 {
		result.Valid = false
		result.Errors["ssn"] = LOAN_020
	}

	// Validate date of birth (must be at least 18 years old)
	eighteenYearsAgo := time.Now().AddDate(-18, 0, 0)
	if u.DateOfBirth.After(eighteenYearsAgo) {
		result.Valid = false
		result.Errors["date_of_birth"] = LOAN_020
	}

	return result
}

// ValidateAddress validates address information
func (a *Address) ValidateAddress() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string]string),
	}

	if a.StreetAddress == "" {
		result.Valid = false
		result.Errors["street_address"] = LOAN_020
	}
	if a.City == "" {
		result.Valid = false
		result.Errors["city"] = LOAN_020
	}
	if a.State == "" {
		result.Valid = false
		result.Errors["state"] = LOAN_020
	}
	if a.ZipCode == "" {
		result.Valid = false
		result.Errors["zip_code"] = LOAN_020
	}
	if a.Country == "" {
		result.Valid = false
		result.Errors["country"] = LOAN_020
	}
	if a.TimeAtAddress < 0 {
		result.Valid = false
		result.Errors["time_at_address_months"] = LOAN_020
	}

	return result
}

// ValidateEmploymentInfo validates employment information
func (e *EmploymentInfo) ValidateEmploymentInfo() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string]string),
	}

	if e.EmployerName == "" {
		result.Valid = false
		result.Errors["employer_name"] = LOAN_020
	}
	if e.JobTitle == "" {
		result.Valid = false
		result.Errors["job_title"] = LOAN_020
	}
	if e.TimeEmployed < 0 {
		result.Valid = false
		result.Errors["time_employed_months"] = LOAN_020
	}
	if e.WorkPhone == "" {
		result.Valid = false
		result.Errors["work_phone"] = LOAN_020
	}

	return result
}

// ValidateBankingInfo validates banking information
func (b *BankingInfo) ValidateBankingInfo() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string]string),
	}

	if b.BankName == "" {
		result.Valid = false
		result.Errors["bank_name"] = LOAN_020
	}
	if b.AccountNumber == "" {
		result.Valid = false
		result.Errors["account_number"] = LOAN_020
	}
	if b.RoutingNumber == "" || len(b.RoutingNumber) != 9 {
		result.Valid = false
		result.Errors["routing_number"] = LOAN_020
	}

	return result
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// GetAge calculates and returns the user's age
func (u *User) GetAge() int {
	now := time.Now()
	age := now.Year() - u.DateOfBirth.Year()
	if now.YearDay() < u.DateOfBirth.YearDay() {
		age--
	}
	return age
}

// IsAdult checks if the user is at least 18 years old
func (u *User) IsAdult() bool {
	return u.GetAge() >= 18
}

// MaskSSN returns a masked version of the SSN for display
func (u *User) MaskSSN() string {
	if len(u.SSN) != 9 {
		return ""
	}
	return "***-**-" + u.SSN[5:]
}

// MaskAccountNumber returns a masked version of the account number for display
func (u *User) MaskAccountNumber() string {
	if len(u.BankingInfo.AccountNumber) < 4 {
		return ""
	}
	return "****" + u.BankingInfo.AccountNumber[len(u.BankingInfo.AccountNumber)-4:]
}

// CreateUserFromRequest creates a new user from the create application request
func (req *CreateApplicationRequest) CreateUserFromRequest() *User {
	return &req.User
}

// GetUserInfo returns user information for display purposes (with sensitive data masked)
func (req *CreateApplicationRequest) GetUserInfo() map[string]interface{} {
	user := req.User
	return map[string]interface{}{
		"full_name":       user.GetFullName(),
		"email":           user.Email,
		"phone_number":    user.PhoneNumber,
		"age":             user.GetAge(),
		"address":         user.Address,
		"employment_info": user.EmploymentInfo,
		"banking_info": map[string]interface{}{
			"bank_name":      user.BankingInfo.BankName,
			"account_type":   user.BankingInfo.AccountType,
			"account_number": user.MaskAccountNumber(),
			"routing_number": "****" + user.BankingInfo.RoutingNumber[len(user.BankingInfo.RoutingNumber)-4:],
		},
	}
}

// CalculateDTI calculates debt-to-income ratio
func (app *LoanApplication) CalculateDTI() float64 {
	if app.MonthlyIncome <= 0 {
		return 0
	}
	return app.MonthlyDebt / app.MonthlyIncome
}

// IsExpired checks if a loan offer has expired
func (offer *LoanOffer) IsExpired() bool {
	return time.Now().After(offer.ExpiresAt)
}

// CanTransitionTo checks if the application can transition to the given state
func (app *LoanApplication) CanTransitionTo(newState ApplicationState) bool {
	validTransitions := map[ApplicationState][]ApplicationState{
		StateInitiated:          {StatePreQualified},
		StatePreQualified:       {StateDocumentsSubmitted},
		StateDocumentsSubmitted: {StateIdentityVerified},
		StateIdentityVerified:   {StateUnderwriting},
		StateUnderwriting:       {StateApproved, StateDenied, StateManualReview},
		StateManualReview:       {StateApproved, StateDenied},
		StateApproved:           {StateDocumentsSigned},
		StateDocumentsSigned:    {StateFunded},
		StateFunded:             {StateActive},
		StateActive:             {StateClosed},
	}

	allowedStates, exists := validTransitions[app.CurrentState]
	if !exists {
		return false
	}

	for _, allowedState := range allowedStates {
		if allowedState == newState {
			return true
		}
	}

	return false
}
