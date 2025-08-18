package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Environment string            `yaml:"environment"`
	Service     ServiceConfig     `yaml:"service"`
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Redis       RedisConfig       `yaml:"redis"`
	Conductor   ConductorConfig   `yaml:"conductor"`
	Services    ServicesConfig    `yaml:"services"`
	Logging     LoggingConfig     `yaml:"logging"`
	Security    SecurityConfig    `yaml:"security"`
	Application ApplicationConfig `yaml:"application"`
	I18n        I18nConfig        `yaml:"i18n"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Performance PerformanceConfig `yaml:"performance,omitempty"`
	Backup      BackupConfig      `yaml:"backup,omitempty"`
}

// ServiceConfig represents service-level configuration
type ServiceConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Port    string `yaml:"port"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port                    int        `yaml:"port"`
	Host                    string     `yaml:"host"`
	ReadTimeout             int        `yaml:"read_timeout"`
	WriteTimeout            int        `yaml:"write_timeout"`
	GracefulShutdownTimeout int        `yaml:"graceful_shutdown_timeout"`
	IdleTimeout             int        `yaml:"idle_timeout"`
	CORS                    CORSConfig `yaml:"cors"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Origins     []string `yaml:"origins"`
	Methods     []string `yaml:"methods"`
	Headers     []string `yaml:"headers"`
	Credentials bool     `yaml:"credentials"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// ConductorConfig represents Conductor configuration
type ConductorConfig struct {
	BaseURL           string `yaml:"base_url"`
	Timeout           int    `yaml:"timeout"`
	RetryAttempts     int    `yaml:"retry_attempts"`
	RetryDelay        int    `yaml:"retry_delay"`
	WorkerPoolSize    int    `yaml:"worker_pool_size"`
	PollingIntervalMs int    `yaml:"polling_interval_ms"`
	UpdateRetryTimeMs int    `yaml:"update_retry_time_ms"`
}

// ServicesConfig represents external services configuration
type ServicesConfig struct {
	CreditBureau       ServiceEndpointConfig `yaml:"credit_bureau"`
	RiskScoring        ServiceEndpointConfig `yaml:"risk_scoring"`
	IncomeVerification ServiceEndpointConfig `yaml:"income_verification"`
	DecisionEngine     ServiceEndpointConfig `yaml:"decision_engine"`
	Notification       ServiceEndpointConfig `yaml:"notification"`
}

// ServiceEndpointConfig represents a service endpoint configuration
type ServiceEndpointConfig struct {
	Provider           string `yaml:"provider"`
	BaseURL            string `yaml:"base_url"`
	APIKey             string `yaml:"api_key"`
	TimeoutSeconds     int    `yaml:"timeout_seconds"`
	MockEnabled        bool   `yaml:"mock_enabled,omitempty"`
	RateLimitPerMinute int    `yaml:"rate_limit_per_minute,omitempty"`
	CacheTTL           string `yaml:"cache_ttl,omitempty"`
	WebhookEnabled     bool   `yaml:"webhook_enabled,omitempty"`
	FallbackEnabled    bool   `yaml:"fallback_enabled,omitempty"`
	RetryEnabled       bool   `yaml:"retry_enabled,omitempty"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level              string `yaml:"level"`
	Format             string `yaml:"format"`
	Output             string `yaml:"output"`
	EnableConsole      bool   `yaml:"enable_console"`
	EnableFile         bool   `yaml:"enable_file"`
	FilePath           string `yaml:"file_path"`
	MaxSize            int    `yaml:"max_size"`
	MaxAge             int    `yaml:"max_age"`
	MaxBackups         int    `yaml:"max_backups"`
	CompressionEnabled bool   `yaml:"compression_enabled,omitempty"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWTSecret            string   `yaml:"jwt_secret"`
	JWTExpiration        int      `yaml:"jwt_expiration"`
	BcryptCost           int      `yaml:"bcrypt_cost"`
	CORSAllowedOrigins   []string `yaml:"cors_allowed_origins"`
	CORSAllowedMethods   []string `yaml:"cors_allowed_methods"`
	CORSAllowedHeaders   []string `yaml:"cors_allowed_headers"`
	RateLimitingEnabled  bool     `yaml:"rate_limiting_enabled,omitempty"`
	MaxRequestsPerMinute int      `yaml:"max_requests_per_minute,omitempty"`
}

// ApplicationConfig represents application-specific configuration
type ApplicationConfig struct {
	Name                  string  `yaml:"name"`
	Version               string  `yaml:"version"`
	Environment           string  `yaml:"environment"`
	MaxLoanAmount         float64 `yaml:"max_loan_amount"`
	MinLoanAmount         float64 `yaml:"min_loan_amount"`
	MaxDTIRatio           float64 `yaml:"max_dti_ratio"`
	DefaultInterestRate   float64 `yaml:"default_interest_rate"`
	MaxInterestRate       float64 `yaml:"max_interest_rate"`
	MinInterestRate       float64 `yaml:"min_interest_rate"`
	OfferExpirationHours  int     `yaml:"offer_expiration_hours"`
	DebugMode             bool    `yaml:"debug_mode,omitempty"`
	EnableMockServices    bool    `yaml:"enable_mock_services,omitempty"`
	PerformanceMonitoring bool    `yaml:"performance_monitoring,omitempty"`
}

// I18nConfig represents internationalization configuration
type I18nConfig struct {
	DefaultLanguage    string   `yaml:"default_language"`
	SupportedLanguages []string `yaml:"supported_languages"`
	FallbackLanguage   string   `yaml:"fallback_language"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	MetricsEnabled     bool   `yaml:"metrics_enabled"`
	HealthCheckEnabled bool   `yaml:"health_check_enabled"`
	ProfilingEnabled   bool   `yaml:"profiling_enabled"`
	TraceEnabled       bool   `yaml:"trace_enabled"`
	PrometheusEndpoint string `yaml:"prometheus_endpoint,omitempty"`
	JaegerEndpoint     string `yaml:"jaeger_endpoint,omitempty"`
	AlertingEnabled    bool   `yaml:"alerting_enabled,omitempty"`
	SLAMonitoring      bool   `yaml:"sla_monitoring,omitempty"`
	BusinessMetrics    bool   `yaml:"business_metrics,omitempty"`
}

// PerformanceConfig represents performance configuration
type PerformanceConfig struct {
	ConnectionPoolSize    int    `yaml:"connection_pool_size,omitempty"`
	WorkerThreads         int    `yaml:"worker_threads,omitempty"`
	MaxConcurrentRequests int    `yaml:"max_concurrent_requests,omitempty"`
	RequestTimeout        int    `yaml:"request_timeout,omitempty"`
	CircuitBreakerEnabled bool   `yaml:"circuit_breaker_enabled,omitempty"`
	CacheEnabled          bool   `yaml:"cache_enabled,omitempty"`
	CacheTTL              string `yaml:"cache_ttl,omitempty"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	Enabled           bool   `yaml:"enabled,omitempty"`
	Schedule          string `yaml:"schedule,omitempty"`
	RetentionDays     int    `yaml:"retention_days,omitempty"`
	StorageType       string `yaml:"storage_type,omitempty"`
	S3Bucket          string `yaml:"s3_bucket,omitempty"`
	EncryptionEnabled bool   `yaml:"encryption_enabled,omitempty"`
}

// LoadConfig loads configuration from files and environment variables
func LoadConfig(configDir string) (*Config, error) {
	// Determine environment
	env := getEnvironment()

	// Load base configuration
	baseConfig, err := loadYAMLFile(filepath.Join(configDir, "base.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	// Load environment-specific configuration
	envConfig, err := loadYAMLFile(filepath.Join(configDir, env+".yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load %s config: %w", env, err)
	}

	// Merge configurations (environment overrides base)
	mergedConfig := mergeConfigs(baseConfig, envConfig)

	// Parse into struct
	var config Config
	if err := yaml.Unmarshal(mergedConfig, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set environment
	config.Environment = env

	// Substitute environment variables
	substituteEnvVars(&config)

	return &config, nil
}

// getEnvironment determines the current environment
func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = "development" // Default to development
	}
	return strings.ToLower(env)
}

// loadYAMLFile loads a YAML file and returns its content
func loadYAMLFile(filepath string) ([]byte, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// mergeConfigs merges two YAML configurations (env overrides base)
func mergeConfigs(base, env []byte) []byte {
	// For simplicity, we'll just return the environment config
	// In a more sophisticated implementation, you might want to do deep merging
	return env
}

// substituteEnvVars substitutes environment variables in the configuration
func substituteEnvVars(config *Config) {
	// This is a simplified version - in practice, you might want to use a library
	// like github.com/joho/godotenv or implement recursive substitution

	// Example substitutions for database
	if config.Database.Password == "${DB_PASSWORD}" {
		config.Database.Password = os.Getenv("DB_PASSWORD")
	}

	// Example substitutions for services
	if config.Services.CreditBureau.APIKey == "${CREDIT_BUREAU_API_KEY}" {
		config.Services.CreditBureau.APIKey = os.Getenv("CREDIT_BUREAU_API_KEY")
	}

	if config.Services.RiskScoring.APIKey == "${RISK_SCORING_API_KEY}" {
		config.Services.RiskScoring.APIKey = os.Getenv("RISK_SCORING_API_KEY")
	}

	if config.Services.IncomeVerification.APIKey == "${INCOME_VERIFICATION_API_KEY}" {
		config.Services.IncomeVerification.APIKey = os.Getenv("INCOME_VERIFICATION_API_KEY")
	}

	if config.Services.DecisionEngine.APIKey == "${DECISION_ENGINE_API_KEY}" {
		config.Services.DecisionEngine.APIKey = os.Getenv("DECISION_ENGINE_API_KEY")
	}

	if config.Services.Notification.APIKey == "${NOTIFICATION_API_KEY}" {
		config.Services.Notification.APIKey = os.Getenv("NOTIFICATION_API_KEY")
	}

	// Security substitutions
	if config.Security.JWTSecret == "${JWT_SECRET}" {
		config.Security.JWTSecret = os.Getenv("JWT_SECRET")
	}

	// Redis substitutions
	if config.Redis.Password == "${REDIS_PASSWORD}" {
		config.Redis.Password = os.Getenv("REDIS_PASSWORD")
	}
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if config.Server.Port == 0 {
		return fmt.Errorf("server port is required")
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Conductor.BaseURL == "" {
		return fmt.Errorf("conductor base URL is required")
	}

	return nil
}
