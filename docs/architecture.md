# Loan Origination System (LOS) - Technical Architecture Document

## Document Information

| Field | Value |
|-------|-------|
| Document Type | Technical Architecture |
| Project | Loan Origination System (LOS) |
| Version | 1.0 |
| Date | August 14, 2025 |
| Author | Architect Agent |
| Status | Draft - Ready for Review |

## Executive Summary

This document defines the technical architecture for a consumer Loan Origination System (LOS) supporting real-time loan applications, automated underwriting, and regulatory compliance. The system processes personal loans ($5K-$50K) with target processing times of <2 minutes for pre-qualification and <24 hours for funding.

**Key Architecture Decisions:**
- **Microservices Architecture** for scalability and service isolation
- **Event-Driven Design** for audit trails and real-time tracking
- **React/Node.js Stack** for rapid development and strong ecosystem
- **PostgreSQL + Redis** for data persistence and session management
- **API-First Design** for third-party integrations and future mobile apps

## High-Level System Architecture

### System Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │  Mobile Client  │    │  Admin Portal   │
│   (React SPA)   │    │   (Future)      │    │   (Internal)    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │      API Gateway          │
                    │   (Rate Limiting,         │
                    │    Authentication)        │
                    └─────────────┬─────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                        │                         │
┌───────┴────────┐    ┌──────────┴─────────┐    ┌─────────┴──────────┐
│  Core Services │    │  Integration Layer │    │  External Services │
│                │    │                    │    │                    │
│ • Auth Service │    │ • Credit Bureau    │    │ • Experian/Equifax │
│ • User Service │    │ • KYC/AML Service  │    │ • Jumio/Onfido     │
│ • Loan Service │    │ • E-Sign Service   │    │ • DocuSign         │
│ • Decision Eng │    │ • Payment Service  │    │ • Plaid/Dwolla     │
│ • Audit Service│    │ • Document Service │    │ • AWS S3           │
└────────────────┘    └────────────────────┘    └────────────────────┘
        │                        │                         
        └─────────────────────────┼─────────────────────────
                                  │
                    ┌─────────────┴─────────────┐
                    │     Data Layer            │
                    │                           │
                    │ ┌─────────┐ ┌───────────┐ │
                    │ │PostgreSQL│ │   Redis   │ │
                    │ │(Primary) │ │ (Sessions)│ │
                    │ └─────────┘ └───────────┘ │
                    └───────────────────────────┘
```

### Technology Stack

| Layer | Technology | Purpose | Justification |
|-------|------------|---------|---------------|
| **Frontend** | React 18 + Next.js 13 | Web application framework | Server-side rendering for SEO, excellent developer experience |
| **API Gateway** | Go-Gin + HTTP Signature | Request routing, middleware, and security | High performance, built-in security with HTTP signature authentication |
| **Backend Services** | Go 1.21 + Gin Framework | Microservices runtime | High performance, strong concurrency, excellent for financial systems |
| **Workflow Engine** | Netflix Conductor | Business process orchestration | Distributed workflow management for complex loan processes |
| **Database** | PostgreSQL 15 | Primary data store | ACID compliance for financial data, mature ecosystem |
| **Cache/Sessions** | Redis 7 | Session and cache store | High performance, pub/sub for real-time updates |
| **Message Queue** | Apache Kafka + Go clients | Event streaming and async processing | High throughput, durable messaging for financial events |
| **File Storage** | AWS S3 + Go SDK | Document storage with encryption | Compliance-ready, scalable, server-side encryption |
| **Configuration** | Viper | Configuration management | Flexible config with multiple sources, hot reloading |
| **Dependency Injection** | Wire | Compile-time dependency injection | Type-safe DI with zero runtime overhead |
| **Logging** | Zap | High-performance structured logging | Blazing fast, structured logging optimized for Go |
| **Internationalization** | go-i18n | Multi-language support | Comprehensive i18n with message templates |
| **Monitoring** | Prometheus + Grafana | System monitoring | Industry standard, excellent Go metrics support |
| **Log Aggregation** | ELK Stack | Log collection and analysis | Centralized logging with search and analytics |

## Detailed Service Architecture

### Error Code Management Framework

**Standardized Error Response Structure:**
```go
type ErrorResponse struct {
    Success   bool              `json:"success"`
    Error     *ErrorDetail      `json:"error,omitempty"`
    RequestID string            `json:"request_id"`
    Timestamp time.Time         `json:"timestamp"`
    Service   string            `json:"service"`
}

type ErrorDetail struct {
    Code        string            `json:"code"`
    Message     string            `json:"message"`
    Description string            `json:"description"`
    Field       string            `json:"field,omitempty"`
    Metadata    map[string]any    `json:"metadata,omitempty"`
    RetryAfter  *int              `json:"retry_after,omitempty"`
}
```

**Error Code Naming Convention:**
- Format: `{SERVICE}_{ERROR_CODE}`
- Service Names: `AUTH`, `USER`, `LOAN`, `DECISION`, `CREDIT`, `KYC`, `AUDIT`, `STORAGE`
- Examples: `AUTH_001`, `LOAN_015`, `USER_008`

**HTTP Status Code Mapping:**
- `400` - Validation errors, business rule violations
- `401` - Authentication failures
- `403` - Authorization failures
- `404` - Resource not found
- `409` - Conflict errors (duplicate resources)
- `422` - Business logic errors
- `429` - Rate limiting
- `500` - Internal system errors
- `502` - External service errors
- `503` - Service unavailable

### Clean Architecture Implementation

**Directory Structure per Service:**
```
service-name/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   ├── repositories/
│   │   └── services/
│   ├── infrastructure/
│   │   ├── database/
│   │   ├── external/
│   │   └── messaging/
│   ├── application/
│   │   ├── usecases/
│   │   └── dtos/
│   └── interfaces/
│       ├── http/
│       ├── grpc/
│       └── messaging/
├── pkg/
│   ├── errors/
│   ├── logger/
│   ├── config/
│   └── i18n/
├── configs/
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── locales/
│   ├── en/
│   │   └── messages.yaml
│   ├── es/
│   │   └── messages.yaml
│   └── fr/
│       └── messages.yaml
├── wire.go
├── wire_gen.go
└── migrations/
```

### Configuration Management with Viper

**Configuration Structure:**
```go
package config

import (
    "github.com/spf13/viper"
    "go.uber.org/zap"
)

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    AWS      AWSConfig      `mapstructure:"aws"`
    Kafka    KafkaConfig    `mapstructure:"kafka"`
    Auth     AuthConfig     `mapstructure:"auth"`
    Logging  LoggingConfig  `mapstructure:"logging"`
    I18n     I18nConfig     `mapstructure:"i18n"`
}

type ServerConfig struct {
    Port            int    `mapstructure:"port"`
    Host            string `mapstructure:"host"`
    ReadTimeout     int    `mapstructure:"read_timeout"`
    WriteTimeout    int    `mapstructure:"write_timeout"`
    ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

type I18nConfig struct {
    DefaultLanguage string   `mapstructure:"default_language"`
    SupportedLangs  []string `mapstructure:"supported_languages"`
    LocalesPath     string   `mapstructure:"locales_path"`
}

func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(configPath)
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")
    
    // Environment variable support
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}

// Hot reload support
func (c *Config) WatchConfig(logger *zap.Logger) {
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        logger.Info("Config file changed", zap.String("file", e.Name))
        // Reload configuration logic here
    })
}
```

### Dependency Injection with Wire

**Wire Provider Set:**
```go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "go.uber.org/zap"
    
    "our-los/internal/domain/repositories"
    "our-los/internal/domain/services"
    "our-los/internal/application/usecases"
    "our-los/internal/infrastructure/database"
    "our-los/internal/infrastructure/external"
    "our-los/internal/infrastructure/messaging"
    "our-los/internal/interfaces/http"
    "our-los/pkg/config"
    "our-los/pkg/logger"
    "our-los/pkg/i18n"
)

// ProviderSet is the Wire provider set for the application
var ProviderSet = wire.NewSet(
    // Infrastructure
    database.NewPostgresConnection,
    database.NewRedisConnection,
    messaging.NewKafkaProducer,
    external.NewS3Client,
    
    // Repositories
    repositories.NewUserRepository,
    repositories.NewLoanRepository,
    repositories.NewAuditRepository,
    
    // Domain Services
    services.NewAuthService,
    services.NewUserService,
    services.NewLoanService,
    services.NewDecisionService,
    
    // Use Cases
    usecases.NewAuthUseCase,
    usecases.NewUserUseCase,
    usecases.NewLoanUseCase,
    
    // HTTP Handlers
    http.NewAuthHandler,
    http.NewUserHandler,
    http.NewLoanHandler,
    
    // Utilities
    logger.NewZapLogger,
    i18n.NewLocalizer,
    config.LoadConfig,
)

// Application represents the entire application with all dependencies
type Application struct {
    Config      *config.Config
    Logger      *zap.Logger
    AuthHandler *http.AuthHandler
    UserHandler *http.UserHandler
    LoanHandler *http.LoanHandler
    I18n        *i18n.Localizer
}

// InitializeApplication creates a new application instance with all dependencies wired
func InitializeApplication(configPath string) (*Application, error) {
    wire.Build(
        ProviderSet,
        wire.Struct(new(Application), "*"),
    )
    return &Application{}, nil
}
```

### Zap Logging Implementation

**Logger Configuration:**
```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "our-los/pkg/config"
)

func NewZapLogger(cfg *config.Config) (*zap.Logger, error) {
    var zapConfig zap.Config
    
    switch cfg.Logging.Level {
    case "debug":
        zapConfig = zap.NewDevelopmentConfig()
    case "production":
        zapConfig = zap.NewProductionConfig()
    default:
        zapConfig = zap.NewProductionConfig()
    }
    
    // Customize output paths
    zapConfig.OutputPaths = cfg.Logging.OutputPaths
    zapConfig.ErrorOutputPaths = cfg.Logging.ErrorOutputPaths
    
    // Add custom fields
    zapConfig.InitialFields = map[string]interface{}{
        "service": cfg.Service.Name,
        "version": cfg.Service.Version,
        "env":     cfg.Environment,
    }
    
    // Configure encoding
    zapConfig.EncoderConfig.TimeKey = "timestamp"
    zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    zapConfig.EncoderConfig.LevelKey = "level"
    zapConfig.EncoderConfig.MessageKey = "message"
    zapConfig.EncoderConfig.CallerKey = "caller"
    zapConfig.EncoderConfig.StacktraceKey = "stacktrace"
    
    logger, err := zapConfig.Build(
        zap.AddCaller(),
        zap.AddStacktrace(zapcore.ErrorLevel),
    )
    if err != nil {
        return nil, err
    }
    
    return logger, nil
}

// ServiceLogger wraps zap.Logger with service-specific context
type ServiceLogger struct {
    logger *zap.Logger
    fields []zap.Field
}

func NewServiceLogger(logger *zap.Logger, service string) *ServiceLogger {
    return &ServiceLogger{
        logger: logger,
        fields: []zap.Field{
            zap.String("service", service),
        },
    }
}

func (sl *ServiceLogger) WithRequestID(requestID string) *ServiceLogger {
    return &ServiceLogger{
        logger: sl.logger,
        fields: append(sl.fields, zap.String("request_id", requestID)),
    }
}

func (sl *ServiceLogger) WithUserID(userID string) *ServiceLogger {
    return &ServiceLogger{
        logger: sl.logger,
        fields: append(sl.fields, zap.String("user_id", userID)),
    }
}

func (sl *ServiceLogger) Info(msg string, fields ...zap.Field) {
    sl.logger.Info(msg, append(sl.fields, fields...)...)
}

func (sl *ServiceLogger) Error(msg string, fields ...zap.Field) {
    sl.logger.Error(msg, append(sl.fields, fields...)...)
}

func (sl *ServiceLogger) Debug(msg string, fields ...zap.Field) {
    sl.logger.Debug(msg, append(sl.fields, fields...)...)
}

func (sl *ServiceLogger) Warn(msg string, fields ...zap.Field) {
    sl.logger.Warn(msg, append(sl.fields, fields...)...)
}
```

### Internationalization with go-i18n

**I18n Setup:**
```go
package i18n

import (
    "fmt"
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "gopkg.in/yaml.v2"
    "our-los/pkg/config"
)

type Localizer struct {
    bundle   *i18n.Bundle
    locales  map[string]*i18n.Localizer
    fallback *i18n.Localizer
}

func NewLocalizer(cfg *config.Config) (*Localizer, error) {
    bundle := i18n.NewBundle(language.English)
    bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
    
    // Load message files
    for _, lang := range cfg.I18n.SupportedLangs {
        _, err := bundle.LoadMessageFile(fmt.Sprintf("%s/%s/messages.yaml", cfg.I18n.LocalesPath, lang))
        if err != nil {
            return nil, fmt.Errorf("failed to load messages for %s: %w", lang, err)
        }
    }
    
    // Create localizers for each supported language
    locales := make(map[string]*i18n.Localizer)
    for _, lang := range cfg.I18n.SupportedLangs {
        locales[lang] = i18n.NewLocalizer(bundle, lang)
    }
    
    // Fallback to default language
    fallback := i18n.NewLocalizer(bundle, cfg.I18n.DefaultLanguage)
    
    return &Localizer{
        bundle:   bundle,
        locales:  locales,
        fallback: fallback,
    }, nil
}

func (l *Localizer) Localize(lang, messageID string, templateData map[string]interface{}) string {
    localizer, exists := l.locales[lang]
    if !exists {
        localizer = l.fallback
    }
    
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

// Error messages with localization
func (l *Localizer) GetErrorMessage(lang, errorCode string, metadata map[string]interface{}) string {
    messageID := fmt.Sprintf("error.%s", errorCode)
    return l.Localize(lang, messageID, metadata)
}
```

**Message Files Structure:**
```yaml
# locales/en/messages.yaml
error.AUTH_001: "Invalid credentials provided"
error.AUTH_002: "Authentication token is missing"
error.AUTH_003: "Authentication token has expired"
error.USER_001: "Invalid email address format"
error.USER_008: "Email address is already registered"
error.LOAN_001: "Loan amount must be between ${{.MinAmount}} and ${{.MaxAmount}}"
error.LOAN_007: "Loan amount of ${{.Amount}} is below minimum requirement"

validation.required: "This field is required"
validation.email: "Please enter a valid email address"
validation.phone: "Please enter a valid phone number"

success.application.created: "Your loan application has been submitted successfully"
success.user.created: "Account created successfully"
success.document.uploaded: "Document uploaded successfully"

# locales/es/messages.yaml
error.AUTH_001: "Credenciales inválidas proporcionadas"
error.AUTH_002: "Falta el token de autenticación"
error.AUTH_003: "El token de autenticación ha expirado"
error.USER_001: "Formato de dirección de correo electrónico inválido"
error.USER_008: "La dirección de correo electrónico ya está registrada"
error.LOAN_001: "El monto del préstamo debe estar entre ${{.MinAmount}} y ${{.MaxAmount}}"

validation.required: "Este campo es obligatorio"
validation.email: "Por favor ingrese una dirección de correo electrónico válida"
validation.phone: "Por favor ingrese un número de teléfono válido"

success.application.created: "Su solicitud de préstamo ha sido enviada exitosamente"
success.user.created: "Cuenta creada exitosamente"
success.document.uploaded: "Documento subido exitosamente"
```

### 1. Authentication & Authorization Service

**High-Level Purpose:** Secure user authentication, session management, and role-based access control.

#### Service Interface
```go
type AuthService interface {
    // User authentication
    Authenticate(ctx context.Context, credentials *LoginCredentials) (*AuthResult, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
    Logout(ctx context.Context, userID string) error
    
    // Session management
    ValidateSession(ctx context.Context, token string) (*SessionInfo, error)
    RevokeSession(ctx context.Context, sessionID string) error
    
    // Authorization
    Authorize(ctx context.Context, userID, resource, action string) (bool, error)
}
```

#### Error Codes
```go
const (
    // Authentication errors
    AUTH_001 = "AUTH_001" // Invalid credentials
    AUTH_002 = "AUTH_002" // Missing token
    AUTH_003 = "AUTH_003" // Expired token
    AUTH_004 = "AUTH_004" // Malformed token
    AUTH_005 = "AUTH_005" // Invalid signature
    AUTH_006 = "AUTH_006" // Account locked
    AUTH_007 = "AUTH_007" // Account suspended
    AUTH_008 = "AUTH_008" // Password expired
    AUTH_009 = "AUTH_009" // Too many attempts
    AUTH_010 = "AUTH_010" // MFA service unavailable
    AUTH_011 = "AUTH_011" // Identity provider error
    AUTH_012 = "AUTH_012" // Session store error
    AUTH_013 = "AUTH_013" // Token generation error
)
```

#### HTTP Signature Implementation
```go
type HTTPSignatureMiddleware struct {
    keyStore   KeyStore
    logger     *zap.Logger
    localizer  *i18n.Localizer
}

func (h *HTTPSignatureMiddleware) ValidateSignature() gin.HandlerFunc {
    return func(c *gin.Context) {
        signature := c.GetHeader("Signature")
        lang := c.GetHeader("Accept-Language")
        if lang == "" {
            lang = "en" // Default language
        }
        
        if signature == "" {
            errorMsg := h.localizer.GetErrorMessage(lang, "AUTH_002", nil)
            c.JSON(401, ErrorResponse{
                Success: false,
                Error: &ErrorDetail{
                    Code: AUTH_002,
                    Message: errorMsg,
                    Description: h.localizer.Localize(lang, "error.AUTH_002.description", nil),
                },
                RequestID: c.GetString("request_id"),
                Timestamp: time.Now(),
                Service: "auth-service",
            })
            c.Abort()
            return
        }
        
        // Validate signature logic here
        if !h.validateHTTPSignature(c, signature) {
            errorMsg := h.localizer.GetErrorMessage(lang, "AUTH_005", nil)
            c.JSON(401, ErrorResponse{
                Success: false,
                Error: &ErrorDetail{
                    Code: AUTH_005,
                    Message: errorMsg,
                    Description: h.localizer.Localize(lang, "error.AUTH_005.description", nil),
                },
                RequestID: c.GetString("request_id"),
                Timestamp: time.Now(),
                Service: "auth-service",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

#### Low-Level Implementation Details

**Database Schema:**
```sql
-- Users table
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  phone VARCHAR(20),
  email_verified BOOLEAN DEFAULT FALSE,
  phone_verified BOOLEAN DEFAULT FALSE,
  status VARCHAR(20) DEFAULT 'active',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Sessions table
CREATE TABLE user_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  refresh_token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  device_info JSONB,
  ip_address INET,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Roles and permissions
CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(50) UNIQUE NOT NULL,
  permissions JSONB NOT NULL
);

CREATE TABLE user_roles (
  user_id UUID REFERENCES users(id),
  role_id UUID REFERENCES roles(id),
  PRIMARY KEY (user_id, role_id)
);
```

**Key Components:**
- **JWT Token Manager:** Issues short-lived access tokens (15 min) and long-lived refresh tokens (30 days)
- **Password Security:** bcrypt with salt rounds of 12, password policy enforcement
- **Session Store:** Redis-based session storage with automatic expiration
- **Rate Limiting:** Gin middleware for login attempts (5 attempts per 15 minutes)
- **HTTP Signature Validation:** Request signing for API security
- **Configuration Management:** Viper for flexible config with hot reloading
- **Dependency Injection:** Wire for compile-time DI with zero runtime overhead
- **Structured Logging:** Zap for high-performance logging with request correlation
- **Internationalization:** go-i18n for multi-language error messages
- **Clean Architecture Layers:**
  - Domain: User entities, authentication rules
  - Application: Authentication use cases
  - Infrastructure: JWT handling, Redis sessions
  - Interface: Gin HTTP handlers

**Security Features:**
- Refresh token rotation on each use
- Device fingerprinting for fraud detection
- Automatic session revocation on suspicious activity
- OWASP compliance for authentication flows

### 2. User Management Service

**High-Level Purpose:** Customer profile management, KYC data storage, and user lifecycle operations.

#### Service Interface
```go
type UserService interface {
    // Profile management
    CreateUser(ctx context.Context, userData *CreateUserRequest) (*User, error)
    GetUser(ctx context.Context, userID string) (*User, error)
    UpdateUser(ctx context.Context, userID string, updates *UpdateUserRequest) (*User, error)
    
    // KYC operations
    InitiateKYC(ctx context.Context, userID string) (*KYCSession, error)
    UpdateKYCStatus(ctx context.Context, userID string, status KYCStatus, data *KYCData) error
    GetKYCStatus(ctx context.Context, userID string) (*KYCStatus, error)
    
    // Document management with S3
    UploadDocument(ctx context.Context, userID string, document *DocumentUpload) (*Document, error)
    GetDocuments(ctx context.Context, userID string) ([]*Document, error)
    DownloadDocument(ctx context.Context, userID, documentID string) (*DocumentStream, error)
}
```

#### Error Codes
```go
const (
    // User management errors
    USER_001 = "USER_001" // Invalid email
    USER_002 = "USER_002" // Invalid phone
    USER_003 = "USER_003" // Invalid SSN
    USER_004 = "USER_004" // Invalid date of birth
    USER_005 = "USER_005" // Missing required field
    USER_006 = "USER_006" // Email already exists
    USER_007 = "USER_007" // Phone already exists
    USER_008 = "USER_008" // SSN already exists
    USER_009 = "USER_009" // Under age
    USER_010 = "USER_010" // KYC already completed
    USER_011 = "USER_011" // Invalid document format
    USER_012 = "USER_012" // File too large
    USER_013 = "USER_013" // Upload failed
    USER_014 = "USER_014" // Document not found
    USER_015 = "USER_015" // Encryption failed
    USER_016 = "USER_016" // S3 upload failed
    USER_017 = "USER_017" // KYC provider error
    USER_018 = "USER_018" // Database error
    USER_019 = "USER_019" // System encryption error
)
```

#### S3 Document Management Implementation
```go
type S3DocumentManager struct {
    s3Client  *s3.Client
    bucket    string
    kmsKeyID  string
    logger    *zap.Logger
    localizer *i18n.Localizer
}

func (s *S3DocumentManager) UploadDocument(ctx context.Context, userID string, doc *DocumentUpload) (*Document, error) {
    logger := s.logger.With(
        zap.String("user_id", userID),
        zap.String("document_type", doc.Type),
        zap.String("operation", "upload_document"),
    )
    
    // Validate file type and size
    if err := s.validateDocument(doc); err != nil {
        logger.Error("Document validation failed", zap.Error(err))
        return nil, &errors.ServiceError{
            Code: USER_011,
            Message: s.localizer.GetErrorMessage("en", "USER_011", nil),
            Description: err.Error(),
        }
    }
    
    // Generate secure file path
    objectKey := fmt.Sprintf("users/%s/documents/%s/%s", 
        userID, doc.Type, uuid.New().String())
    
    logger.Info("Starting S3 upload", zap.String("object_key", objectKey))
    
    // Upload with server-side encryption
    _, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:               aws.String(s.bucket),
        Key:                  aws.String(objectKey),
        Body:                 bytes.NewReader(doc.Content),
        ContentType:          aws.String(doc.MimeType),
        ServerSideEncryption: types.ServerSideEncryptionAwsKms,
        SSEKMSKeyId:         aws.String(s.kmsKeyID),
        Metadata: map[string]string{
            "user-id":      userID,
            "document-type": doc.Type,
            "upload-ip":    doc.UploadIP,
        },
    })
    
    if err != nil {
        logger.Error("Failed to upload document to S3", 
            zap.Error(err),
            zap.String("bucket", s.bucket),
            zap.String("object_key", objectKey),
        )
        return nil, &errors.ServiceError{
            Code: USER_016,
            Message: s.localizer.GetErrorMessage("en", "USER_016", nil),
            Description: "Unable to store document securely",
        }
    }
    
    logger.Info("Document uploaded successfully", 
        zap.String("object_key", objectKey),
        zap.Int64("file_size", int64(len(doc.Content))),
    )
    
    return &Document{
        ID:       uuid.New().String(),
        UserID:   userID,
        Type:     doc.Type,
        FilePath: objectKey,
        FileSize: int64(len(doc.Content)),
        MimeType: doc.MimeType,
    }, nil
}
```

#### Low-Level Implementation Details

**Database Schema:**
```sql
-- User profiles
CREATE TABLE user_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) UNIQUE,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  date_of_birth DATE NOT NULL,
  ssn_encrypted TEXT NOT NULL, -- AES-256 encrypted
  phone VARCHAR(20),
  address JSONB NOT NULL,
  employment_info JSONB,
  financial_info JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- KYC verification data
CREATE TABLE kyc_verifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  verification_type VARCHAR(50) NOT NULL, -- 'identity', 'address', 'income'
  provider VARCHAR(50) NOT NULL, -- 'jumio', 'onfido', 'manual'
  status VARCHAR(20) NOT NULL, -- 'pending', 'verified', 'failed', 'manual_review'
  provider_reference VARCHAR(255),
  verification_data JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Document storage metadata
CREATE TABLE user_documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  document_type VARCHAR(50) NOT NULL, -- 'drivers_license', 'passport', 'pay_stub'
  file_path VARCHAR(500) NOT NULL, -- S3 path
  file_size INTEGER NOT NULL,
  mime_type VARCHAR(100) NOT NULL,
  encryption_key VARCHAR(255) NOT NULL, -- For client-side encryption
  upload_ip INET,
  created_at TIMESTAMP DEFAULT NOW()
);
```

**Key Components:**
- **Profile Validator:** Real-time validation using JSON Schema
- **Document Processor:** Image optimization, format conversion, virus scanning
- **Encryption Manager:** Client-side encryption for sensitive documents
- **KYC Orchestrator:** Manages multi-step verification workflows

**Data Protection:**
- Field-level encryption for PII data
- Audit logging for all profile changes
- GDPR-compliant data deletion capabilities
- Data retention policy enforcement (7 years for loan data)

### 3. Loan Management Service

**High-Level Purpose:** Loan application lifecycle, state management, and business logic orchestration.

#### Service Interface
```go
type LoanService interface {
    // Application lifecycle
    CreateApplication(ctx context.Context, userID string, loanRequest *LoanRequest) (*Application, error)
    GetApplication(ctx context.Context, applicationID string) (*Application, error)
    UpdateApplicationStatus(ctx context.Context, applicationID string, status ApplicationStatus) error
    
    // Pre-qualification
    PreQualify(ctx context.Context, userID string, basicInfo *PreQualRequest) (*PreQualResult, error)
    
    // Loan terms and offers
    GenerateOffer(ctx context.Context, applicationID string) (*LoanOffer, error)
    AcceptOffer(ctx context.Context, applicationID string, acceptance *OfferAcceptance) error
    
    // State machine operations with Conductor workflow
    TransitionState(ctx context.Context, applicationID string, newState ApplicationState) error
    GetStateHistory(ctx context.Context, applicationID string) ([]*StateTransition, error)
    StartWorkflow(ctx context.Context, applicationID string, workflowDef string) (*WorkflowExecution, error)
}
```

#### Error Codes
```go
const (
    // Loan management errors
    LOAN_001 = "LOAN_001" // Invalid amount
    LOAN_002 = "LOAN_002" // Invalid purpose
    LOAN_003 = "LOAN_003" // Invalid term
    LOAN_004 = "LOAN_004" // Invalid income
    LOAN_005 = "LOAN_005" // Amount too low
    LOAN_006 = "LOAN_006" // Amount too high
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
)
```

#### Netflix Conductor Workflow Integration
```go
type ConductorWorkflowManager struct {
    client    *conductor.Client
    logger    *zap.Logger
    localizer *i18n.Localizer
}

func (c *ConductorWorkflowManager) StartLoanProcessingWorkflow(ctx context.Context, appID string) error {
    logger := c.logger.With(
        zap.String("application_id", appID),
        zap.String("operation", "start_workflow"),
    )
    
    workflowDef := &conductor.WorkflowDefinition{
        Name:    "loan_processing_workflow",
        Version: 1,
        Tasks: []conductor.Task{
            {
                Name:              "validate_application",
                TaskReferenceName: "validate_app_ref",
                Type:              "SIMPLE",
                InputParameters: map[string]interface{}{
                    "application_id": appID,
                },
            },
            {
                Name:              "credit_check",
                TaskReferenceName: "credit_check_ref",
                Type:              "SIMPLE",
                InputParameters: map[string]interface{}{
                    "application_id": appID,
                },
            },
            {
                Name:              "underwriting_decision",
                TaskReferenceName: "underwriting_ref",
                Type:              "DECISION",
                CaseValueParam:    "credit_score",
                DecisionCases: map[string][]conductor.Task{
                    "high_score": {{
                        Name: "auto_approve",
                        Type: "SIMPLE",
                    }},
                    "low_score": {{
                        Name: "manual_review",
                        Type: "HUMAN",
                    }},
                },
            },
        },
    }
    
    logger.Info("Starting Conductor workflow", 
        zap.String("workflow_name", workflowDef.Name),
        zap.Int("version", workflowDef.Version),
    )
    
    execution, err := c.client.StartWorkflow(ctx, workflowDef, map[string]interface{}{
        "application_id": appID,
        "start_time":     time.Now(),
    })
    
    if err != nil {
        logger.Error("Failed to start Conductor workflow", 
            zap.Error(err),
            zap.String("workflow_name", workflowDef.Name),
        )
        return &errors.ServiceError{
            Code: LOAN_011,
            Message: c.localizer.GetErrorMessage("en", "LOAN_011", nil),
            Description: "Unable to start loan processing workflow",
        }
    }
    
    logger.Info("Started loan processing workflow successfully", 
        zap.String("workflow_id", execution.WorkflowID),
        zap.String("correlation_id", execution.CorrelationID),
    )
    return nil
}
```

#### Low-Level Implementation Details

**Database Schema:**
```sql
-- Loan applications
CREATE TABLE loan_applications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  application_number VARCHAR(20) UNIQUE NOT NULL, -- Human-readable ID
  loan_amount DECIMAL(10,2) NOT NULL,
  loan_purpose VARCHAR(100) NOT NULL,
  requested_term_months INTEGER NOT NULL,
  employment_income DECIMAL(10,2),
  other_income DECIMAL(10,2),
  monthly_debt_payments DECIMAL(10,2),
  current_state VARCHAR(50) NOT NULL DEFAULT 'initiated',
  risk_score INTEGER,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- State machine transitions
CREATE TABLE application_state_transitions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  application_id UUID REFERENCES loan_applications(id),
  from_state VARCHAR(50),
  to_state VARCHAR(50) NOT NULL,
  transition_reason VARCHAR(255),
  automated BOOLEAN DEFAULT TRUE,
  user_id UUID, -- Who triggered the transition
  metadata JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Loan offers
CREATE TABLE loan_offers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  application_id UUID REFERENCES loan_applications(id),
  offer_amount DECIMAL(10,2) NOT NULL,
  interest_rate DECIMAL(5,4) NOT NULL,
  term_months INTEGER NOT NULL,
  monthly_payment DECIMAL(10,2) NOT NULL,
  total_interest DECIMAL(10,2) NOT NULL,
  apr DECIMAL(5,4) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'accepted', 'declined', 'expired'
  created_at TIMESTAMP DEFAULT NOW()
);
```

**State Machine Implementation:**
```go
type ApplicationState string

const (
    INITIATED           ApplicationState = "initiated"
    PRE_QUALIFIED      ApplicationState = "pre_qualified"
    DOCUMENTS_SUBMITTED ApplicationState = "documents_submitted"
    IDENTITY_VERIFIED   ApplicationState = "identity_verified"
    UNDERWRITING       ApplicationState = "underwriting"
    MANUAL_REVIEW      ApplicationState = "manual_review"
    APPROVED           ApplicationState = "approved"
    DENIED             ApplicationState = "denied"
    DOCUMENTS_SIGNED   ApplicationState = "documents_signed"
    FUNDED             ApplicationState = "funded"
    ACTIVE             ApplicationState = "active"
)

type StateMachine struct {
    transitions map[ApplicationState][]ApplicationState
    conductor   *ConductorWorkflowManager
}

func NewStateMachine(conductor *ConductorWorkflowManager) *StateMachine {
    return &StateMachine{
        transitions: map[ApplicationState][]ApplicationState{
            INITIATED:           {PRE_QUALIFIED},
            PRE_QUALIFIED:      {DOCUMENTS_SUBMITTED},
            DOCUMENTS_SUBMITTED: {IDENTITY_VERIFIED},
            IDENTITY_VERIFIED:   {UNDERWRITING},
            UNDERWRITING:       {APPROVED, DENIED, MANUAL_REVIEW},
            MANUAL_REVIEW:      {APPROVED, DENIED},
            APPROVED:           {DOCUMENTS_SIGNED},
            DOCUMENTS_SIGNED:   {FUNDED},
            FUNDED:             {ACTIVE},
        },
        conductor: conductor,
    }
}

func (sm *StateMachine) TransitionTo(ctx context.Context, appID string, currentState, newState ApplicationState) error {
    validTransitions, exists := sm.transitions[currentState]
    if !exists {
        return &errors.ServiceError{
            Code: LOAN_008,
            Message: "Invalid current state",
            Description: fmt.Sprintf("State %s is not recognized", currentState),
        }
    }
    
    isValidTransition := false
    for _, validState := range validTransitions {
        if validState == newState {
            isValidTransition = true
            break
        }
    }
    
    if !isValidTransition {
        return &errors.ServiceError{
            Code: LOAN_008,
            Message: "Invalid state transition",
            Description: fmt.Sprintf("Cannot transition from %s to %s", currentState, newState),
        }
    }
    
    // Trigger workflow tasks based on state transition
    if err := sm.conductor.HandleStateTransition(ctx, appID, currentState, newState); err != nil {
        return err
    }
    
    return nil
}
```

**Key Components:**
- **State Machine Engine:** Enforces valid state transitions with audit trails
- **Business Rules Engine:** Configurable loan policies and eligibility rules
- **Offer Calculator:** Interest rate and payment calculation engine
- **Application Validator:** Multi-step validation with progressive enhancement

### 4. Decision Engine Service

**High-Level Purpose:** Automated underwriting, manual review queue management, and configurable approval workflows.

#### Service Interface
```go
type DecisionService interface {
    // Automated decisions
    EvaluateApplication(ctx context.Context, applicationID string) (*DecisionResult, error)
    RecalculateDecision(ctx context.Context, applicationID string, overrides *RuleOverrides) (*DecisionResult, error)
    
    // Manual review queue with Conductor task management
    GetReviewQueue(ctx context.Context, reviewerLevel ReviewerLevel) ([]*ReviewQueueItem, error)
    AssignReviewer(ctx context.Context, applicationID, reviewerID string) error
    SubmitManualDecision(ctx context.Context, applicationID string, decision *ManualDecision) error
    
    // Configuration management
    UpdateRules(ctx context.Context, ruleSet *DecisionRules) error
    GetRules(ctx context.Context) (*DecisionRules, error)
    UpdateSLAConfig(ctx context.Context, config *SLAConfiguration) error
}
```

#### Error Codes
```go
const (
    // Decision engine errors
    DECISION_001 = "DECISION_001" // Invalid application
    DECISION_002 = "DECISION_002" // Invalid rules
    DECISION_003 = "DECISION_003" // Invalid override
    DECISION_004 = "DECISION_004" // Insufficient data
    DECISION_005 = "DECISION_005" // Conflicting rules
    DECISION_006 = "DECISION_006" // Manual review required
    DECISION_007 = "DECISION_007" // SLA violated
    DECISION_008 = "DECISION_008" // Reviewer not qualified
    DECISION_009 = "DECISION_009" // Credit bureau error
    DECISION_010 = "DECISION_010" // Fraud service error
    DECISION_011 = "DECISION_011" // Conductor task failed
    DECISION_012 = "DECISION_012" // Rules engine error
    DECISION_013 = "DECISION_013" // Queue management error
)
```

#### Conductor Integration for Manual Review
```go
type ManualReviewOrchestrator struct {
    conductor *conductor.Client
    logger    *zap.Logger
    localizer *i18n.Localizer
}

func (m *ManualReviewOrchestrator) CreateReviewTask(ctx context.Context, appID string, reviewLevel ReviewerLevel) error {
    logger := m.logger.With(
        zap.String("application_id", appID),
        zap.String("review_level", string(reviewLevel)),
        zap.String("operation", "create_review_task"),
    )
    
    taskDef := &conductor.TaskDefinition{
        Name:        "manual_review_task",
        Description: "Human review required for loan application",
        InputKeys:   []string{"application_id", "review_level", "sla_deadline"},
        OutputKeys:  []string{"decision", "reviewer_id", "review_notes"},
    }
    
    task := &conductor.Task{
        TaskID:           uuid.New().String(),
        TaskType:         "HUMAN",
        TaskDefName:      "manual_review_task",
        WorkflowTaskType: conductor.TaskTypeHuman,
        InputData: map[string]interface{}{
            "application_id": appID,
            "review_level":   string(reviewLevel),
            "sla_deadline":   time.Now().Add(24 * time.Hour),
        },
    }
    
    logger.Info("Creating manual review task", 
        zap.String("task_id", task.TaskID),
        zap.String("task_type", string(task.TaskType)),
    )
    
    if err := m.conductor.PollAndExecute(ctx, task); err != nil {
        logger.Error("Failed to create manual review task", 
            zap.Error(err),
            zap.String("task_id", task.TaskID),
        )
        return &errors.ServiceError{
            Code: DECISION_011,
            Message: m.localizer.GetErrorMessage("en", "DECISION_011", nil),
            Description: "Unable to create manual review workflow task",
        }
    }
    
    logger.Info("Manual review task created successfully", 
        zap.String("task_id", task.TaskID),
    )
    return nil
}
```

#### Low-Level Implementation Details

**Database Schema:**
```sql
-- Decision rules configuration
CREATE TABLE decision_rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_name VARCHAR(100) NOT NULL,
  rule_type VARCHAR(50) NOT NULL, -- 'credit_score', 'dti_ratio', 'income_verification'
  conditions JSONB NOT NULL, -- Configurable rule conditions
  action VARCHAR(50) NOT NULL, -- 'approve', 'deny', 'manual_review'
  priority INTEGER NOT NULL,
  active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Decision results
CREATE TABLE decision_results (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  application_id UUID REFERENCES loan_applications(id),
  decision_type VARCHAR(20) NOT NULL, -- 'automated', 'manual'
  decision VARCHAR(20) NOT NULL, -- 'approved', 'denied', 'manual_review'
  confidence_score DECIMAL(3,2), -- 0.00 to 1.00
  decision_factors JSONB NOT NULL, -- Factors that influenced decision
  rules_applied JSONB NOT NULL, -- Which rules were evaluated
  processing_time_ms INTEGER,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Manual review queue
CREATE TABLE manual_review_queue (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  application_id UUID REFERENCES loan_applications(id),
  review_level VARCHAR(20) NOT NULL, -- 'junior', 'senior', 'manager'
  assigned_reviewer_id UUID,
  review_reason VARCHAR(255) NOT NULL,
  priority INTEGER DEFAULT 5, -- 1 = highest, 10 = lowest
  sla_expires_at TIMESTAMP NOT NULL,
  status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'in_progress', 'completed'
  created_at TIMESTAMP DEFAULT NOW(),
  assigned_at TIMESTAMP,
  completed_at TIMESTAMP
);

-- Manual review decisions
CREATE TABLE manual_review_decisions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  review_queue_id UUID REFERENCES manual_review_queue(id),
  reviewer_id UUID NOT NULL,
  decision VARCHAR(20) NOT NULL, -- 'approved', 'denied', 'escalate'
  decision_reason TEXT NOT NULL,
  conditions JSONB, -- Additional conditions or modified terms
  review_notes TEXT,
  time_spent_minutes INTEGER,
  created_at TIMESTAMP DEFAULT NOW()
);
```

**Decision Engine Implementation:**
```go
type DecisionRule struct {
    ID         string                 `json:"id"`
    Name       string                 `json:"name"`
    Type       string                 `json:"type"` // 'credit_score', 'dti_ratio', 'income_verification', 'employment_stability'
    Conditions []RuleCondition        `json:"conditions"`
    Action     string                 `json:"action"` // 'approve', 'deny', 'manual_review'
    Priority   int                    `json:"priority"`
}

type RuleCondition struct {
    Operator string      `json:"operator"` // 'gt', 'lt', 'eq', 'gte', 'lte', 'in', 'not_in'
    Value    interface{} `json:"value"`
    Field    string      `json:"field"`
}

type RulesEngine struct {
    rules   []*DecisionRule
    logger  *logrus.Logger
}

func (re *RulesEngine) EvaluateApplication(ctx context.Context, appData *ApplicationData) (*DecisionResult, error) {
    result := &DecisionResult{
        ApplicationID:   appData.ID,
        DecisionType:    "automated",
        ProcessingStart: time.Now(),
        FactorsApplied:  make([]string, 0),
        RulesApplied:    make([]string, 0),
    }
    
    // Sort rules by priority
    sort.Slice(re.rules, func(i, j int) bool {
        return re.rules[i].Priority < re.rules[j].Priority
    })
    
    for _, rule := range re.rules {
        if re.evaluateRule(rule, appData) {
            result.RulesApplied = append(result.RulesApplied, rule.ID)
            result.FactorsApplied = append(result.FactorsApplied, rule.Name)
            
            switch rule.Action {
            case "approve":
                result.Decision = "approved"
                result.ConfidenceScore = re.calculateConfidence(appData)
                return result, nil
            case "deny":
                result.Decision = "denied"
                result.ConfidenceScore = re.calculateConfidence(appData)
                return result, nil
            case "manual_review":
                result.Decision = "manual_review"
                result.ReviewReason = rule.Name
                return result, nil
            }
        }
    }
    
    // Default to manual review if no rules matched
    result.Decision = "manual_review"
    result.ReviewReason = "No matching automated rules"
    
    return result, nil
}

func (re *RulesEngine) evaluateRule(rule *DecisionRule, appData *ApplicationData) bool {
    for _, condition := range rule.Conditions {
        if !re.evaluateCondition(condition, appData) {
            return false
        }
    }
    return true
}

func (re *RulesEngine) evaluateCondition(condition RuleCondition, appData *ApplicationData) bool {
    fieldValue := re.getFieldValue(condition.Field, appData)
    if fieldValue == nil {
        return false
    }
    
    switch condition.Operator {
    case "gt":
        return re.compareNumeric(fieldValue, condition.Value, ">")
    case "gte":
        return re.compareNumeric(fieldValue, condition.Value, ">=")
    case "lt":
        return re.compareNumeric(fieldValue, condition.Value, "<")
    case "lte":
        return re.compareNumeric(fieldValue, condition.Value, "<=")
    case "eq":
        return reflect.DeepEqual(fieldValue, condition.Value)
    case "in":
        return re.inSlice(fieldValue, condition.Value)
    case "not_in":
        return !re.inSlice(fieldValue, condition.Value)
    default:
        re.logger.WithField("operator", condition.Operator).Warn("Unknown operator")
        return false
    }
}
```

**Key Components:**
- **Rules Engine:** Configurable business rules with real-time updates
- **SLA Manager:** Automatic escalation based on configurable time limits
- **Queue Balancer:** Load balancing for manual reviewers
- **Decision Audit:** Complete audit trail of all decision factors

**Manual Review Workflow:**
1. **Automatic Assignment:** Based on reviewer level and current workload
2. **SLA Tracking:** Real-time monitoring with escalation alerts
3. **Decision Templates:** Pre-configured approval/denial reasons
4. **Override Authority:** Manager-level overrides with justification requirements

### 5. Integration Services Layer

**High-Level Purpose:** Abstraction layer for external service integrations with circuit breakers, retry logic, and failover handling.

#### Credit Bureau Integration Service

```go
type CreditBureauService interface {
    // Credit pulls
    SoftCreditPull(ctx context.Context, userID string, bureau CreditBureau) (*CreditReport, error)
    HardCreditPull(ctx context.Context, userID string, bureau CreditBureau) (*CreditReport, error)
    
    // Identity verification
    VerifyIdentity(ctx context.Context, personalInfo *PersonalInfo) (*IdentityResult, error)
    
    // Fraud detection
    CheckFraud(ctx context.Context, application *ApplicationData) (*FraudScore, error)
}

// Error codes for Credit Bureau service
const (
    CREDIT_001 = "CREDIT_001" // Invalid SSN
    CREDIT_002 = "CREDIT_002" // Invalid DOB
    CREDIT_003 = "CREDIT_003" // Thin file
    CREDIT_004 = "CREDIT_004" // Frozen report
    CREDIT_005 = "CREDIT_005" // Bureau unavailable
    CREDIT_006 = "CREDIT_006" // Rate limited
    CREDIT_007 = "CREDIT_007" // Circuit breaker open
)

type CreditBureauClient struct {
    circuitBreaker *CircuitBreaker
    rateLimiter    *RateLimiter
    httpClient     *http.Client
    logger         *logrus.Logger
}

func (c *CreditBureauClient) SoftCreditPull(ctx context.Context, userID string, bureau CreditBureau) (*CreditReport, error) {
    if !c.circuitBreaker.AllowRequest() {
        return nil, &errors.ServiceError{
            Code: CREDIT_007,
            Message: "Credit bureau service unavailable",
            Description: "Circuit breaker is open due to service failures",
            RetryAfter: &[]int{300}[0], // 5 minutes
        }
    }
    
    if !c.rateLimiter.Allow() {
        return nil, &errors.ServiceError{
            Code: CREDIT_006,
            Message: "Rate limit exceeded",
            Description: "Too many credit bureau requests",
            RetryAfter: &[]int{60}[0], // 1 minute
        }
    }
    
    // Implementation with circuit breaker and retry logic
    return c.performCreditCheck(ctx, userID, bureau, "soft")
}
```

#### KYC/AML Integration Service

```go
type KYCService interface {
    // Identity verification
    VerifyIdentity(ctx context.Context, documents *IdentityDocuments) (*VerificationResult, error)
    
    // Address verification
    VerifyAddress(ctx context.Context, addressInfo *AddressInfo) (*AddressVerification, error)
    
    // Sanctions screening
    ScreenSanctions(ctx context.Context, personalInfo *PersonalInfo) (*SanctionsResult, error)
    
    // PEP checking
    CheckPEP(ctx context.Context, personalInfo *PersonalInfo) (*PEPResult, error)
}

// Error codes for KYC service
const (
    KYC_001 = "KYC_001" // Invalid document
    KYC_002 = "KYC_002" // Document expired
    KYC_003 = "KYC_003" // Identity mismatch
    KYC_004 = "KYC_004" // Sanctions hit
    KYC_005 = "KYC_005" // PEP identified
    KYC_006 = "KYC_006" // Provider error
    KYC_007 = "KYC_007" // Document processing failed
)
```

### 6. Audit & Compliance Service

**High-Level Purpose:** Comprehensive audit trails, regulatory reporting, and compliance monitoring.

#### Service Interface
```go
type AuditService interface {
    // Event logging
    LogEvent(ctx context.Context, event *AuditEvent) error
    LogUserAction(ctx context.Context, userID string, action *UserAction, metadata map[string]interface{}) error
    LogSystemEvent(ctx context.Context, event *SystemEvent, metadata map[string]interface{}) error
    
    // Compliance reporting
    GenerateReport(ctx context.Context, reportType ReportType, dateRange *DateRange) (*ComplianceReport, error)
    ScheduleReport(ctx context.Context, reportConfig *ReportConfiguration) error
    
    // Audit queries
    GetAuditTrail(ctx context.Context, entityID, entityType string) ([]*AuditEvent, error)
    SearchEvents(ctx context.Context, criteria *SearchCriteria) ([]*AuditEvent, error)
}
```

#### Error Codes
```go
const (
    // Audit service errors
    AUDIT_001 = "AUDIT_001" // Invalid event type
    AUDIT_002 = "AUDIT_002" // Missing entity ID
    AUDIT_003 = "AUDIT_003" // Invalid date range
    AUDIT_004 = "AUDIT_004" // Report not found
    AUDIT_005 = "AUDIT_005" // Retention policy violation
    AUDIT_006 = "AUDIT_006" // Duplicate event
    AUDIT_007 = "AUDIT_007" // Kafka publish failed
    AUDIT_008 = "AUDIT_008" // S3 storage failed
    AUDIT_009 = "AUDIT_009" // Event processing error
    AUDIT_010 = "AUDIT_010" // Report generation error
)
```

#### Kafka Event Streaming Implementation
```go
type AuditEventPublisher struct {
    producer *kafka.Producer
    topic    string
    logger   *logrus.Logger
}

func (a *AuditEventPublisher) PublishEvent(ctx context.Context, event *AuditEvent) error {
    eventBytes, err := json.Marshal(event)
    if err != nil {
        return &errors.ServiceError{
            Code: AUDIT_009,
            Message: "Event serialization failed",
            Description: "Unable to serialize audit event to JSON",
        }
    }
    
    message := &kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &a.topic,
            Partition: kafka.PartitionAny,
        },
        Key:   []byte(event.EntityID),
        Value: eventBytes,
        Headers: []kafka.Header{
            {Key: "event_type", Value: []byte(event.EventType)},
            {Key: "service", Value: []byte(event.Service)},
            {Key: "correlation_id", Value: []byte(event.CorrelationID)},
        },
    }
    
    deliveryChan := make(chan kafka.Event)
    err = a.producer.Produce(message, deliveryChan)
    if err != nil {
        return &errors.ServiceError{
            Code: AUDIT_007,
            Message: "Audit event publish failed",
            Description: "Unable to publish event to Kafka topic",
        }
    }
    
    // Wait for delivery confirmation
    select {
    case e := <-deliveryChan:
        if msg, ok := e.(*kafka.Message); ok {
            if msg.TopicPartition.Error != nil {
                a.logger.WithError(msg.TopicPartition.Error).Error("Failed to deliver audit event")
                return &errors.ServiceError{
                    Code: AUDIT_007,
                    Message: "Event delivery failed",
                    Description: msg.TopicPartition.Error.Error(),
                }
            }
        }
    case <-ctx.Done():
        return ctx.Err()
    }
    
    return nil
}
```

#### Low-Level Implementation Details

**Database Schema:**
```sql
-- Comprehensive audit log
CREATE TABLE audit_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_type VARCHAR(50) NOT NULL, -- 'user_action', 'system_event', 'decision_made'
  entity_type VARCHAR(50) NOT NULL, -- 'user', 'application', 'loan'
  entity_id UUID NOT NULL,
  user_id UUID, -- NULL for system events
  action VARCHAR(100) NOT NULL,
  old_values JSONB,
  new_values JSONB,
  metadata JSONB,
  ip_address INET,
  user_agent TEXT,
  session_id UUID,
  correlation_id UUID, -- For tracing related events
  severity VARCHAR(20) DEFAULT 'info', -- 'debug', 'info', 'warn', 'error', 'critical'
  created_at TIMESTAMP DEFAULT NOW()
);

-- Regulatory reporting data
CREATE TABLE compliance_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  report_type VARCHAR(50) NOT NULL, -- 'hmda', 'cra', 'sar', 'fair_lending'
  report_period_start DATE NOT NULL,
  report_period_end DATE NOT NULL,
  generated_by UUID NOT NULL,
  file_path VARCHAR(500), -- S3 path to generated report
  status VARCHAR(20) DEFAULT 'generating', -- 'generating', 'completed', 'filed', 'error'
  created_at TIMESTAMP DEFAULT NOW(),
  completed_at TIMESTAMP
);

-- Performance metrics
CREATE TABLE performance_metrics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  metric_name VARCHAR(100) NOT NULL,
  metric_value DECIMAL(15,4) NOT NULL,
  metric_unit VARCHAR(20), -- 'ms', 'count', 'percentage'
  dimensions JSONB, -- Additional metric dimensions
  recorded_at TIMESTAMP DEFAULT NOW()
);
```

**Key Components:**
- **Event Stream Processor:** Real-time processing of audit events via Kafka
- **Compliance Engine:** Automated regulatory report generation with Go templates
- **Retention Manager:** Automated data archival based on retention policies
- **Performance Monitor:** Real-time metrics collection using Prometheus Go client
- **Clean Architecture Layers:**
  - Domain: Event entities, compliance rules
  - Application: Audit use cases, report generation
  - Infrastructure: Kafka producers, S3 storage, database persistence
  - Interface: Gin HTTP handlers, gRPC endpoints

## Data Architecture

### Database Design Principles

1. **Data Normalization:** 3NF normalization with selective denormalization for performance
2. **Audit-First Design:** Every table includes audit fields (created_at, updated_at, created_by)
3. **Soft Deletes:** Logical deletion with retention for compliance requirements
4. **Encryption at Rest:** Field-level encryption for PII and sensitive financial data
5. **Read Replicas:** Separate read replicas for reporting and analytics

### Data Flow Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Write OPS  │    │ Event Store │    │  Read OPS   │
│             │    │   (Kafka)   │    │             │
│ • User Mgmt │───▶│             │───▶│ • Reporting │
│ • Loan Mgmt │    │ • Audit Log │    │ • Analytics │
│ • Decisions │    │ • State Chg │    │ • Dashboards│
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│Primary DB   │    │  Audit DB   │    │Analytics DB │
│(PostgreSQL) │    │(PostgreSQL) │    │(PostgreSQL)│
│             │    │             │    │Read Replica │
│• Users      │    │• Events     │    │             │
│• Apps       │    │• Metrics    │    │• Aggregated │
│• Loans      │    │• Logs       │    │  Data       │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Data Security & Compliance

**Encryption Strategy:**
- **At Rest:** AES-256 encryption for all PII data fields
- **In Transit:** TLS 1.3 for all API communications
- **Key Management:** AWS KMS with automatic key rotation

**Data Retention:**
- **Application Data:** 7 years minimum for regulatory compliance
- **Audit Logs:** 7 years with immutable storage
- **Credit Reports:** 25 months after adverse action notice
- **KYC Documents:** 5 years after account closure

## API Design

### RESTful API Standards

**Base URL Structure:**
```
https://api.lendingplatform.com/v1/{service}/{resource}
```

**Standard HTTP Methods:**
- `GET` - Retrieve resources
- `POST` - Create new resources
- `PUT` - Update entire resources
- `PATCH` - Partial resource updates
- `DELETE` - Remove resources

**Response Format:**
```json
{
  "success": true,
  "data": { /* resource data */ },
  "metadata": {
    "request_id": "uuid",
    "timestamp": "2025-08-14T10:30:00Z",
    "version": "v1",
    "service": "loan-service"
  },
  "pagination": { /* if applicable */ },
  "error": null
}
```

**Error Response Format:**
```json
{
  "success": false,
  "data": null,
  "metadata": {
    "request_id": "uuid",
    "timestamp": "2025-08-14T10:30:00Z",
    "version": "v1",
    "service": "loan-service"
  },
  "error": {
    "code": "LOAN_001",
    "message": "Invalid loan amount",
    "description": "Loan amount must be between $5,000 and $50,000",
    "field": "loan_amount",
    "metadata": {
      "min_amount": 5000,
      "max_amount": 50000,
      "provided_amount": 60000
    }
  }
}
```

### Gin Framework Implementation

**Middleware Stack:**
```go
func SetupRouter() *gin.Engine {
    router := gin.New()
    
    // Core middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(middleware.RequestID())
    router.Use(middleware.CORS())
    router.Use(middleware.HTTPSignature())
    router.Use(middleware.RateLimit())
    router.Use(middleware.ErrorHandler())
    
    // API versioning
    v1 := router.Group("/v1")
    {
        // Service groups
        auth := v1.Group("/auth")
        users := v1.Group("/users")
        applications := v1.Group("/applications")
        decisions := v1.Group("/decisions")
    }
    
    return router
}
```

**Error Handling Middleware:**
```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            var serviceErr *errors.ServiceError
            if errors.As(err.Err, &serviceErr) {
                c.JSON(serviceErr.HTTPStatus(), ErrorResponse{
                    Success: false,
                    Error:   serviceErr.ToErrorDetail(),
                    Metadata: ResponseMetadata{
                        RequestID: c.GetString("request_id"),
                        Timestamp: time.Now(),
                        Version:   "v1",
                        Service:   c.GetString("service_name"),
                    },
                })
                return
            }
            
            // Default error handling
            c.JSON(500, ErrorResponse{
                Success: false,
                Error: &ErrorDetail{
                    Code:        "SYSTEM_INTERNAL_ERROR",
                    Message:     "Internal server error",
                    Description: "An unexpected error occurred",
                },
                Metadata: ResponseMetadata{
                    RequestID: c.GetString("request_id"),
                    Timestamp: time.Now(),
                    Version:   "v1",
                    Service:   c.GetString("service_name"),
                },
            })
        }
    }
}
```

### Key API Endpoints

#### Authentication API
```
POST /v1/auth/login              # User login with HTTP signature
POST /v1/auth/refresh            # Refresh access token
POST /v1/auth/logout             # User logout
GET  /v1/auth/me                 # Get current user info
```

#### User Management API
```
POST /v1/users                   # Create user
GET  /v1/users/{userId}          # Get user profile
PUT  /v1/users/{userId}          # Update user profile
POST /v1/users/{userId}/kyc      # Initiate KYC process
GET  /v1/users/{userId}/kyc      # Get KYC status
POST /v1/users/{userId}/documents # Upload document to S3
GET  /v1/users/{userId}/documents/{docId} # Download document from S3
```

#### Loan Application API
```
POST /v1/applications            # Create application (triggers Conductor workflow)
GET  /v1/applications/{appId}    # Get application status
PUT  /v1/applications/{appId}    # Update application
POST /v1/applications/prequalify # Pre-qualification check
GET  /v1/applications/{appId}/status # Get current state
POST /v1/applications/{appId}/accept # Accept loan offer
GET  /v1/applications/{appId}/workflow # Get workflow status
```

#### Decision Engine API
```
POST /v1/decisions/evaluate      # Trigger automated decision
GET  /v1/decisions/queue         # Get manual review queue
POST /v1/decisions/manual        # Submit manual decision
GET  /v1/decisions/rules         # Get current decision rules
PUT  /v1/decisions/rules         # Update decision rules
GET  /v1/decisions/{appId}/history # Get decision history
```

#### Conductor Workflow API
```
GET  /v1/workflows/{workflowId}  # Get workflow status
POST /v1/workflows/{workflowId}/pause # Pause workflow
POST /v1/workflows/{workflowId}/resume # Resume workflow
GET  /v1/workflows/tasks/pending # Get pending manual tasks
POST /v1/workflows/tasks/{taskId}/complete # Complete manual task
```

## Security Architecture

### Authentication & Authorization

**JWT Token Strategy:**
- **Access Tokens:** 15-minute expiration, stateless validation with Go-JWT
- **Refresh Tokens:** 30-day expiration, stored in Redis with Go-Redis client
- **Token Rotation:** New refresh token issued on each refresh
- **Revocation:** Immediate token invalidation on security events

**HTTP Signature Authentication:**
```go
type HTTPSignatureConfig struct {
    Algorithm    string        // "hmac-sha256", "rsa-sha256"
    Headers      []string      // Required headers to sign
    MaxClockSkew time.Duration // Allowed time drift
    KeyStore     KeyStore      // Key management interface
}

func ValidateHTTPSignature(c *gin.Context, config *HTTPSignatureConfig) error {
    sigHeader := c.GetHeader("Signature")
    if sigHeader == "" {
        return &errors.ServiceError{
            Code: AUTH_002,
            Message: "Missing HTTP signature",
        }
    }
    
    // Parse signature components
    sigParams := parseSignatureHeader(sigHeader)
    
    // Validate timestamp (prevent replay attacks)
    if time.Since(sigParams.Created) > config.MaxClockSkew {
        return &errors.ServiceError{
            Code: AUTH_003,
            Message: "Signature timestamp too old",
        }
    }
    
    // Reconstruct signing string
    signingString := buildSigningString(c, sigParams.Headers)
    
    // Verify signature
    key, err := config.KeyStore.GetKey(sigParams.KeyID)
    if err != nil {
        return &errors.ServiceError{
            Code: AUTH_005,
            Message: "Invalid key ID",
        }
    }
    
    if !verifySignature(signingString, sigParams.Signature, key, config.Algorithm) {
        return &errors.ServiceError{
            Code: AUTH_005,
            Message: "Signature verification failed",
        }
    }
    
    return nil
}
```

**Role-Based Access Control (RBAC):**
```json
{
  "roles": {
    "customer": {
      "permissions": ["application:create", "application:read", "profile:update"]
    },
    "loan_officer": {
      "permissions": ["application:read", "decision:review", "queue:manage"]
    },
    "underwriter": {
      "permissions": ["application:read", "decision:approve", "decision:deny"]
    },
    "admin": {
      "permissions": ["*"]
    }
  }
}
```

### Data Protection

**PII Encryption:**
- **Algorithm:** AES-256-GCM for field-level encryption
- **Key Management:** AWS KMS with automatic rotation
- **Access Control:** Principle of least privilege for encryption keys

**API Security:**
- **Rate Limiting:** 100 requests per minute per user
- **CORS Policy:** Strict origin validation
- **Input Validation:** JSON Schema validation for all inputs
- **SQL Injection Prevention:** Parameterized queries only

### Compliance Controls

**SOC 2 Type II Requirements:**
- **Security:** Multi-factor authentication, encryption, access controls
- **Availability:** 99.9% uptime SLA with redundancy
- **Processing Integrity:** Data validation and error handling
- **Confidentiality:** Data classification and handling procedures
- **Privacy:** GDPR and CCPA compliance controls

## Comprehensive Error Management Framework

### Global Error Handling Strategy

**Service Error Interface:**
```go
package errors

import (
    "fmt"
    "net/http"
    "time"
)

type ServiceError struct {
    Code        string                 `json:"code"`
    Message     string                 `json:"message"`
    Description string                 `json:"description"`
    Field       string                 `json:"field,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    RetryAfter  *int                   `json:"retry_after,omitempty"`
    Cause       error                  `json:"-"` // Original error for logging
    Service     string                 `json:"service"`
    Timestamp   time.Time              `json:"timestamp"`
}

func (e *ServiceError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Description)
}

func (e *ServiceError) HTTPStatus() int {
    switch {
    case strings.HasPrefix(e.Code, "AUTH_001"),
         strings.HasPrefix(e.Code, "AUTH_002"),
         strings.HasPrefix(e.Code, "AUTH_003"),
         strings.HasPrefix(e.Code, "USER_001"),
         strings.HasPrefix(e.Code, "USER_002"),
         strings.HasPrefix(e.Code, "LOAN_001"),
         strings.HasPrefix(e.Code, "LOAN_002"),
         strings.HasPrefix(e.Code, "DECISION_001"):
        return http.StatusBadRequest
    case strings.HasPrefix(e.Code, "AUTH_"):
        return http.StatusUnauthorized
    case strings.Contains(e.Code, "USER_008"),
         strings.Contains(e.Code, "USER_009"),
         strings.Contains(e.Code, "LOAN_007"),
         strings.Contains(e.Code, "LOAN_008"):
        return http.StatusUnprocessableEntity
    case strings.Contains(e.Code, "CREDIT_"),
         strings.Contains(e.Code, "KYC_"),
         strings.Contains(e.Code, "STORAGE_"):
        return http.StatusBadGateway
    default:
        return http.StatusInternalServerError
    }
}

func (e *ServiceError) ToErrorDetail() *ErrorDetail {
    return &ErrorDetail{
        Code:        e.Code,
        Message:     e.Message,
        Description: e.Description,
        Field:       e.Field,
        Metadata:    e.Metadata,
        RetryAfter:  e.RetryAfter,
    }
}
```

### Error Code Registry by Service

#### Authentication Service Error Codes
```go
const (
    // Authentication validation errors (400)
    AUTH_001 = "AUTH_001" // Invalid credentials
    AUTH_002 = "AUTH_002" // Missing token
    AUTH_003 = "AUTH_003" // Expired token
    AUTH_004 = "AUTH_004" // Malformed token
    AUTH_005 = "AUTH_005" // Invalid signature
    AUTH_006 = "AUTH_006" // Missing signature
    AUTH_007 = "AUTH_007" // Invalid request format
    
    // Authentication business errors (401/403)
    AUTH_008 = "AUTH_008" // Account locked
    AUTH_009 = "AUTH_009" // Account suspended
    AUTH_010 = "AUTH_010" // Password expired
    AUTH_011 = "AUTH_011" // Too many attempts
    AUTH_012 = "AUTH_012" // Insufficient privileges
    AUTH_013 = "AUTH_013" // Session expired
    
    // Authentication external errors (502)
    AUTH_014 = "AUTH_014" // MFA service unavailable
    AUTH_015 = "AUTH_015" // Identity provider error
    AUTH_016 = "AUTH_016" // Key management error
    
    // Authentication system errors (500)
    AUTH_017 = "AUTH_017" // Session store error
    AUTH_018 = "AUTH_018" // Token generation error
    AUTH_019 = "AUTH_019" // Encryption error
)
```

#### User Management Service Error Codes
```go
const (
    // User validation errors (400)
    USER_001 = "USER_001" // Invalid email
    USER_002 = "USER_002" // Invalid phone
    USER_003 = "USER_003" // Invalid SSN
    USER_004 = "USER_004" // Invalid DOB
    USER_005 = "USER_005" // Missing required field
    USER_006 = "USER_006" // Invalid address
    USER_007 = "USER_007" // Invalid employment
    
    // User business errors (422)
    USER_008 = "USER_008" // Email already exists
    USER_009 = "USER_009" // Phone already exists
    USER_010 = "USER_010" // SSN already exists
    USER_011 = "USER_011" // Under age
    USER_012 = "USER_012" // KYC already completed
    USER_013 = "USER_013" // Profile incomplete
    USER_014 = "USER_014" // Verification failed
    
    // User document errors (400/422)
    USER_015 = "USER_015" // Invalid document format
    USER_016 = "USER_016" // File too large
    USER_017 = "USER_017" // Upload failed
    USER_018 = "USER_018" // Document not found
    USER_019 = "USER_019" // Encryption failed
    USER_020 = "USER_020" // Virus detected
    USER_021 = "USER_021" // Processing timeout
    
    // User external errors (502)
    USER_022 = "USER_022" // S3 upload failed
    USER_023 = "USER_023" // KYC provider error
    USER_024 = "USER_024" // Address validation error
    
    // User system errors (500)
    USER_025 = "USER_025" // Database error
    USER_026 = "USER_026" // Encryption error
    USER_027 = "USER_027" // Profile processing error
)
```

#### Loan Management Service Error Codes
```go
const (
    // Loan validation errors (400)
    LOAN_001 = "LOAN_001" // Invalid amount
    LOAN_002 = "LOAN_002" // Invalid purpose
    LOAN_003 = "LOAN_003" // Invalid term
    LOAN_004 = "LOAN_004" // Invalid income
    LOAN_005 = "LOAN_005" // Invalid employment
    LOAN_006 = "LOAN_006" // Missing documents
    
    // Loan business errors (422)
    LOAN_007 = "LOAN_007" // Amount too low
    LOAN_008 = "LOAN_008" // Amount too high
    LOAN_009 = "LOAN_009" // Insufficient income
    LOAN_010 = "LOAN_010" // Invalid state transition
    LOAN_011 = "LOAN_011" // Offer expired
    LOAN_012 = "LOAN_012" // Application not found
    LOAN_013 = "LOAN_013" // Duplicate application
    LOAN_014 = "LOAN_014" // Eligibility failed
    
    // Loan workflow errors (422)
    LOAN_015 = "LOAN_015" // Workflow start failed
    LOAN_016 = "LOAN_016" // Workflow execution failed
    LOAN_017 = "LOAN_017" // State conflict
    LOAN_018 = "LOAN_018" // Task timeout
    LOAN_019 = "LOAN_019" // Invalid transition
    
    // Loan external errors (502)
    LOAN_020 = "LOAN_020" // Conductor unavailable
    LOAN_021 = "LOAN_021" // Decision engine error
    LOAN_022 = "LOAN_022" // Credit check failed
    LOAN_023 = "LOAN_023" // Payment processor error
    
    // Loan system errors (500)
    LOAN_024 = "LOAN_024" // State machine error
    LOAN_025 = "LOAN_025" // Offer calculation error
    LOAN_026 = "LOAN_026" // Workflow persistence error
)
```

#### Decision Engine Service Error Codes
```go
const (
    // Decision validation errors (400)
    DECISION_001 = "DECISION_001" // Invalid application
    DECISION_002 = "DECISION_002" // Invalid rules
    DECISION_003 = "DECISION_003" // Invalid override
    DECISION_004 = "DECISION_004" // Missing data
    
    // Decision business errors (422)
    DECISION_005 = "DECISION_005" // Insufficient data
    DECISION_006 = "DECISION_006" // Conflicting rules
    DECISION_007 = "DECISION_007" // Manual review required
    DECISION_008 = "DECISION_008" // SLA violated
    DECISION_009 = "DECISION_009" // Reviewer not qualified
    DECISION_010 = "DECISION_010" // Queue full
    DECISION_011 = "DECISION_011" // Decision locked
    
    // Decision external errors (502)
    DECISION_012 = "DECISION_012" // Credit bureau error
    DECISION_013 = "DECISION_013" // Fraud service error
    DECISION_014 = "DECISION_014" // Conductor task failed
    DECISION_015 = "DECISION_015" // ML model error
    
    // Decision system errors (500)
    DECISION_016 = "DECISION_016" // Rules engine error
    DECISION_017 = "DECISION_017" // Queue management error
    DECISION_018 = "DECISION_018" // Scoring error
)
```

#### Integration Services Error Codes
```go
const (
    // Credit Bureau errors
    CREDIT_001 = "CREDIT_001" // Invalid SSN
    CREDIT_002 = "CREDIT_002" // Invalid DOB
    CREDIT_003 = "CREDIT_003" // Thin file
    CREDIT_004 = "CREDIT_004" // Frozen report
    CREDIT_005 = "CREDIT_005" // Bureau unavailable
    CREDIT_006 = "CREDIT_006" // Rate limited
    CREDIT_007 = "CREDIT_007" // Circuit breaker open
    
    // KYC/AML errors
    KYC_001 = "KYC_001" // Invalid document
    KYC_002 = "KYC_002" // Document expired
    KYC_003 = "KYC_003" // Identity mismatch
    KYC_004 = "KYC_004" // Sanctions hit
    KYC_005 = "KYC_005" // PEP identified
    KYC_006 = "KYC_006" // Provider error
    KYC_007 = "KYC_007" // Document processing failed
    
    // Document Storage errors
    STORAGE_001 = "STORAGE_001" // Invalid file type
    STORAGE_002 = "STORAGE_002" // Quota exceeded
    STORAGE_003 = "STORAGE_003" // S3 error
    STORAGE_004 = "STORAGE_004" // Encryption error
)
```

#### Audit Service Error Codes
```go
const (
    // Audit validation errors (400)
    AUDIT_001 = "AUDIT_001" // Invalid event type
    AUDIT_002 = "AUDIT_002" // Missing entity ID
    AUDIT_003 = "AUDIT_003" // Invalid date range
    AUDIT_004 = "AUDIT_004" // Invalid search criteria
    
    // Audit business errors (422)
    AUDIT_005 = "AUDIT_005" // Report not found
    AUDIT_006 = "AUDIT_006" // Retention policy violation
    AUDIT_007 = "AUDIT_007" // Duplicate event
    AUDIT_008 = "AUDIT_008" // Access denied
    
    // Audit external errors (502)
    AUDIT_009 = "AUDIT_009" // Kafka publish failed
    AUDIT_010 = "AUDIT_010" // S3 storage failed
    AUDIT_011 = "AUDIT_011" // Compliance API error
    
    // Audit system errors (500)
    AUDIT_012 = "AUDIT_012" // Event processing error
    AUDIT_013 = "AUDIT_013" // Report generation error
    AUDIT_014 = "AUDIT_014" // Indexing error
)
```

### Error Handler Implementation

**Centralized Error Handler:**
```go
package middleware

import (
    "context"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "our-los/pkg/errors"
)

type ErrorHandler struct {
    logger *logrus.Logger
    audit  AuditService
}

func NewErrorHandler(logger *logrus.Logger, audit AuditService) *ErrorHandler {
    return &ErrorHandler{
        logger: logger,
        audit:  audit,
    }
}

func (eh *ErrorHandler) Handle() gin.HandlerFunc {
    return gin.ErrorLoggerT(gin.ErrorTypeAny, func(c *gin.Context, err error) {
        requestID := c.GetString("request_id")
        serviceName := c.GetString("service_name")
        
        // Log error for debugging
        eh.logger.WithFields(logrus.Fields{
            "request_id": requestID,
            "service":    serviceName,
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "error":      err.Error(),
        }).Error("Request error occurred")
        
        // Audit log the error
        eh.auditError(c, err, requestID, serviceName)
        
        var serviceErr *errors.ServiceError
        if errors.As(err, &serviceErr) {
            eh.handleServiceError(c, serviceErr, requestID, serviceName)
            return
        }
        
        // Handle unknown errors
        eh.handleUnknownError(c, err, requestID, serviceName)
    })
}

func (eh *ErrorHandler) handleServiceError(c *gin.Context, err *errors.ServiceError, requestID, serviceName string) {
    response := ErrorResponse{
        Success: false,
        Data:    nil,
        Error:   err.ToErrorDetail(),
        Metadata: ResponseMetadata{
            RequestID: requestID,
            Timestamp: time.Now(),
            Version:   "v1",
            Service:   serviceName,
        },
    }
    
    c.JSON(err.HTTPStatus(), response)
}

func (eh *ErrorHandler) handleUnknownError(c *gin.Context, err error, requestID, serviceName string) {
    response := ErrorResponse{
        Success: false,
        Data:    nil,
        Error: &ErrorDetail{
            Code:        "SYSTEM_INTERNAL_ERROR",
            Message:     "Internal server error",
            Description: "An unexpected error occurred. Please try again later.",
        },
        Metadata: ResponseMetadata{
            RequestID: requestID,
            Timestamp: time.Now(),
            Version:   "v1",
            Service:   serviceName,
        },
    }
    
    c.JSON(http.StatusInternalServerError, response)
}

func (eh *ErrorHandler) auditError(c *gin.Context, err error, requestID, serviceName string) {
    auditEvent := &AuditEvent{
        EventType:     "error_occurred",
        EntityType:    "http_request",
        EntityID:      requestID,
        UserID:        c.GetString("user_id"),
        Action:        "request_processing",
        Metadata: map[string]interface{}{
            "error_message": err.Error(),
            "method":        c.Request.Method,
            "path":          c.Request.URL.Path,
            "service":       serviceName,
            "user_agent":    c.GetHeader("User-Agent"),
        },
        IPAddress:     c.ClientIP(),
        CorrelationID: requestID,
        Severity:      "error",
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := eh.audit.LogEvent(ctx, auditEvent); err != nil {
        eh.logger.WithError(err).Error("Failed to audit error event")
    }
}
```

### Circuit Breaker for External Services

**Circuit Breaker Implementation:**
```go
package resilience

import (
    "context"
    "sync"
    "time"
    
    "our-los/pkg/errors"
)

type CircuitBreakerState int

const (
    Closed CircuitBreakerState = iota
    Open
    HalfOpen
)

type CircuitBreaker struct {
    mu                  sync.RWMutex
    state               CircuitBreakerState
    failureCount        int
    lastFailureTime     time.Time
    successCount        int
    failureThreshold    int
    recoveryTimeout     time.Duration
    halfOpenMaxRequests int
    onStateChange       func(from, to CircuitBreakerState)
}

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
    return &CircuitBreaker{
        state:               Closed,
        failureThreshold:    config.FailureThreshold,
        recoveryTimeout:     config.RecoveryTimeout,
        halfOpenMaxRequests: config.HalfOpenMaxRequests,
        onStateChange:       config.OnStateChange,
    }
}

func (cb *CircuitBreaker) Call(ctx context.Context, operation func(context.Context) error, service string) error {
    if !cb.allowRequest() {
        return &errors.ServiceError{
            Code: service + "_007",
            Message:     "Service temporarily unavailable",
            Description: "Circuit breaker is open due to repeated failures",
            RetryAfter:  &[]int{int(cb.recoveryTimeout.Seconds())}[0],
            Service:     service,
            Timestamp:   time.Now(),
        }
    }
    
    err := operation(ctx)
    cb.recordResult(err == nil)
    
    return err
}

func (cb *CircuitBreaker) allowRequest() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    
    switch cb.state {
    case Closed:
        return true
    case Open:
        return time.Since(cb.lastFailureTime) >= cb.recoveryTimeout
    case HalfOpen:
        return cb.successCount < cb.halfOpenMaxRequests
    default:
        return false
    }
}

func (cb *CircuitBreaker) recordResult(success bool) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if success {
        cb.onSuccess()
    } else {
        cb.onFailure()
    }
}

func (cb *CircuitBreaker) onSuccess() {
    switch cb.state {
    case HalfOpen:
        cb.successCount++
        if cb.successCount >= cb.halfOpenMaxRequests {
            cb.setState(Closed)
            cb.reset()
        }
    case Closed:
        cb.reset()
    }
}

func (cb *CircuitBreaker) onFailure() {
    cb.failureCount++
    cb.lastFailureTime = time.Now()
    
    switch cb.state {
    case Closed:
        if cb.failureCount >= cb.failureThreshold {
            cb.setState(Open)
        }
    case HalfOpen:
        cb.setState(Open)
    }
}

func (cb *CircuitBreaker) setState(state CircuitBreakerState) {
    if cb.state == state {
        return
    }
    
    oldState := cb.state
    cb.state = state
    
    if cb.onStateChange != nil {
        cb.onStateChange(oldState, state)
    }
}

func (cb *CircuitBreaker) reset() {
    cb.failureCount = 0
    cb.successCount = 0
}
```

This comprehensive error management framework provides:

1. **Standardized Error Codes** - Consistent naming and categorization across all services
2. **HTTP Status Mapping** - Automatic mapping of error codes to appropriate HTTP status codes
3. **Rich Error Context** - Detailed error information with metadata and retry guidance
4. **Circuit Breaker Protection** - Automatic failover for external service dependencies
5. **Audit Integration** - Error events are automatically logged for compliance
6. **Clean Architecture Compliance** - Error handling follows clean architecture principles
7. **Go-Gin Integration** - Native middleware support for Gin framework
8. **HTTP Signature Support** - Enhanced security with request signing validation

## Performance & Scalability

### Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Response Time | <500ms (95th percentile) | Application endpoints |
| Page Load Time | <3 seconds | Full page on 3G connection |
| Database Query Time | <100ms | Standard CRUD operations |
| Credit Bureau Response | <15 seconds | External API calls |
| Decision Engine Processing | <30 seconds | Automated underwriting |
| System Availability | 99.9% | Monthly uptime target |

### Scalability Strategy

**Horizontal Scaling:**
- **Stateless Services:** All application services are stateless for easy scaling
- **Load Balancing:** Application Load Balancer with health checks
- **Auto Scaling:** CPU and memory-based scaling triggers
- **Database Scaling:** Read replicas for reporting and analytics

**Caching Strategy:**
- **API Gateway:** Response caching for static data (5 minutes)
- **Application Cache:** Redis for session data and frequently accessed data
- **Database Cache:** Query result caching for complex reports
- **CDN:** Static asset caching with CloudFront

**Performance Monitoring:**
- **APM:** Application Performance Monitoring with New Relic
- **Real User Monitoring:** Client-side performance tracking
- **Synthetic Monitoring:** Automated performance testing
- **Alert Thresholds:** Performance degradation alerts

## Deployment & Infrastructure

### Cloud Infrastructure (AWS)

**Compute:**
- **ECS Fargate:** Container orchestration for microservices
- **Application Load Balancer:** Traffic distribution and SSL termination
- **Auto Scaling Groups:** Automatic capacity management

**Data Storage:**
- **RDS PostgreSQL:** Primary database with Multi-AZ deployment
- **ElastiCache Redis:** Session storage and application cache
- **S3:** Document storage with encryption and versioning

**Security & Networking:**
- **VPC:** Private network with public/private subnets
- **Security Groups:** Fine-grained network access control
- **WAF:** Web Application Firewall for DDoS protection
- **Secrets Manager:** Secure storage for API keys and credentials

### Deployment Pipeline

**CI/CD Workflow:**
1. **Code Commit:** Developer pushes to feature branch
2. **Automated Testing:** Unit tests, integration tests, security scans
3. **Code Review:** Peer review and approval process
4. **Build & Package:** Docker image creation and vulnerability scanning
5. **Staging Deployment:** Automatic deployment to staging environment
6. **End-to-End Testing:** Automated testing against staging
7. **Production Deployment:** Blue-green deployment with rollback capability

**Environment Strategy:**
- **Development:** Individual developer environments
- **Staging:** Production-like environment for integration testing
- **Production:** High-availability production environment

### Monitoring & Observability

**Application Monitoring:**
- **Prometheus:** Metrics collection and storage
- **Grafana:** Metrics visualization and dashboards
- **AlertManager:** Alert routing and notification management

**Log Management:**
- **ELK Stack:** Elasticsearch, Logstash, Kibana for log analysis
- **Structured Logging:** JSON-formatted logs with correlation IDs
- **Log Retention:** 90 days for application logs, 7 years for audit logs

**Health Checks:**
- **Service Health:** HTTP health check endpoints
- **Database Health:** Connection pool monitoring
- **External Service Health:** Integration endpoint monitoring
- **Business Metrics:** Real-time business KPI monitoring

## Risk Mitigation & Disaster Recovery

### Business Continuity

**Disaster Recovery Plan:**
- **RTO (Recovery Time Objective):** 4 hours maximum downtime
- **RPO (Recovery Point Objective):** 15 minutes maximum data loss
- **Backup Strategy:** Continuous database replication to secondary region
- **Failover Testing:** Monthly disaster recovery testing

**Data Backup:**
- **Database Backups:** Daily automated backups with 30-day retention
- **Document Backups:** Cross-region replication for S3 storage
- **Configuration Backups:** Infrastructure as Code with version control

### Security Risk Mitigation

**Threat Detection:**
- **SIEM:** Security Information and Event Management system
- **Anomaly Detection:** Machine learning-based fraud detection
- **Penetration Testing:** Quarterly security assessments
- **Vulnerability Management:** Automated security scanning

**Incident Response:**
- **Security Team:** 24/7 security operations center
- **Response Plan:** Documented incident response procedures
- **Communication Plan:** Customer and stakeholder notification procedures
- **Forensics:** Evidence preservation and analysis capabilities

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)
- Core infrastructure setup (AWS, databases, monitoring)
- Authentication and user management services
- Basic API gateway and security framework
- Development environment and CI/CD pipeline

### Phase 2: Core Services (Weeks 5-8)
- Loan management service with state machine
- Decision engine with basic rules
- Credit bureau integration
- Document upload and storage

### Phase 3: Advanced Features (Weeks 9-12)
- KYC/AML integration
- E-signature integration
- Manual review queue system
- Payment and funding services

### Phase 4: Compliance & Launch (Weeks 13-16)
- Comprehensive audit system
- Regulatory reporting
- Performance optimization
- Security hardening and penetration testing

### Phase 5: Post-Launch (Weeks 17-20)
- Performance monitoring and optimization
- Feature enhancements based on user feedback
- Advanced analytics and reporting
- Preparation for mobile app development

## Conclusion

This architecture provides a robust, scalable, and compliant foundation for the Loan Origination System. The microservices approach enables independent development and scaling of components, while the event-driven audit system ensures regulatory compliance. The extensive use of external integrations (credit bureaus, KYC providers, e-signature) allows the system to focus on core lending business logic while leveraging specialized services for complex operations.

Key architectural strengths:
- **Scalability:** Microservices with auto-scaling capability
- **Security:** Multi-layered security with encryption and audit trails
- **Compliance:** Built-in regulatory compliance and reporting
- **Performance:** Optimized for sub-second response times
- **Maintainability:** Clear service boundaries and API contracts
- **Extensibility:** Plugin architecture for future enhancements

The implementation roadmap provides a clear path to MVP delivery within 16 weeks while establishing a foundation for future growth and feature development.
