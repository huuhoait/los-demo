package i18n

import (
	"context"
	"fmt"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

// Localizer handles internationalization for the loan service
type Localizer struct {
	bundle *i18n.Bundle
}

// NewLocalizer creates a new localizer instance
func NewLocalizer() (*Localizer, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// For now, we'll load the files from disk in development
	// In production, you would embed these files
	files := map[string][]byte{
		"en.toml": []byte(enTranslations),
		"vi.toml": []byte(viTranslations),
	}

	for filename, data := range files {
		if _, err := bundle.ParseMessageFileBytes(data, filename); err != nil {
			return nil, fmt.Errorf("failed to parse locale file %s: %w", filename, err)
		}
	}

	return &Localizer{
		bundle: bundle,
	}, nil
}

// GetLocalizer returns a localizer for the given languages
func (l *Localizer) GetLocalizer(langs ...string) *i18n.Localizer {
	return i18n.NewLocalizer(l.bundle, langs...)
}

// Localize localizes a message with the given context and template data
func (l *Localizer) Localize(ctx context.Context, messageID string, templateData map[string]interface{}) string {
	lang := GetLanguageFromContext(ctx)
	localizer := l.GetLocalizer(lang)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		// Fallback to message ID if localization fails
		return messageID
	}
	return msg
}

// LocalizeError localizes an error message
func (l *Localizer) LocalizeError(ctx context.Context, errorCode string, templateData map[string]interface{}) string {
	return l.Localize(ctx, errorCode, templateData)
}

// DetectLanguage detects language from Accept-Language header
func DetectLanguage(acceptLang string) string {
	if acceptLang == "" {
		return "en"
	}

	// Parse Accept-Language header
	langs := strings.Split(acceptLang, ",")
	for _, lang := range langs {
		// Clean up the language tag
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])

		// Check for Vietnamese
		if strings.HasPrefix(strings.ToLower(lang), "vi") {
			return "vi"
		}

		// Default to English for other languages
		if strings.HasPrefix(strings.ToLower(lang), "en") {
			return "en"
		}
	}

	return "en" // Default fallback
}

// Context keys
type contextKey string

const (
	LanguageContextKey contextKey = "language"
)

// SetLanguageInContext sets the language in context
func SetLanguageInContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, LanguageContextKey, lang)
}

// GetLanguageFromContext gets the language from context
func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(LanguageContextKey).(string); ok {
		return lang
	}
	return "en" // Default fallback
}

// LocalizedError represents an error with localization support
type LocalizedError struct {
	Code         string
	TemplateData map[string]interface{}
	HTTPStatus   int
}

func (e *LocalizedError) Error() string {
	return e.Code
}

// NewLocalizedError creates a new localized error
func NewLocalizedError(code string, httpStatus int, templateData map[string]interface{}) *LocalizedError {
	return &LocalizedError{
		Code:         code,
		TemplateData: templateData,
		HTTPStatus:   httpStatus,
	}
}

// FormatTemplateData formats template data for localization
func FormatTemplateData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}
	return data
}

// Embedded translation data
const enTranslations = `# English translations for Loan Service
# Error messages
[LOAN_001]
other = "Invalid loan amount"

[LOAN_002]
other = "Invalid loan purpose"

[LOAN_003]
other = "Invalid loan term"

[LOAN_004]
other = "Invalid income information"

[LOAN_005]
other = "Loan amount is below minimum requirement"

[LOAN_006]
other = "Loan amount exceeds maximum limit"

[LOAN_007]
other = "Insufficient income for requested loan amount"

[LOAN_008]
other = "Invalid state transition"

[LOAN_009]
other = "Loan offer has expired"

[LOAN_010]
other = "Loan application not found"

[LOAN_011]
other = "Failed to start workflow"

[LOAN_012]
other = "Workflow execution failed"

[LOAN_013]
other = "State conflict detected"

[LOAN_014]
other = "Conductor service unavailable"

[LOAN_015]
other = "Decision engine error"

[LOAN_016]
other = "State machine error"

[LOAN_017]
other = "Offer calculation error"

[LOAN_018]
other = "Application validation failed"

[LOAN_019]
other = "Invalid application status"

[LOAN_020]
other = "Invalid request format - please check your JSON data and field validation"

[LOAN_021]
other = "User not found"

[LOAN_022]
other = "Unauthorized access"

[LOAN_023]
other = "Database connection error"

[LOAN_024]
other = "External service error"

[LOAN_025]
other = "Document verification required"

[LOAN_026]
other = "Credit check failed"

[LOAN_027]
other = "KYC verification pending"

[LOAN_028]
other = "Manual review required"

[LOAN_029]
other = "Application already exists"

[LOAN_030]
other = "Invalid offer terms"

# User error messages
[USER_001]
other = "Invalid email format"

[USER_002]
other = "Invalid phone format"

[USER_003]
other = "Invalid SSN format"

[USER_004]
other = "Invalid date of birth"

[USER_005]
other = "Missing required field"

[USER_006]
other = "Email already exists"

[USER_007]
other = "Phone already exists"

[USER_008]
other = "SSN already exists"

[USER_009]
other = "User under minimum age"

[USER_010]
other = "KYC already completed"

[USER_011]
other = "Invalid document format"

[USER_012]
other = "File too large"

[USER_013]
other = "Upload failed"

[USER_014]
other = "Document not found"

[USER_015]
other = "Encryption failed"

[USER_016]
other = "S3 upload failed"

[USER_017]
other = "Unsupported document type"

[USER_018]
other = "Virus detected in file"

[USER_019]
other = "Document expired"

[USER_020]
other = "Document already exists"

[USER_021]
other = "KYC provider error"

[USER_022]
other = "KYC session expired"

[USER_023]
other = "KYC verification failed"

[USER_024]
other = "KYC manual review required"

[USER_025]
other = "KYC provider unavailable"

[USER_026]
other = "Database error"

[USER_027]
other = "Cache error"

[USER_028]
other = "Encryption error"

[USER_029]
other = "Notification error"

[USER_030]
other = "User not found"

[USER_031]
other = "Profile not found"

[USER_032]
other = "Unauthorized access"

[USER_033]
other = "Rate limit exceeded"

[USER_034]
other = "Service unavailable"

[USER_035]
other = "Data integrity error"

# Success messages
[APPLICATION_CREATED]
other = "Loan application created successfully"

[APPLICATION_UPDATED]
other = "Loan application updated successfully"

[APPLICATION_SUBMITTED]
other = "Loan application submitted successfully"

[PRE_QUALIFICATION_SUCCESS]
other = "Pre-qualification completed successfully"

[OFFER_GENERATED]
other = "Loan offer generated successfully"

[OFFER_ACCEPTED]
other = "Loan offer accepted successfully"

[WORKFLOW_STARTED]
other = "Loan processing workflow started"

[STATE_TRANSITION_SUCCESS]
other = "Application state updated successfully"`

const viTranslations = `# Vietnamese translations for Loan Service
# Error messages
[LOAN_001]
other = "Số tiền vay không hợp lệ"

[LOAN_002]
other = "Mục đích vay không hợp lệ"

[LOAN_003]
other = "Thời hạn vay không hợp lệ"

[LOAN_004]
other = "Thông tin thu nhập không hợp lệ"

[LOAN_005]
other = "Số tiền vay thấp hơn yêu cầu tối thiểu"

[LOAN_006]
other = "Số tiền vay vượt quá giới hạn tối đa"

[LOAN_007]
other = "Thu nhập không đủ cho số tiền vay yêu cầu"

[LOAN_008]
other = "Chuyển đổi trạng thái không hợp lệ"

[LOAN_009]
other = "Đề nghị vay đã hết hạn"

[LOAN_010]
other = "Không tìm thấy đơn xin vay"

[LOAN_011]
other = "Không thể khởi tạo quy trình"

[LOAN_012]
other = "Thực thi quy trình thất bại"

[LOAN_013]
other = "Phát hiện xung đột trạng thái"

[LOAN_014]
other = "Dịch vụ Conductor không khả dụng"

[LOAN_015]
other = "Lỗi hệ thống quyết định"

[LOAN_016]
other = "Lỗi máy trạng thái"

[LOAN_017]
other = "Lỗi tính toán đề nghị"

[LOAN_018]
other = "Xác thực đơn xin vay thất bại"

[LOAN_019]
other = "Trạng thái đơn xin vay không hợp lệ"

[LOAN_020]
other = "Định dạng yêu cầu không hợp lệ"

[LOAN_021]
other = "Không tìm thấy người dùng"

[LOAN_022]
other = "Truy cập không được phép"

[LOAN_023]
other = "Lỗi kết nối cơ sở dữ liệu"

[LOAN_024]
other = "Lỗi dịch vụ bên ngoài"

[LOAN_025]
other = "Yêu cầu xác minh tài liệu"

[LOAN_026]
other = "Kiểm tra tín dụng thất bại"

[LOAN_027]
other = "Xác minh KYC đang chờ xử lý"

[LOAN_028]
other = "Yêu cầu xem xét thủ công"

[LOAN_029]
other = "Đơn xin vay đã tồn tại"

[LOAN_030]
other = "Điều khoản đề nghị không hợp lệ"

# User error messages
[USER_001]
other = "Định dạng email không hợp lệ"

[USER_002]
other = "Định dạng số điện thoại không hợp lệ"

[USER_003]
other = "Định dạng SSN không hợp lệ"

[USER_004]
other = "Ngày sinh không hợp lệ"

[USER_005]
other = "Thiếu trường bắt buộc"

[USER_006]
other = "Email đã tồn tại"

[USER_007]
other = "Số điện thoại đã tồn tại"

[USER_008]
other = "SSN đã tồn tại"

[USER_009]
other = "Người dùng dưới độ tuổi tối thiểu"

[USER_010]
other = "KYC đã hoàn thành"

[USER_011]
other = "Định dạng tài liệu không hợp lệ"

[USER_012]
other = "Tệp quá lớn"

[USER_013]
other = "Tải lên thất bại"

[USER_014]
other = "Không tìm thấy tài liệu"

[USER_015]
other = "Mã hóa thất bại"

[USER_016]
other = "Tải lên S3 thất bại"

[USER_017]
other = "Loại tài liệu không được hỗ trợ"

[USER_018]
other = "Phát hiện vi-rút trong tệp"

[USER_019]
other = "Tài liệu đã hết hạn"

[USER_020]
other = "Tài liệu đã tồn tại"

[USER_021]
other = "Lỗi nhà cung cấp KYC"

[USER_022]
other = "Phiên KYC đã hết hạn"

[USER_023]
other = "Xác minh KYC thất bại"

[USER_024]
other = "Yêu cầu xem xét thủ công KYC"

[USER_025]
other = "Nhà cung cấp KYC không khả dụng"

[USER_026]
other = "Lỗi cơ sở dữ liệu"

[USER_027]
other = "Lỗi bộ nhớ đệm"

[USER_028]
other = "Lỗi mã hóa"

[USER_029]
other = "Lỗi thông báo"

[USER_030]
other = "Không tìm thấy người dùng"

[USER_031]
other = "Không tìm thấy hồ sơ"

[USER_032]
other = "Truy cập không được phép"

[USER_033]
other = "Vượt quá giới hạn tốc độ"

[USER_034]
other = "Dịch vụ không khả dụng"

[USER_035]
other = "Lỗi tính toàn vẹn dữ liệu"

# Success messages
[APPLICATION_CREATED]
other = "Đơn xin vay đã được tạo thành công"

[APPLICATION_UPDATED]
other = "Đơn xin vay đã được cập nhật thành công"

[APPLICATION_SUBMITTED]
other = "Đơn xin vay đã được nộp thành công"

[PRE_QUALIFICATION_SUCCESS]
other = "Thẩm định sơ bộ hoàn thành thành công"

[OFFER_GENERATED]
other = "Đề nghị vay đã được tạo thành công"

[OFFER_ACCEPTED]
other = "Đề nghị vay đã được chấp nhận thành công"

[WORKFLOW_STARTED]
other = "Quy trình xử lý vay đã được khởi tạo"

[STATE_TRANSITION_SUCCESS]
other = "Trạng thái đơn xin vay đã được cập nhật thành công"`
