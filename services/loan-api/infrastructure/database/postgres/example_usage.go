package postgres

import (
	"context"
	"log"
	"time"

	"go.uber.org/zap"

	"loan-api/domain"
)

// ExampleUsage demonstrates how to use the database repositories
func ExampleUsage() {
	// 1. Create logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 2. Create database configuration
	config := &Config{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "password",
		Database:        "loan_service",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * 60, // 5 minutes
	}

	// 3. Create database connection
	connection, err := NewConnection(config, logger)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer connection.Close()

	// 4. Create repository factory
	factory := NewFactory(connection, logger)

	// 5. Get repositories
	userRepo := factory.GetUserRepository()
	loanRepo := factory.GetLoanRepository()

	// 6. Use repositories
	ctx := context.Background()

	// Create a user
	user := &domain.User{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@example.com",
		PhoneNumber: "+1234567890",
		DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		SSN:         "123456789",
		Address: domain.Address{
			StreetAddress: "123 Main St",
			City:          "New York",
			State:         "NY",
			ZipCode:       "10001",
			Country:       "USA",
			ResidenceType: domain.ResidenceOwn,
			TimeAtAddress: 24,
		},
		EmploymentInfo: domain.EmploymentInfo{
			EmployerName: "ABC Company",
			JobTitle:     "Software Engineer",
			TimeEmployed: 36,
			WorkPhone:    "+1234567890",
			WorkEmail:    "john.doe@abccompany.com",
		},
		BankingInfo: domain.BankingInfo{
			BankName:      "Chase Bank",
			AccountType:   domain.AccountChecking,
			AccountNumber: "1234567890",
			RoutingNumber: "021000021",
		},
	}

	userID, err := userRepo.CreateUser(ctx, user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}
	log.Printf("Created user with ID: %s", userID)

	// Retrieve the user
	retrievedUser, err := userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return
	}
	log.Printf("Retrieved user: %s %s", retrievedUser.FirstName, retrievedUser.LastName)

	// Create a loan application
	application := &domain.LoanApplication{
		ID:                "app-123",
		UserID:            userID,
		ApplicationNumber: "LOAN123456",
		LoanAmount:        25000,
		LoanPurpose:       domain.PurposeDebtConsolidation,
		RequestedTerm:     60,
		AnnualIncome:      75000,
		MonthlyIncome:     6250,
		EmploymentStatus:  domain.EmploymentFullTime,
		MonthlyDebt:       1500,
		CurrentState:      domain.StateInitiated,
		Status:            domain.StatusDraft,
	}

	err = loanRepo.CreateApplication(ctx, application)
	if err != nil {
		log.Printf("Failed to create application: %v", err)
		return
	}
	log.Printf("Created loan application: %s", application.ID)

	// Retrieve the application
	retrievedApp, err := loanRepo.GetApplicationByID(ctx, application.ID)
	if err != nil {
		log.Printf("Failed to get application: %v", err)
		return
	}
	log.Printf("Retrieved application: %s", retrievedApp.ApplicationNumber)

	// Create a state transition
	fromState := domain.StateInitiated
	transition := &domain.StateTransition{
		ID:               "trans-123",
		ApplicationID:    application.ID,
		FromState:        &fromState,
		ToState:          domain.StatePreQualified,
		TransitionReason: "Application validated and pre-qualified",
		Automated:        true,
		UserID:           &userID,
	}

	err = loanRepo.CreateStateTransition(ctx, transition)
	if err != nil {
		log.Printf("Failed to create state transition: %v", err)
		return
	}
	log.Printf("Created state transition: %s -> %s", *transition.FromState, transition.ToState)

	// Get all state transitions for the application
	transitions, err := loanRepo.GetStateTransitions(ctx, application.ID)
	if err != nil {
		log.Printf("Failed to get state transitions: %v", err)
		return
	}
	log.Printf("Found %d state transitions", len(transitions))

	// Get all applications for the user
	applications, err := loanRepo.GetApplicationsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Failed to get applications: %v", err)
		return
	}
	log.Printf("Found %d applications for user", len(applications))

	// Update application status
	application.CurrentState = domain.StatePreQualified
	application.Status = domain.StatusSubmitted
	err = loanRepo.UpdateApplication(ctx, application)
	if err != nil {
		log.Printf("Failed to update application: %v", err)
		return
	}
	log.Printf("Updated application status to: %s", application.CurrentState)

	// Clean up (optional)
	err = loanRepo.DeleteApplication(ctx, application.ID)
	if err != nil {
		log.Printf("Failed to delete application: %v", err)
	}

	err = userRepo.DeleteUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
	}

	log.Println("Example completed successfully!")
}
