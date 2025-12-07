// Package testing provides optional testing utilities for EchoNext applications.
//
// This is a contrib package and is completely optional. You can use standard testing
// if you prefer, or use these helpers to reduce boilerplate code in your tests.
//
// Features:
//   - APIClient for testing HTTP endpoints
//   - FixtureManager for managing test data
//   - Factory pattern for creating test entities
//   - Suite base class with setup/teardown
//   - IntegrationSuite with transaction rollback
//
// Example usage:
//
//	import (
//	    "testing"
//	    "github.com/abdussamadbello/echonext"
//	    echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"
//	)
//
//	func TestUserAPI(t *testing.T) {
//	    app := echonext.New()
//	    // Register your routes...
//
//	    client := echonexttest.NewAPIClient(app)
//
//	    // Test GET request
//	    resp := client.GET("/users/1")
//	    resp.AssertStatus(t, 200)
//
//	    var user User
//	    resp.JSON(&user)
//	    if user.ID != 1 {
//	        t.Errorf("Expected user ID 1, got %d", user.ID)
//	    }
//
//	    // Test POST request
//	    newUser := CreateUserRequest{Name: "John", Email: "john@example.com"}
//	    resp = client.POST("/users", newUser)
//	    resp.AssertStatus(t, 201)
//	}
//
// Using fixtures:
//
//	func TestWithFixtures(t *testing.T) {
//	    fixtures := echonexttest.NewFixtureManager(db)
//	    defer fixtures.Clear()
//
//	    // Load test data
//	    fixtures.Load(
//	        &User{ID: 1, Name: "Alice"},
//	        &User{ID: 2, Name: "Bob"},
//	    )
//
//	    // Run tests...
//	}
//
// Using test suite:
//
//	type UserTestSuite struct {
//	    echonexttest.Suite
//	}
//
//	func TestUserSuite(t *testing.T) {
//	    suite := echonexttest.NewSuite(app, db)
//	    // Run your tests with suite...
//	}
package testing
