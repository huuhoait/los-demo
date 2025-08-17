package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents the core user entity
type User struct {
	ID            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	Phone         string    `json:"phone" db:"phone"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	PhoneVerified bool      `json:"phone_verified" db:"phone_verified"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile represents the extended user profile information
type UserProfile struct {
	ID             string         `json:"id" db:"id"`
	UserID         string         `json:"user_id" db:"user_id"`
	FirstName      string         `json:"first_name" db:"first_name"`
	LastName       string         `json:"last_name" db:"last_name"`
	DateOfBirth    time.Time      `json:"date_of_birth" db:"date_of_birth"`
	SSNEncrypted   string         `json:"-" db:"ssn_encrypted"`
	Phone          string         `json:"phone" db:"phone"`
	Address        Address        `json:"address" db:"address"`
	EmploymentInfo EmploymentInfo `json:"employment_info" db:"employment_info"`
	FinancialInfo  FinancialInfo  `json:"financial_info" db:"financial_info"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// Address represents a physical address
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

// Value implements the driver.Valuer interface for database storage
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for database retrieval
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into Address", value)
	}

	return json.Unmarshal(bytes, a)
}

// EmploymentInfo represents employment details
type EmploymentInfo struct {
	EmployerName    string     `json:"employer_name"`
	JobTitle        string     `json:"job_title"`
	EmploymentType  string     `json:"employment_type"`
	StartDate       *time.Time `json:"start_date"`
	MonthlyIncome   float64    `json:"monthly_income"`
	EmployerPhone   string     `json:"employer_phone"`
	EmployerAddress Address    `json:"employer_address"`
}

// Value implements the driver.Valuer interface for database storage
func (e EmploymentInfo) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements the sql.Scanner interface for database retrieval
func (e *EmploymentInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into EmploymentInfo", value)
	}

	return json.Unmarshal(bytes, e)
}

// FinancialInfo represents financial details
type FinancialInfo struct {
	AnnualIncome        float64 `json:"annual_income"`
	MonthlyIncome       float64 `json:"monthly_income"`
	OtherIncome         float64 `json:"other_income"`
	MonthlyDebtPayments float64 `json:"monthly_debt_payments"`
	MonthlyRent         float64 `json:"monthly_rent"`
	SavingsAmount       float64 `json:"savings_amount"`
	CheckingAmount      float64 `json:"checking_amount"`
}

// Value implements the driver.Valuer interface for database storage
func (f FinancialInfo) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan implements the sql.Scanner interface for database retrieval
func (f *FinancialInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into FinancialInfo", value)
	}

	return json.Unmarshal(bytes, f)
}

// KYCVerification represents KYC verification data
type KYCVerification struct {
	ID                string                 `json:"id" db:"id"`
	UserID            string                 `json:"user_id" db:"user_id"`
	VerificationType  string                 `json:"verification_type" db:"verification_type"`
	Provider          string                 `json:"provider" db:"provider"`
	Status            KYCStatus              `json:"status" db:"status"`
	ProviderReference string                 `json:"provider_reference" db:"provider_reference"`
	VerificationData  map[string]interface{} `json:"verification_data" db:"verification_data"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// KYCStatus represents the status of KYC verification
type KYCStatus string

const (
	KYCStatusPending      KYCStatus = "pending"
	KYCStatusVerified     KYCStatus = "verified"
	KYCStatusFailed       KYCStatus = "failed"
	KYCStatusManualReview KYCStatus = "manual_review"
)

// Document represents a user-uploaded document
type Document struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	DocumentType  string    `json:"document_type" db:"document_type"`
	FilePath      string    `json:"file_path" db:"file_path"`
	FileSize      int64     `json:"file_size" db:"file_size"`
	MimeType      string    `json:"mime_type" db:"mime_type"`
	EncryptionKey string    `json:"-" db:"encryption_key"`
	UploadIP      string    `json:"upload_ip" db:"upload_ip"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// DocumentType constants
const (
	DocumentTypeDriversLicense = "drivers_license"
	DocumentTypePassport       = "passport"
	DocumentTypePayStub        = "pay_stub"
	DocumentTypeBankStatement  = "bank_statement"
	DocumentTypeUtilityBill    = "utility_bill"
	DocumentTypeW2             = "w2"
	DocumentType1099           = "1099"
)

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Phone     string `json:"phone" validate:"required,phone"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// UpdateUserRequest represents a request to update user information
type UpdateUserRequest struct {
	Phone     *string `json:"phone,omitempty" validate:"omitempty,phone"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
}

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	DateOfBirth    *time.Time      `json:"date_of_birth,omitempty"`
	Phone          *string         `json:"phone,omitempty" validate:"omitempty,phone"`
	Address        *Address        `json:"address,omitempty"`
	EmploymentInfo *EmploymentInfo `json:"employment_info,omitempty"`
	FinancialInfo  *FinancialInfo  `json:"financial_info,omitempty"`
}

// DocumentUpload represents a document upload request
type DocumentUpload struct {
	Type     string `json:"type" validate:"required"`
	Content  []byte `json:"-"`
	MimeType string `json:"mime_type" validate:"required"`
	UploadIP string `json:"upload_ip"`
}

// KYCSession represents a KYC verification session
type KYCSession struct {
	ID                string                 `json:"id"`
	UserID            string                 `json:"user_id"`
	Provider          string                 `json:"provider"`
	SessionURL        string                 `json:"session_url"`
	ProviderReference string                 `json:"provider_reference"`
	ExpiresAt         time.Time              `json:"expires_at"`
	Status            string                 `json:"status"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
}

// DocumentStream represents a downloadable document
type DocumentStream struct {
	Content     []byte `json:"-"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
}

// GetFullName returns the user's full name
func (u *UserProfile) GetFullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// IsAdult checks if the user is at least 18 years old
func (u *UserProfile) IsAdult() bool {
	return time.Since(u.DateOfBirth).Hours() >= 18*365*24
}

// CalculateAge returns the user's age in years
func (u *UserProfile) CalculateAge() int {
	now := time.Now()
	years := now.Year() - u.DateOfBirth.Year()

	// Adjust if birthday hasn't occurred this year
	if now.YearDay() < u.DateOfBirth.YearDay() {
		years--
	}

	return years
}

// IsComplete checks if the profile has all required information
func (u *UserProfile) IsComplete() bool {
	return u.FirstName != "" &&
		u.LastName != "" &&
		!u.DateOfBirth.IsZero() &&
		u.SSNEncrypted != "" &&
		u.Phone != "" &&
		u.Address.Street != "" &&
		u.Address.City != "" &&
		u.Address.State != "" &&
		u.Address.ZipCode != ""
}

// GetMonthlyDebtToIncomeRatio calculates the debt-to-income ratio
func (f *FinancialInfo) GetMonthlyDebtToIncomeRatio() float64 {
	if f.MonthlyIncome == 0 {
		return 0
	}
	return f.MonthlyDebtPayments / f.MonthlyIncome
}

// GetTotalMonthlyIncome returns total monthly income including other sources
func (f *FinancialInfo) GetTotalMonthlyIncome() float64 {
	return f.MonthlyIncome + f.OtherIncome
}

// GetNetMonthlyIncome returns net monthly income after debt payments
func (f *FinancialInfo) GetNetMonthlyIncome() float64 {
	return f.GetTotalMonthlyIncome() - f.MonthlyDebtPayments
}
