package testing

import (
	"github.com/abdussamadbello/echonext"
	"gorm.io/gorm"
)

// Suite provides a base test suite with common setup/teardown
type Suite struct {
	App      *echonext.App
	DB       *gorm.DB
	Client   *APIClient
	Fixtures *FixtureManager
}

// NewSuite creates a new test suite
func NewSuite(app *echonext.App, db *gorm.DB) *Suite {
	client := NewAPIClient(app)
	fixtures := NewFixtureManager(db)

	return &Suite{
		App:      app,
		DB:       db,
		Client:   client,
		Fixtures: fixtures,
	}
}

// Setup is called before each test (override in your suite)
func (s *Suite) Setup() error {
	return nil
}

// Teardown is called after each test (override in your suite)
func (s *Suite) Teardown() error {
	// Clear fixtures by default
	if s.Fixtures != nil {
		return s.Fixtures.Clear()
	}
	return nil
}

// WithAuth returns a client with authentication
func (s *Suite) WithAuth(token string) *APIClient {
	return s.Client.WithAuth(token)
}

// LoadFixtures is a convenience method to load test fixtures
func (s *Suite) LoadFixtures(records ...interface{}) error {
	return s.Fixtures.Load(records...)
}

// IntegrationSuite extends Suite with transaction rollback for each test
type IntegrationSuite struct {
	Suite
	tx *gorm.DB
}

// NewIntegrationSuite creates a new integration test suite with transaction support
func NewIntegrationSuite(app *echonext.App, db *gorm.DB) *IntegrationSuite {
	suite := NewSuite(app, db)
	return &IntegrationSuite{
		Suite: *suite,
	}
}

// BeginTx starts a transaction before each test
func (s *IntegrationSuite) BeginTx() error {
	s.tx = s.DB.Begin()
	if s.tx.Error != nil {
		return s.tx.Error
	}

	// Update fixtures to use transaction
	s.Fixtures = NewFixtureManager(s.tx)

	return nil
}

// RollbackTx rolls back the transaction after each test
func (s *IntegrationSuite) RollbackTx() error {
	if s.tx != nil {
		return s.tx.Rollback().Error
	}
	return nil
}

// Helper methods for common test patterns

// AssertRecordExists checks if a record exists in the database
func (s *Suite) AssertRecordExists(t TestingT, model interface{}, conditions ...interface{}) {
	err := s.DB.Where(conditions[0], conditions[1:]...).First(model).Error
	if err == gorm.ErrRecordNotFound {
		t.Errorf("Expected record to exist but it was not found")
	} else if err != nil {
		t.Errorf("Database error: %v", err)
	}
}

// AssertRecordNotExists checks if a record does not exist in the database
func (s *Suite) AssertRecordNotExists(t TestingT, model interface{}, conditions ...interface{}) {
	err := s.DB.Where(conditions[0], conditions[1:]...).First(model).Error
	if err == nil {
		t.Errorf("Expected record to not exist but it was found")
	} else if err != gorm.ErrRecordNotFound {
		t.Errorf("Database error: %v", err)
	}
}

// AssertRecordCount checks the count of records matching conditions
func (s *Suite) AssertRecordCount(t TestingT, model interface{}, expected int64, conditions ...interface{}) {
	var count int64
	query := s.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	err := query.Count(&count).Error
	if err != nil {
		t.Errorf("Database error: %v", err)
		return
	}

	if count != expected {
		t.Errorf("Expected %d records, got %d", expected, count)
	}
}
