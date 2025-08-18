package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// UnderwritingConfig holds the underwriting service configuration
type UnderwritingConfig struct {
	Server      ServerConfig    `yaml:"server" json:"server"`
	Database    DatabaseConfig  `yaml:"database" json:"database"`
	Conductor   ConductorConfig `yaml:"conductor" json:"conductor"`
	Logging     LoggingConfig   `yaml:"logging" json:"logging"`
	Security    SecurityConfig  `yaml:"security" json:"security"`
	Application AppConfig       `yaml:"application" json:"application"`
	I18n        I18nConfig      `yaml:"i18n" json:"i18n"`
	Services    ServicesConfig  `yaml:"services" json:"services"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port                    int    `yaml:"port" json:"port"`
	Host                    string `yaml:"host" json:"host"`
	ReadTimeout             int    `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout            int    `yaml:"write_timeout" json:"write_timeout"`
	GracefulShutdownTimeout int    `yaml:"graceful_shutdown_timeout" json:"graceful_shutdown_timeout"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	User            string        `yaml:"user" json:"user"`
	Password        string        `yaml:"password" json:"password"`
	Name            string        `yaml:"name" json:"name"`
	SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// ConductorConfig holds Netflix Conductor-related configuration
type ConductorConfig struct {
	BaseURL         string `yaml:"base_url" json:"base_url"`
	Timeout         int    `yaml:"timeout" json:"timeout"`
	RetryAttempts   int    `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay      int    `yaml:"retry_delay" json:"retry_delay"`
	WorkerPoolSize  int    `yaml:"worker_pool_size" json:"worker_pool_size"`
	PollingInterval int    `yaml:"polling_interval_ms" json:"polling_interval_ms"`
	UpdateRetryTime int    `yaml:"update_retry_time_ms" json:"update_retry_time_ms"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level         string `yaml:"level" json:"level"`
	Format        string `yaml:"format" json:"format"`
	Output        string `yaml:"output" json:"output"`
	EnableConsole bool   `yaml:"enable_console" json:"enable_console"`
	EnableFile    bool   `yaml:"enable_file" json:"enable_file"`
	FilePath      string `yaml:"file_path" json:"file_path"`
	MaxSize       int    `yaml:"max_size" json:"max_size"`
	MaxAge        int    `yaml:"max_age" json:"max_age"`
	MaxBackups    int    `yaml:"max_backups" json:"max_backups"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret          string   `yaml:"jwt_secret" json:"jwt_secret"`
	JWTExpiration      int      `yaml:"jwt_expiration" json:"jwt_expiration"`
	BcryptCost         int      `yaml:"bcrypt_cost" json:"bcrypt_cost"`
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins" json:"cors_allowed_origins"`
	CORSAllowedMethods []string `yaml:"cors_allowed_methods" json:"cors_allowed_methods"`
	CORSAllowedHeaders []string `yaml:"cors_allowed_headers" json:"cors_allowed_headers"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
}

// I18nConfig holds internationalization configuration
type I18nConfig struct {
	DefaultLanguage    string   `yaml:"default_language" json:"default_language"`
	SupportedLanguages []string `yaml:"supported_languages" json:"supported_languages"`
	FallbackLanguage   string   `yaml:"fallback_language" json:"fallback_language"`
}

// ServicesConfig represents external services configuration for underwriting
type ServicesConfig struct {
	CreditBureau       CreditBureauConfig       `yaml:"credit_bureau" json:"credit_bureau"`
	RiskScoring        RiskScoringConfig        `yaml:"risk_scoring" json:"risk_scoring"`
	IncomeVerification IncomeVerificationConfig `yaml:"income_verification" json:"income_verification"`
	DecisionEngine     DecisionEngineConfig     `yaml:"decision_engine" json:"decision_engine"`
	Notification       NotificationConfig       `yaml:"notification" json:"notification"`
}

// CreditBureauConfig represents credit bureau service configuration
type CreditBureauConfig struct {
	Provider string `yaml:"provider" json:"provider"`
	BaseURL  string `yaml:"base_url" json:"base_url"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	Timeout  int    `yaml:"timeout_seconds" json:"timeout_seconds"`
}

// RiskScoringConfig represents risk scoring service configuration
type RiskScoringConfig struct {
	Provider     string `yaml:"provider" json:"provider"`
	BaseURL      string `yaml:"base_url" json:"base_url"`
	APIKey       string `yaml:"api_key" json:"api_key"`
	ModelVersion string `yaml:"model_version" json:"model_version"`
	Timeout      int    `yaml:"timeout_seconds" json:"timeout_seconds"`
}

// IncomeVerificationConfig represents income verification service configuration
type IncomeVerificationConfig struct {
	Provider string `yaml:"provider" json:"provider"`
	BaseURL  string `yaml:"base_url" json:"base_url"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	Timeout  int    `yaml:"timeout_seconds" json:"timeout_seconds"`
}

// DecisionEngineConfig represents decision engine service configuration
type DecisionEngineConfig struct {
	Provider string `yaml:"provider" json:"provider"`
	BaseURL  string `yaml:"base_url" json:"base_url"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	Timeout  int    `yaml:"timeout_seconds" json:"timeout_seconds"`
}

// NotificationConfig represents notification service configuration
type NotificationConfig struct {
	Provider string `yaml:"provider" json:"provider"`
	BaseURL  string `yaml:"base_url" json:"base_url"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	Timeout  int    `yaml:"timeout_seconds" json:"timeout_seconds"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*UnderwritingConfig, error) {
	// Get environment
	env := getEnvironment()

	// Load YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var configMap map[string]UnderwritingConfig
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Get environment-specific config
	config, exists := configMap[env]
	if !exists {
		// Fallback to default
		config, exists = configMap["default"]
		if !exists {
			return nil, fmt.Errorf("no configuration found for environment '%s' and no default config", env)
		}
	}

	// Override with environment variables
	overrideWithEnvVars(&config)

	// Set defaults
	SetDefaults(&config)

	return &config, nil
}

// getEnvironment returns the current environment
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}
	return env
}

// overrideWithEnvVars overrides configuration with environment variables
func overrideWithEnvVars(config *UnderwritingConfig) {
	// Server configuration
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}
	if timeout := os.Getenv("READ_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Server.ReadTimeout = t
		}
	}
	if timeout := os.Getenv("WRITE_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Server.WriteTimeout = t
		}
	}

	// Database configuration
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		config.Database.Name = name
	}
	if sslMode := os.Getenv("DB_SSLMODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// Conductor configuration
	if baseURL := os.Getenv("CONDUCTOR_BASE_URL"); baseURL != "" {
		config.Conductor.BaseURL = baseURL
	}
	if timeout := os.Getenv("CONDUCTOR_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Conductor.Timeout = t
		}
	}
	if poolSize := os.Getenv("CONDUCTOR_WORKER_POOL_SIZE"); poolSize != "" {
		if ps, err := strconv.Atoi(poolSize); err == nil {
			config.Conductor.WorkerPoolSize = ps
		}
	}
	if pollingInterval := os.Getenv("CONDUCTOR_POLLING_INTERVAL"); pollingInterval != "" {
		if pi, err := strconv.Atoi(pollingInterval); err == nil {
			config.Conductor.PollingInterval = pi
		}
	}

	// Security configuration
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Security.JWTSecret = jwtSecret
	}
	if jwtExp := os.Getenv("JWT_EXPIRATION"); jwtExp != "" {
		if exp, err := strconv.Atoi(jwtExp); err == nil {
			config.Security.JWTExpiration = exp
		}
	}

	// Application configuration
	if env := os.Getenv("APP_ENV"); env != "" {
		config.Application.Environment = env
	}

	// Services configuration
	if creditAPIKey := os.Getenv("CREDIT_BUREAU_API_KEY"); creditAPIKey != "" {
		config.Services.CreditBureau.APIKey = creditAPIKey
	}
	if creditURL := os.Getenv("CREDIT_BUREAU_BASE_URL"); creditURL != "" {
		config.Services.CreditBureau.BaseURL = creditURL
	}
	if riskAPIKey := os.Getenv("RISK_SCORING_API_KEY"); riskAPIKey != "" {
		config.Services.RiskScoring.APIKey = riskAPIKey
	}
	if riskURL := os.Getenv("RISK_SCORING_BASE_URL"); riskURL != "" {
		config.Services.RiskScoring.BaseURL = riskURL
	}
	if incomeAPIKey := os.Getenv("INCOME_VERIFICATION_API_KEY"); incomeAPIKey != "" {
		config.Services.IncomeVerification.APIKey = incomeAPIKey
	}
	if incomeURL := os.Getenv("INCOME_VERIFICATION_BASE_URL"); incomeURL != "" {
		config.Services.IncomeVerification.BaseURL = incomeURL
	}
	if decisionAPIKey := os.Getenv("DECISION_ENGINE_API_KEY"); decisionAPIKey != "" {
		config.Services.DecisionEngine.APIKey = decisionAPIKey
	}
	if decisionURL := os.Getenv("DECISION_ENGINE_BASE_URL"); decisionURL != "" {
		config.Services.DecisionEngine.BaseURL = decisionURL
	}
	if notificationAPIKey := os.Getenv("NOTIFICATION_API_KEY"); notificationAPIKey != "" {
		config.Services.Notification.APIKey = notificationAPIKey
	}
	if notificationURL := os.Getenv("NOTIFICATION_BASE_URL"); notificationURL != "" {
		config.Services.Notification.BaseURL = notificationURL
	}
}

// SetDefaults sets default values for configuration fields
func SetDefaults(config *UnderwritingConfig) {
	// Server defaults
	if config.Server.Port == 0 {
		config.Server.Port = 8081
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30
	}
	if config.Server.GracefulShutdownTimeout == 0 {
		config.Server.GracefulShutdownTimeout = 30
	}

	// Database defaults
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.User == "" {
		config.Database.User = "postgres"
	}
	if config.Database.Name == "" {
		config.Database.Name = "underwriting_service"
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 10
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 5 * time.Minute
	}

	// Conductor defaults
	if config.Conductor.BaseURL == "" {
		config.Conductor.BaseURL = "http://localhost:8082"
	}
	if config.Conductor.Timeout == 0 {
		config.Conductor.Timeout = 30
	}
	if config.Conductor.RetryAttempts == 0 {
		config.Conductor.RetryAttempts = 3
	}
	if config.Conductor.RetryDelay == 0 {
		config.Conductor.RetryDelay = 1000
	}
	if config.Conductor.WorkerPoolSize == 0 {
		config.Conductor.WorkerPoolSize = 5
	}
	if config.Conductor.PollingInterval == 0 {
		config.Conductor.PollingInterval = 1000
	}
	if config.Conductor.UpdateRetryTime == 0 {
		config.Conductor.UpdateRetryTime = 1000
	}

	// Security defaults
	if config.Security.JWTSecret == "" {
		config.Security.JWTSecret = "your-secret-key-change-in-production"
	}
	if config.Security.JWTExpiration == 0 {
		config.Security.JWTExpiration = 3600
	}
	if config.Security.BcryptCost == 0 {
		config.Security.BcryptCost = 12
	}
	if len(config.Security.CORSAllowedOrigins) == 0 {
		config.Security.CORSAllowedOrigins = []string{"*"}
	}
	if len(config.Security.CORSAllowedMethods) == 0 {
		config.Security.CORSAllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(config.Security.CORSAllowedHeaders) == 0 {
		config.Security.CORSAllowedHeaders = []string{"Content-Type", "Authorization", "X-Request-ID", "X-Language"}
	}

	// Application defaults
	if config.Application.Name == "" {
		config.Application.Name = "underwriting-worker"
	}
	if config.Application.Version == "" {
		config.Application.Version = "v1.0.0"
	}
	if config.Application.Environment == "" {
		config.Application.Environment = "development"
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	// I18n defaults
	if config.I18n.DefaultLanguage == "" {
		config.I18n.DefaultLanguage = "en"
	}
	if len(config.I18n.SupportedLanguages) == 0 {
		config.I18n.SupportedLanguages = []string{"en", "vi"}
	}
	if config.I18n.FallbackLanguage == "" {
		config.I18n.FallbackLanguage = "en"
	}

	// Services defaults
	if config.Services.CreditBureau.Provider == "" {
		config.Services.CreditBureau.Provider = "experian"
	}
	if config.Services.CreditBureau.Timeout == 0 {
		config.Services.CreditBureau.Timeout = 30
	}
	if config.Services.RiskScoring.Provider == "" {
		config.Services.RiskScoring.Provider = "fico"
	}
	if config.Services.RiskScoring.Timeout == 0 {
		config.Services.RiskScoring.Timeout = 30
	}
	if config.Services.IncomeVerification.Provider == "" {
		config.Services.IncomeVerification.Provider = "plaid"
	}
	if config.Services.IncomeVerification.Timeout == 0 {
		config.Services.IncomeVerification.Timeout = 30
	}
	if config.Services.DecisionEngine.Provider == "" {
		config.Services.DecisionEngine.Provider = "internal"
	}
	if config.Services.DecisionEngine.Timeout == 0 {
		config.Services.DecisionEngine.Timeout = 30
	}
	if config.Services.Notification.Provider == "" {
		config.Services.Notification.Provider = "sendgrid"
	}
	if config.Services.Notification.Timeout == 0 {
		config.Services.Notification.Timeout = 30
	}
}

// GetDSN returns the database connection string
func (c *UnderwritingConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetServerAddr returns the server address string
func (c *UnderwritingConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if running in development mode
func (c *UnderwritingConfig) IsDevelopment() bool {
	return strings.ToLower(c.Application.Environment) == "development"
}

// IsProduction returns true if running in production mode
func (c *UnderwritingConfig) IsProduction() bool {
	return strings.ToLower(c.Application.Environment) == "production"
}

// IsTest returns true if running in test mode
func (c *UnderwritingConfig) IsTest() bool {
	return strings.ToLower(c.Application.Environment) == "test"
}

// IsDocker returns true if running in docker mode
func (c *UnderwritingConfig) IsDocker() bool {
	return strings.ToLower(c.Application.Environment) == "docker"
}

// Validate validates the configuration
func (c *UnderwritingConfig) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Database.Port <= 0 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	return nil
}
