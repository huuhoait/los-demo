package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Host            string `yaml:"host" json:"host"`
	Port            int    `yaml:"port" json:"port"`
	User            string `yaml:"user" json:"user"`
	Password        string `yaml:"password" json:"password"`
	Database        string `yaml:"database" json:"database"`
	SSLMode         string `yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"` // minutes
	LogLevel        string `yaml:"log_level" json:"log_level"`
}

// Connection wraps gorm.DB with additional functionality
type Connection struct {
	*gorm.DB
	config Config
}

// NewConnection creates a new database connection
func NewConnection(config Config) (*Connection, error) {
	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	// Configure GORM logger
	var logLevel logger.LogLevel
	switch config.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Minute)
	}

	return &Connection{
		DB:     db,
		config: config,
	}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping pings the database
func (c *Connection) Ping() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Health returns database health information
func (c *Connection) Health() map[string]interface{} {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"status":              "healthy",
		"open_connections":    stats.OpenConnections,
		"idle_connections":    stats.Idle,
		"in_use":              stats.InUse,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// Transaction executes a function within a database transaction
func (c *Connection) Transaction(fn func(*gorm.DB) error) error {
	return c.DB.Transaction(fn)
}

// WithContext returns a new connection with context
func (c *Connection) WithContext(ctx context.Context) *Connection {
	return &Connection{
		DB:     c.DB.WithContext(ctx),
		config: c.config,
	}
}

// Repository provides base repository functionality
type Repository struct {
	db *Connection
}

// NewRepository creates a new repository
func NewRepository(db *Connection) *Repository {
	return &Repository{db: db}
}

// DB returns the database connection
func (r *Repository) DB() *gorm.DB {
	return r.db.DB
}

// WithTx executes function within a transaction
func (r *Repository) WithTx(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create creates a new record
func (r *Repository) Create(ctx context.Context, entity interface{}) error {
	return r.db.DB.WithContext(ctx).Create(entity).Error
}

// Update updates a record
func (r *Repository) Update(ctx context.Context, entity interface{}) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// Delete soft deletes a record
func (r *Repository) Delete(ctx context.Context, entity interface{}) error {
	return r.db.DB.WithContext(ctx).Delete(entity).Error
}

// FindByID finds a record by ID
func (r *Repository) FindByID(ctx context.Context, id interface{}, entity interface{}) error {
	return r.db.DB.WithContext(ctx).First(entity, id).Error
}

// FindAll finds all records with optional conditions
func (r *Repository) FindAll(ctx context.Context, entities interface{}, conditions ...interface{}) error {
	query := r.db.DB.WithContext(ctx)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	return query.Find(entities).Error
}

// FindWithPagination finds records with pagination
func (r *Repository) FindWithPagination(ctx context.Context, entities interface{}, offset, limit int, conditions ...interface{}) (int64, error) {
	query := r.db.DB.WithContext(ctx)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	var total int64
	if err := query.Model(entities).Count(&total).Error; err != nil {
		return 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(entities).Error
	return total, err
}

// Exists checks if a record exists
func (r *Repository) Exists(ctx context.Context, entity interface{}, conditions ...interface{}) (bool, error) {
	query := r.db.DB.WithContext(ctx).Model(entity)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate sets created_at and updated_at
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UTC()
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

// BeforeUpdate sets updated_at
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now().UTC()
	return nil
}

// Paginator handles pagination logic
type Paginator struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
	Pages   int   `json:"pages"`
	Offset  int   `json:"-"`
}

// NewPaginator creates a new paginator
func NewPaginator(page, perPage int) *Paginator {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	return &Paginator{
		Page:    page,
		PerPage: perPage,
		Offset:  offset,
	}
}

// SetTotal sets the total count and calculates pages
func (p *Paginator) SetTotal(total int64) {
	p.Total = total
	if total > 0 {
		p.Pages = int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	} else {
		p.Pages = 0
	}
}

// Query builder helpers

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	db *gorm.DB
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{db: db}
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(query interface{}, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Where(query, args...)
	return qb
}

// Or adds an OR condition
func (qb *QueryBuilder) Or(query interface{}, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Or(query, args...)
	return qb
}

// Order adds an ORDER BY clause
func (qb *QueryBuilder) Order(value interface{}) *QueryBuilder {
	qb.db = qb.db.Order(value)
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.db = qb.db.Limit(limit)
	return qb
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.db = qb.db.Offset(offset)
	return qb
}

// Preload adds a preload
func (qb *QueryBuilder) Preload(query string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Preload(query, args...)
	return qb
}

// Joins adds a JOIN clause
func (qb *QueryBuilder) Joins(query string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Joins(query, args...)
	return qb
}

// Find executes the query and returns results
func (qb *QueryBuilder) Find(dest interface{}) error {
	return qb.db.Find(dest).Error
}

// First executes the query and returns the first result
func (qb *QueryBuilder) First(dest interface{}) error {
	return qb.db.First(dest).Error
}

// Count returns the count of records
func (qb *QueryBuilder) Count(count *int64) error {
	return qb.db.Count(count).Error
}

// DB returns the underlying GORM DB
func (qb *QueryBuilder) DB() *gorm.DB {
	return qb.db
}
