package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// BaseConfig contains common configuration fields for all services
type BaseConfig struct {
	Environment string          `yaml:"environment" json:"environment"`
	Service     ServiceConfig   `yaml:"service" json:"service"`
	Server      ServerConfig    `yaml:"server" json:"server"`
	Database    DatabaseConfig  `yaml:"database" json:"database"`
	Redis       RedisConfig     `yaml:"redis" json:"redis"`
	Logging     LoggingConfig   `yaml:"logging" json:"logging"`
	I18n        I18nConfig      `yaml:"i18n" json:"i18n"`
	Conductor   ConductorConfig `yaml:"conductor" json:"conductor"`
	Security    SecurityConfig  `yaml:"security" json:"security"`
	Application AppConfig       `yaml:"application" json:"application"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	Port    string `yaml:"port" json:"port"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port                    int           `yaml:"port" json:"port"`
	Host                    string        `yaml:"host" json:"host"`
	ReadTimeout             int           `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout            int           `yaml:"write_timeout" json:"write_timeout"`
	GracefulShutdownTimeout int           `yaml:"graceful_shutdown_timeout" json:"graceful_shutdown_timeout"`
	IdleTimeout             time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	CORS                    CORSConfig    `yaml:"cors" json:"cors"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Origins     []string `yaml:"origins" json:"origins"`
	Methods     []string `yaml:"methods" json:"methods"`
	Headers     []string `yaml:"headers" json:"headers"`
	Credentials bool     `yaml:"credentials" json:"credentials"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	User            string        `yaml:"user" json:"user"`
	Password        string        `yaml:"password" json:"password"`
	Name            string        `yaml:"name" json:"name"`
	SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode"`
	URL             string        `yaml:"url" json:"url"` // Keep for backward compatibility
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	AutoMigrate     bool          `yaml:"auto_migrate" json:"auto_migrate"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
	PoolSize int    `yaml:"pool_size" json:"pool_size"`
}

// ConductorConfig holds Netflix Conductor-related configuration
type ConductorConfig struct {
	BaseURL       string `yaml:"base_url" json:"base_url"`
	Timeout       int    `yaml:"timeout" json:"timeout"`
	RetryAttempts int    `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay    int    `yaml:"retry_delay" json:"retry_delay"`
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
	Name                 string  `yaml:"name" json:"name"`
	Version              string  `yaml:"version" json:"version"`
	Environment          string  `yaml:"environment" json:"environment"`
	MaxLoanAmount        float64 `yaml:"max_loan_amount" json:"max_loan_amount"`
	MinLoanAmount        float64 `yaml:"min_loan_amount" json:"min_loan_amount"`
	MaxDTIRatio          float64 `yaml:"max_dti_ratio" json:"max_dti_ratio"`
	DefaultInterestRate  float64 `yaml:"default_interest_rate" json:"default_interest_rate"`
	MaxInterestRate      float64 `yaml:"max_interest_rate" json:"max_interest_rate"`
	MinInterestRate      float64 `yaml:"min_interest_rate" json:"min_interest_rate"`
	OfferExpirationHours int     `yaml:"offer_expiration_hours" json:"offer_expiration_hours"`
}

// LoggingConfig holds logging configuration
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

// I18nConfig holds internationalization configuration
type I18nConfig struct {
	DefaultLanguage    string   `yaml:"default_language" json:"default_language"`
	SupportedLanguages []string `yaml:"supported_languages" json:"supported_languages"`
	FallbackLanguage   string   `yaml:"fallback_language" json:"fallback_language"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*BaseConfig, error) {
	// Get environment
	env := getEnvironment()

	// Load YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var configMap map[string]BaseConfig
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

// overrideWithEnvVars overrides configuration with environment variables
func overrideWithEnvVars(config *BaseConfig) {
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

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(filename string, config interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv(config *BaseConfig) {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.Service.Name = serviceName
	}

	if serviceVersion := os.Getenv("SERVICE_VERSION"); serviceVersion != "" {
		config.Service.Version = serviceVersion
	}

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}

	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Name = dbName
	}

	if dbSSLMode := os.Getenv("DB_SSLMODE"); dbSSLMode != "" {
		config.Database.SSLMode = dbSSLMode
	}

	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}

	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		config.Redis.Port = redisPort
	}

	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}

	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		if db, err := strconv.Atoi(redisDB); err == nil {
			config.Redis.DB = db
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
		config.Server.CORS.Origins = strings.Split(corsOrigins, ",")
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

// GetString gets a string environment variable with a default value
func GetString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetInt gets an integer environment variable with a default value
func GetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetBool gets a boolean environment variable with a default value
func GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetDuration gets a duration environment variable with a default value
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetStringSlice gets a comma-separated string environment variable as a slice
func GetStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// SetDefaults sets default values for common configuration fields
func SetDefaults(config *BaseConfig) {
	if config.Environment == "" {
		config.Environment = "development"
	}

	if config.Service.Name == "" {
		config.Service.Name = "unknown-service"
	}

	if config.Service.Version == "" {
		config.Service.Version = "1.0.0"
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8080
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

	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 60 * time.Second
	}

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
		config.Database.Name = "postgres"
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

	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}

	if config.Redis.Port == "" {
		config.Redis.Port = "6379"
	}

	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 10
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}

	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	if config.I18n.DefaultLanguage == "" {
		config.I18n.DefaultLanguage = "en"
	}

	if len(config.I18n.SupportedLanguages) == 0 {
		config.I18n.SupportedLanguages = []string{"en", "vi"}
	}

	if config.I18n.FallbackLanguage == "" {
		config.I18n.FallbackLanguage = "en"
	}

	if len(config.Server.CORS.Origins) == 0 {
		config.Server.CORS.Origins = []string{"*"}
	}

	if len(config.Server.CORS.Methods) == 0 {
		config.Server.CORS.Methods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	if len(config.Server.CORS.Headers) == 0 {
		config.Server.CORS.Headers = []string{"*"}
	}

	// Set conductor defaults
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

	// Set security defaults
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

	// Set application defaults
	if config.Application.Name == "" {
		config.Application.Name = "loan-service"
	}

	if config.Application.Version == "" {
		config.Application.Version = "v1.0.0"
	}

	if config.Application.Environment == "" {
		config.Application.Environment = "development"
	}

	if config.Application.MaxLoanAmount == 0 {
		config.Application.MaxLoanAmount = 50000
	}

	if config.Application.MinLoanAmount == 0 {
		config.Application.MinLoanAmount = 5000
	}

	if config.Application.MaxDTIRatio == 0 {
		config.Application.MaxDTIRatio = 0.43
	}

	if config.Application.DefaultInterestRate == 0 {
		config.Application.DefaultInterestRate = 8.5
	}

	if config.Application.MaxInterestRate == 0 {
		config.Application.MaxInterestRate = 15.0
	}

	if config.Application.MinInterestRate == 0 {
		config.Application.MinInterestRate = 5.0
	}

	if config.Application.OfferExpirationHours == 0 {
		config.Application.OfferExpirationHours = 168 // 7 days
	}
}

// GetDSN returns the database connection string
func (c *BaseConfig) GetDSN() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
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
func (c *BaseConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if running in development mode
func (c *BaseConfig) IsDevelopment() bool {
	return strings.ToLower(c.Application.Environment) == "development"
}

// IsProduction returns true if running in production mode
func (c *BaseConfig) IsProduction() bool {
	return strings.ToLower(c.Application.Environment) == "production"
}

// IsTest returns true if running in test mode
func (c *BaseConfig) IsTest() bool {
	return strings.ToLower(c.Application.Environment) == "test"
}

// IsDocker returns true if running in docker mode
func (c *BaseConfig) IsDocker() bool {
	return strings.ToLower(c.Application.Environment) == "docker"
}

// Validate validates the configuration
func (c *BaseConfig) Validate() error {
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
