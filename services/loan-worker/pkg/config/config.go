package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config holds the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Conductor  ConductorConfig  `yaml:"conductor"`
	Logging    LoggingConfig    `yaml:"logging"`
	Security   SecurityConfig   `yaml:"security"`
	Application AppConfig       `yaml:"application"`
	I18n       I18nConfig       `yaml:"i18n"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port                    int           `yaml:"port"`
	Host                    string        `yaml:"host"`
	ReadTimeout             int           `yaml:"read_timeout"`
	WriteTimeout            int           `yaml:"write_timeout"`
	GracefulShutdownTimeout int           `yaml:"graceful_shutdown_timeout"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// ConductorConfig holds Netflix Conductor-related configuration
type ConductorConfig struct {
	BaseURL      string `yaml:"base_url"`
	Timeout      int    `yaml:"timeout"`
	RetryAttempts int   `yaml:"retry_attempts"`
	RetryDelay   int    `yaml:"retry_delay"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level        string `yaml:"level"`
	Format       string `yaml:"format"`
	Output       string `yaml:"output"`
	EnableConsole bool  `yaml:"enable_console"`
	EnableFile   bool   `yaml:"enable_file"`
	FilePath     string `yaml:"file_path"`
	MaxSize      int    `yaml:"max_size"`
	MaxAge       int    `yaml:"max_age"`
	MaxBackups   int    `yaml:"max_backups"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret           string   `yaml:"jwt_secret"`
	JWTExpiration       int      `yaml:"jwt_expiration"`
	BcryptCost          int      `yaml:"bcrypt_cost"`
	CORSAllowedOrigins  []string `yaml:"cors_allowed_origins"`
	CORSAllowedMethods  []string `yaml:"cors_allowed_methods"`
	CORSAllowedHeaders  []string `yaml:"cors_allowed_headers"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name                 string  `yaml:"name"`
	Version              string  `yaml:"version"`
	Environment          string  `yaml:"environment"`
	MaxLoanAmount        float64 `yaml:"max_loan_amount"`
	MinLoanAmount        float64 `yaml:"min_loan_amount"`
	MaxDTIRatio          float64 `yaml:"max_dti_ratio"`
	DefaultInterestRate  float64 `yaml:"default_interest_rate"`
	MaxInterestRate      float64 `yaml:"max_interest_rate"`
	MinInterestRate      float64 `yaml:"min_interest_rate"`
	OfferExpirationHours int     `yaml:"offer_expiration_hours"`
}

// I18nConfig holds internationalization configuration
type I18nConfig struct {
	DefaultLanguage     string   `yaml:"default_language"`
	SupportedLanguages  []string `yaml:"supported_languages"`
	FallbackLanguage    string   `yaml:"fallback_language"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Get environment
	env := getEnvironment()
	
	// Load YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var configMap map[string]Config
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
func overrideWithEnvVars(config *Config) {
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

	// Logging configuration
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Logging.Format = format
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
	if maxAmount := os.Getenv("MAX_LOAN_AMOUNT"); maxAmount != "" {
		if amount, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			config.Application.MaxLoanAmount = amount
		}
	}
	if minAmount := os.Getenv("MIN_LOAN_AMOUNT"); minAmount != "" {
		if amount, err := strconv.ParseFloat(minAmount, 64); err == nil {
			config.Application.MinLoanAmount = amount
		}
	}
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
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
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Application.Environment) == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Application.Environment) == "production"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	return strings.ToLower(c.Application.Environment) == "test"
}

// IsDocker returns true if running in docker mode
func (c *Config) IsDocker() bool {
	return strings.ToLower(c.Application.Environment) == "docker"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Database.Port <= 0 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	if c.Application.MaxLoanAmount <= 0 {
		return fmt.Errorf("invalid max loan amount: %f", c.Application.MaxLoanAmount)
	}
	if c.Application.MinLoanAmount <= 0 {
		return fmt.Errorf("invalid min loan amount: %f", c.Application.MinLoanAmount)
	}
	if c.Application.MaxLoanAmount <= c.Application.MinLoanAmount {
		return fmt.Errorf("max loan amount must be greater than min loan amount")
	}
	return nil
}
