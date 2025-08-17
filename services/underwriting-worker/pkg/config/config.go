package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Application ApplicationConfig `yaml:"application"`
	Database    DatabaseConfig    `yaml:"database"`
	Conductor   ConductorConfig   `yaml:"conductor"`
	Services    ServicesConfig    `yaml:"services"`
	Logging     LoggingConfig     `yaml:"logging"`
	Security    SecurityConfig    `yaml:"security"`
}

// ApplicationConfig represents application-specific configuration
type ApplicationConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Port        int    `yaml:"port"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int    `yaml:"max_connections"`
	MinConns int    `yaml:"min_connections"`
}

// ConductorConfig represents Conductor workflow engine configuration
type ConductorConfig struct {
	ServerURL       string `yaml:"server_url"`
	WorkerPoolSize  int    `yaml:"worker_pool_size"`
	PollingInterval int    `yaml:"polling_interval_ms"`
	UpdateRetryTime int    `yaml:"update_retry_time_ms"`
}

// ServicesConfig represents external services configuration
type ServicesConfig struct {
	CreditBureau       CreditBureauConfig       `yaml:"credit_bureau"`
	RiskScoring        RiskScoringConfig        `yaml:"risk_scoring"`
	IncomeVerification IncomeVerificationConfig `yaml:"income_verification"`
	DecisionEngine     DecisionEngineConfig     `yaml:"decision_engine"`
	Notification       NotificationConfig       `yaml:"notification"`
}

// CreditBureauConfig represents credit bureau service configuration
type CreditBureauConfig struct {
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Timeout  int    `yaml:"timeout_seconds"`
}

// RiskScoringConfig represents risk scoring service configuration
type RiskScoringConfig struct {
	Provider     string `yaml:"provider"`
	BaseURL      string `yaml:"base_url"`
	APIKey       string `yaml:"api_key"`
	ModelVersion string `yaml:"model_version"`
	Timeout      int    `yaml:"timeout_seconds"`
}

// IncomeVerificationConfig represents income verification service configuration
type IncomeVerificationConfig struct {
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Timeout  int    `yaml:"timeout_seconds"`
}

// DecisionEngineConfig represents decision engine service configuration
type DecisionEngineConfig struct {
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Timeout  int    `yaml:"timeout_seconds"`
}

// NotificationConfig represents notification service configuration
type NotificationConfig struct {
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Timeout  int    `yaml:"timeout_seconds"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EncryptionKey string `yaml:"encryption_key"`
	JWTSecret     string `yaml:"jwt_secret"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Get absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", absPath)
	}

	// Read file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	config.overrideWithEnvVars()

	return &config, nil
}

// overrideWithEnvVars overrides configuration with environment variables
func (c *Config) overrideWithEnvVars() {
	// Database environment variables
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		c.Database.Username = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		c.Database.Password = dbPass
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		c.Database.Database = dbName
	}

	// Conductor environment variables
	if conductorURL := os.Getenv("CONDUCTOR_SERVER_URL"); conductorURL != "" {
		c.Conductor.ServerURL = conductorURL
	}

	// Service API keys
	if creditAPIKey := os.Getenv("CREDIT_BUREAU_API_KEY"); creditAPIKey != "" {
		c.Services.CreditBureau.APIKey = creditAPIKey
	}
	if riskAPIKey := os.Getenv("RISK_SCORING_API_KEY"); riskAPIKey != "" {
		c.Services.RiskScoring.APIKey = riskAPIKey
	}

	// Security environment variables
	if encKey := os.Getenv("ENCRYPTION_KEY"); encKey != "" {
		c.Security.EncryptionKey = encKey
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		c.Security.JWTSecret = jwtSecret
	}

	// Logging environment variables
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.Logging.Level = logLevel
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate application config
	if c.Application.Name == "" {
		return fmt.Errorf("application name is required")
	}
	if c.Application.Version == "" {
		return fmt.Errorf("application version is required")
	}
	if c.Application.Environment == "" {
		return fmt.Errorf("application environment is required")
	}

	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port == 0 {
		return fmt.Errorf("database port is required")
	}
	if c.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate Conductor config
	if c.Conductor.ServerURL == "" {
		return fmt.Errorf("conductor server URL is required")
	}
	if c.Conductor.WorkerPoolSize <= 0 {
		c.Conductor.WorkerPoolSize = 5 // Default value
	}
	if c.Conductor.PollingInterval <= 0 {
		c.Conductor.PollingInterval = 1000 // Default 1 second
	}

	// Validate logging config
	if c.Logging.Level == "" {
		c.Logging.Level = "info" // Default level
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json" // Default format
	}

	return nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Application.Environment == "development" || c.Application.Environment == "dev"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Application.Environment == "production" || c.Application.Environment == "prod"
}

// GetDatabaseConnectionString returns the database connection string
func (c *Config) GetDatabaseConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}
