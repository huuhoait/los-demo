# Internationalization (i18n) Implementation Guide

## Overview

The User Service now includes comprehensive internationalization support for **Vietnamese (vi)** and **English (en)** languages. This implementation provides localized error messages, success messages, validation messages, and UI elements.

## Features Implemented

### 1. Translation Files
- **English**: `/i18n/en.toml` - Complete English translations
- **Vietnamese**: `/i18n/vi.toml` - Complete Vietnamese translations

### 2. Translation Categories
- **Error Messages**: All 35 user service error codes (USER_001 to USER_035)
- **Success Messages**: User actions and notifications
- **Validation Messages**: Form validation and field requirements
- **UI Elements**: Form labels, buttons, status labels, document types
- **Notifications**: Email/SMS templates with placeholder support
- **Help Text**: User guidance and instructions
- **Time/Pagination**: Relative time and pagination labels

### 3. Technical Implementation
- **i18n Package**: `/pkg/i18n/localizer.go` - Core localization functionality
- **Middleware**: `/interfaces/middleware/i18n.go` - HTTP request language detection
- **Configuration**: Built into config system with language preferences
- **Error Handling**: Localized error responses with template data support

## Language Detection

The system detects user language preference in the following order:
1. `lang` query parameter (`?lang=vi`)
2. `X-Language` HTTP header
3. `Accept-Language` HTTP header
4. User profile preferences (if authenticated)
5. Default to English

## API Usage Examples

### Request with Language Parameter
```bash
curl -X GET "http://localhost:8082/api/v1/users/123?lang=vi" \
  -H "Content-Type: application/json"
```

### Request with Language Header
```bash
curl -X GET "http://localhost:8082/api/v1/users/123" \
  -H "Content-Type: application/json" \
  -H "X-Language: vi"
```

### Accept-Language Header
```bash
curl -X GET "http://localhost:8082/api/v1/users/123" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: vi-VN,vi;q=0.9,en;q=0.8"
```

## Response Examples

### English Error Response
```json
{
  "success": false,
  "error": {
    "code": "USER_001",
    "message": "User not found"
  },
  "language": "en",
  "request_id": "req-123456",
  "timestamp": "2024-08-14T10:30:00Z",
  "service": "user-service"
}
```

### Vietnamese Error Response
```json
{
  "success": false,
  "error": {
    "code": "USER_001",
    "message": "Không tìm thấy người dùng"
  },
  "language": "vi",
  "request_id": "req-123456",
  "timestamp": "2024-08-14T10:30:00Z",
  "service": "user-service"
}
```

### Success Response with Template Data
```json
{
  "success": true,
  "message": "Chào mừng Nguyễn Văn A! Tài khoản của bạn đã được tạo thành công.",
  "data": { "user_id": "user-123" },
  "language": "vi",
  "request_id": "req-123456",
  "timestamp": "2024-08-14T10:30:00Z",
  "service": "user-service"
}
```

### Validation Error Response
```json
{
  "success": false,
  "code": "VALIDATION_ERROR",
  "message": "Dữ liệu xác thực không hợp lệ",
  "errors": {
    "email": "Vui lòng nhập địa chỉ email hợp lệ",
    "phone": "Vui lòng nhập số điện thoại hợp lệ",
    "first_name": "Họ là bắt buộc"
  },
  "language": "vi",
  "request_id": "req-123456",
  "timestamp": "2024-08-14T10:30:00Z"
}
```

## Configuration

### Environment Variables
```bash
# i18n Configuration
I18N_DEFAULT_LANGUAGE=en
I18N_SUPPORTED_LANGUAGES=en,vi
I18N_FALLBACK_LANGUAGE=en
```

### YAML Configuration
```yaml
i18n:
  default_language: "en"
  supported_languages: ["en", "vi"]
  fallback_language: "en"
```

## Template Data Support

Many messages support template data for personalization:

### Email Messages
```json
{
  "FirstName": "Nguyễn Văn A",
  "VerificationCode": "123456",
  "ResetLink": "https://example.com/reset"
}
```

### Validation Messages
```json
{
  "MinLength": 8,
  "MaxLength": 100,
  "MinValue": 0,
  "MaxValue": 999999
}
```

## Adding New Languages

1. **Create Translation File**: Add `/i18n/[language_code].toml`
2. **Update Configuration**: Add language to supported languages list
3. **Update Detection**: Modify language detection logic if needed
4. **Test**: Verify all translations work correctly

## Translation Keys Structure

```
[errors]          # Error messages (USER_001 to USER_035)
[messages]        # Success and info messages
[validation]      # Form validation messages
[notifications]   # Email/SMS templates
[ui]             # Form labels, buttons, status
[help]           # Help text and instructions
[time]           # Time-related messages
[pagination]     # Pagination labels
```

## Best Practices

1. **Consistent Tone**: Maintain professional, helpful tone in both languages
2. **Cultural Sensitivity**: Consider cultural differences in messaging
3. **Template Variables**: Use `{VariableName}` format for consistency
4. **Fallback**: Always provide English fallback for missing translations
5. **Context**: Include context information in error messages
6. **Testing**: Test all languages with real data

## Supported Error Codes

All 35 user service error codes are fully localized:
- USER_001 to USER_010: User management errors
- USER_011 to USER_015: Profile management errors  
- USER_016 to USER_025: Document management errors
- USER_026 to USER_032: KYC verification errors
- USER_033 to USER_035: System errors

## Integration with Frontend

Frontend applications can:
1. Set user language preference via API headers
2. Receive localized error messages in responses
3. Use the same translation keys for consistency
4. Implement language switching with `lang` parameter

## Performance Considerations

- Translations are loaded once at startup (embedded in binary)
- Language detection is fast (header parsing)
- No database queries for translations
- Minimal memory overhead per request
- Caching of localizer instances per language

This implementation provides a solid foundation for multi-language support that can easily be extended to additional languages as needed.
