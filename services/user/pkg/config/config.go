package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the user service
type Config struct {
	Environment string        `yaml:"environment" json:"environment"`
	Service     ServiceConfig `yaml:"service" json:"service"`
	Server      ServerConfig  `yaml:"server" json:"server"`
	Database    DatabaseConfig `yaml:"database" json:"database"`
	Redis       RedisConfig   `yaml:"redis" json:"redis"`
	Storage     StorageConfig `yaml:"storage" json:"storage"`
	Encryption  EncryptionConfig `yaml:"encryption" json:"encryption"`
	Logging     LoggingConfig `yaml:"logging" json:"logging"`
	Security    SecurityConfig `yaml:"security" json:"security"`
	ExternalServices ExternalServicesConfig `yaml:"external_services" json:"external_services"`
	Features    FeaturesConfig `yaml:"features" json:"features"`
	I18n        I18nConfig    `yaml:"i18n" json:"i18n"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name    string `yaml:"name" json:"name"`
	Port    int    `yaml:"port" json:"port"`
	Version string `yaml:"version" json:"version"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string `yaml:"host" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	ReadTimeout  int    `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout" json:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string `yaml:"host" json:"host"`
	Port            int    `yaml:"port" json:"port"`
	Name            string `yaml:"name" json:"name"`
	User            string `yaml:"user" json:"user"`
	Password        string `yaml:"password" json:"password"`
	SSLMode         string `yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string `yaml:"host" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	Password     string `yaml:"password" json:"password"`
	DB           int    `yaml:"db" json:"db"`
	PoolSize     int    `yaml:"pool_size" json:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns" json:"min_idle_conns"`
	MaxRetries   int    `yaml:"max_retries" json:"max_retries"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Provider  string `yaml:"provider" json:"provider"`
	Bucket    string `yaml:"bucket" json:"bucket"`
	Region    string `yaml:"region" json:"region"`
	AccessKey string `yaml:"access_key" json:"access_key"`
	SecretKey string `yaml:"secret_key" json:"secret_key"`
	Endpoint  string `yaml:"endpoint" json:"endpoint"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	MasterKey        string `yaml:"master_key" json:"master_key"`
	KeyRotationDays  int    `yaml:"key_rotation_days" json:"key_rotation_days"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	FilePath   string `yaml:"file_path" json:"file_path"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret           string `yaml:"jwt_secret" json:"jwt_secret"`
	BcryptCost          int    `yaml:"bcrypt_cost" json:"bcrypt_cost"`
	RateLimitRequests   int    `yaml:"rate_limit_requests" json:"rate_limit_requests"`
	RateLimitWindow     int    `yaml:"rate_limit_window" json:"rate_limit_window"`
}

// ExternalServicesConfig holds external service configurations
type ExternalServicesConfig struct {
	KYCProvider         ServiceEndpoint `yaml:"kyc_provider" json:"kyc_provider"`
	NotificationService ServiceEndpoint `yaml:"notification_service" json:"notification_service"`
	AuditService        ServiceEndpoint `yaml:"audit_service" json:"audit_service"`
}

// ServiceEndpoint holds service endpoint configuration
type ServiceEndpoint struct {
	URL     string `yaml:"url" json:"url"`
	APIKey  string `yaml:"api_key" json:"api_key"`
	Timeout int    `yaml:"timeout" json:"timeout"`
}

// FeaturesConfig holds feature flags
type FeaturesConfig struct {
	Enable2FA          bool     `yaml:"enable_2fa" json:"enable_2fa"`
	EnableDocumentOCR  bool     `yaml:"enable_document_ocr" json:"enable_document_ocr"`
	EnableRealKYC      bool     `yaml:"enable_real_kyc" json:"enable_real_kyc"`
	MaxFileSize        int64    `yaml:"max_file_size" json:"max_file_size"`
	AllowedFileTypes   []string `yaml:"allowed_file_types" json:"allowed_file_types"`
}

// I18nConfig holds internationalization configuration
type I18nConfig struct {
	DefaultLanguage    string   `yaml:"default_language" json:"default_language"`
	SupportedLanguages []string `yaml:"supported_languages" json:"supported_languages"`
	FallbackLanguage   string   `yaml:"fallback_language" json:"fallback_language"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Default configuration
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Service: ServiceConfig{
			Name:    "user-service",
			Port:    getEnvInt("SERVICE_PORT", 8082),
			Version: "1.0.0",
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("PORT", 8082),
			ReadTimeout:  getEnvInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "user_service_db"),
			User:            getEnv("DB_USER", "user_service_user"),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 1),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 3),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		},
		Storage: StorageConfig{
			Provider:  getEnv("STORAGE_PROVIDER", "s3"),
			Bucket:    getEnv("STORAGE_BUCKET", "user-service-documents"),
			Region:    getEnv("STORAGE_REGION", "us-west-2"),
			AccessKey: getEnv("STORAGE_ACCESS_KEY", ""),
			SecretKey: getEnv("STORAGE_SECRET_KEY", ""),
			Endpoint:  getEnv("STORAGE_ENDPOINT", ""),
		},
		Encryption: EncryptionConfig{
			MasterKey:       getEnv("ENCRYPTION_MASTER_KEY", ""),
			KeyRotationDays: getEnvInt("KEY_ROTATION_DAYS", 90),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			FilePath:   getEnv("LOG_FILE_PATH", "logs/user-service.log"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 5),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 30),
		},
		Security: SecurityConfig{
			JWTSecret:         getEnv("JWT_SECRET", ""),
			BcryptCost:        getEnvInt("BCRYPT_COST", 12),
			RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 60),
		},
		ExternalServices: ExternalServicesConfig{
			KYCProvider: ServiceEndpoint{
				URL:     getEnv("KYC_PROVIDER_URL", ""),
				APIKey:  getEnv("KYC_PROVIDER_API_KEY", ""),
				Timeout: getEnvInt("KYC_PROVIDER_TIMEOUT", 30),
			},
			NotificationService: ServiceEndpoint{
				URL:     getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8084"),
				Timeout: getEnvInt("NOTIFICATION_SERVICE_TIMEOUT", 10),
			},
			AuditService: ServiceEndpoint{
				URL:     getEnv("AUDIT_SERVICE_URL", "http://localhost:8085"),
				Timeout: getEnvInt("AUDIT_SERVICE_TIMEOUT", 5),
			},
		},
		Features: FeaturesConfig{
			Enable2FA:         getEnvBool("ENABLE_2FA", true),
			EnableDocumentOCR: getEnvBool("ENABLE_DOCUMENT_OCR", false),
			EnableRealKYC:     getEnvBool("ENABLE_REAL_KYC", false),
			MaxFileSize:       getEnvInt64("MAX_FILE_SIZE", 52428800), // 50MB
			AllowedFileTypes: getEnvStringSlice("ALLOWED_FILE_TYPES", []string{
				"image/jpeg", "image/png", "application/pdf",
			}),
		},
		I18n: I18nConfig{
			DefaultLanguage:    getEnv("I18N_DEFAULT_LANGUAGE", "en"),
			SupportedLanguages: getEnvStringSlice("I18N_SUPPORTED_LANGUAGES", []string{"en", "vi"}),
			FallbackLanguage:   getEnv("I18N_FALLBACK_LANGUAGE", "en"),
		},
	}

	// Load from config file if provided
	if configPath != "" {
		environment := getEnv("ENVIRONMENT", "development")
		configFile := filepath.Join(configPath, fmt.Sprintf("%s.yaml", environment))

		if _, err := os.Stat(configFile); err == nil {
			data, err := os.ReadFile(configFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	
	if cfg.Encryption.MasterKey == "" && cfg.Environment == "production" {
		return fmt.Errorf("encryption master key is required in production")
	}
	
	if cfg.Security.JWTSecret == "" && cfg.Environment == "production" {
		return fmt.Errorf("JWT secret is required in production")
	}
	
	return nil
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}
