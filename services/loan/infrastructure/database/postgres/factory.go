package postgres

import (
	"go.uber.org/zap"

	"loan-service/application"
)

// Factory manages database repositories
type Factory struct {
	connection *Connection
	logger     *zap.Logger
}

// NewFactory creates a new database factory
func NewFactory(connection *Connection, logger *zap.Logger) *Factory {
	return &Factory{
		connection: connection,
		logger:     logger,
	}
}

// GetUserRepository returns a new UserRepository instance
func (f *Factory) GetUserRepository() application.UserRepository {
	return NewUserRepository(f.connection, f.logger)
}

// GetLoanRepository returns a new LoanRepository instance
func (f *Factory) GetLoanRepository() application.LoanRepository {
	return NewLoanRepository(f.connection, f.logger)
}

// GetConnection returns the database connection
func (f *Factory) GetConnection() *Connection {
	return f.connection
}

// Close closes the database connection
func (f *Factory) Close() error {
	return f.connection.Close()
}
