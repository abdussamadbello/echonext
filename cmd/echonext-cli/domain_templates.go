package main

import (
	"fmt"
	"log"
)

// generateUserModel creates a sample User model
func (g *ProjectGenerator) generateUserModel() string {
	content, err := g.templateGenerator.GenerateUserModel()
	if err != nil {
		log.Fatalf("Failed to generate user model: %v", err)
	}
	return content
}

// generateUserService creates a sample User service
func (g *ProjectGenerator) generateUserService() string {
	return fmt.Sprintf(`package user

import (
	"context"
	"fmt"
)

// Service handles user business logic
type Service interface {
	Create(ctx context.Context, req CreateUserRequest) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id uint, req UpdateUserRequest) (*User, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error)
}

// service implements Service interface
type service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// Create creates a new user
func (s *service) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user with email %%s already exists", req.Email)
	}

	user := &User{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %%w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *service) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID %%d: %%w", id, err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email %%s: %%w", email, err)
	}

	return user, nil
}

// Update updates an existing user
func (s *service) Update(ctx context.Context, id uint, req UpdateUserRequest) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %%w", err)
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// Check if new email already exists
		existing, _ := s.repo.GetByEmail(ctx, req.Email)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("email %%s is already taken", req.Email)
		}
		user.Email = req.Email
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %%w", err)
	}

	return user, nil
}

// Delete soft-deletes a user
func (s *service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %%w", err)
	}

	return nil
}

// List retrieves a paginated list of users
func (s *service) List(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	users, total, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %%w", err)
	}

	// Calculate pagination
	totalPages := (total + req.Limit - 1) / req.Limit

	return &ListUsersResponse{
		Users:      toUserResponses(users),
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

// Helper function to convert users to response DTOs
func toUserResponses(users []*User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}
	return responses
}

// Repository interface for user data access
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req ListUsersRequest) ([]*User, int, error)
}

// repository implements Repository interface using GORM
type repository struct {
	db interface{} // Replace with actual GORM DB type
}

// NewRepository creates a new user repository
func NewRepository(db interface{}) Repository {
	return &repository{db: db}
}

// Create creates a new user in the database
func (r *repository) Create(ctx context.Context, user *User) error {
	// TODO: Implement with actual GORM DB
	// return r.db.WithContext(ctx).Create(user).Error
	return nil
}

// GetByID retrieves a user by ID
func (r *repository) GetByID(ctx context.Context, id uint) (*User, error) {
	// TODO: Implement with actual GORM DB
	// var user User
	// err := r.db.WithContext(ctx).First(&user, id).Error
	// return &user, err
	return nil, nil
}

// GetByEmail retrieves a user by email
func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	// TODO: Implement with actual GORM DB
	// var user User
	// err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	// return &user, err
	return nil, nil
}

// Update updates a user in the database
func (r *repository) Update(ctx context.Context, user *User) error {
	// TODO: Implement with actual GORM DB
	// return r.db.WithContext(ctx).Save(user).Error
	return nil
}

// Delete soft-deletes a user from the database
func (r *repository) Delete(ctx context.Context, id uint) error {
	// TODO: Implement with actual GORM DB
	// return r.db.WithContext(ctx).Delete(&User{}, id).Error
	return nil
}

// List retrieves a paginated list of users
func (r *repository) List(ctx context.Context, req ListUsersRequest) ([]*User, int, error) {
	// TODO: Implement with actual GORM DB
	/*
	var users []*User
	var total int64

	query := r.db.WithContext(ctx).Model(&User{})

	// Apply search filter if provided
	if req.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%%"+req.Search+"%%", "%%"+req.Search+"%%")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
	*/
	return nil, 0, nil
}
`)
}

// generateUserHandler creates the User HTTP handler
func (g *ProjectGenerator) generateUserHandler() string {
	return fmt.Sprintf(`package user

import (
	"strconv"

	"github.com/abdussamadbello/echonext"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for users
type Handler struct {
	service Service
}

// NewHandler creates a new user handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all user routes
func (h *Handler) RegisterRoutes(app *echonext.App) {
	users := app.Group("/users")

	users.POST("", h.Create, echonext.Route{
		Summary:       "Create a new user",
		Description:   "Creates a new user with the provided information",
		Tags:          []string{"Users"},
		SuccessStatus: 201,
	})

	users.GET("", h.List, echonext.Route{
		Summary:     "List users",
		Description: "Returns a paginated list of users",
		Tags:        []string{"Users"},
	})

	users.GET("/:id", h.GetByID, echonext.Route{
		Summary:     "Get user by ID",
		Description: "Returns a single user by their ID",
		Tags:        []string{"Users"},
	})

	users.PUT("/:id", h.Update, echonext.Route{
		Summary:     "Update user",
		Description: "Updates an existing user",
		Tags:        []string{"Users"},
	})

	users.DELETE("/:id", h.Delete, echonext.Route{
		Summary:       "Delete user",
		Description:   "Soft-deletes a user by their ID",
		Tags:          []string{"Users"},
		SuccessStatus: 204,
	})
}

// Create handles POST /users
func (h *Handler) Create(c echo.Context, req CreateUserRequest) (UserResponse, error) {
	user, err := h.service.Create(c.Request().Context(), req)
	if err != nil {
		return UserResponse{}, echo.NewHTTPError(400, err.Error())
	}

	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// List handles GET /users
func (h *Handler) List(c echo.Context, req ListUsersRequest) (ListUsersResponse, error) {
	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	response, err := h.service.List(c.Request().Context(), req)
	if err != nil {
		return ListUsersResponse{}, echo.NewHTTPError(500, err.Error())
	}

	return *response, nil
}

// GetByID handles GET /users/:id
func (h *Handler) GetByID(c echo.Context) (UserResponse, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return UserResponse{}, echo.NewHTTPError(400, "invalid user ID")
	}

	user, err := h.service.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return UserResponse{}, echo.NewHTTPError(404, "user not found")
	}

	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Update handles PUT /users/:id
func (h *Handler) Update(c echo.Context, req UpdateUserRequest) (UserResponse, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return UserResponse{}, echo.NewHTTPError(400, "invalid user ID")
	}

	user, err := h.service.Update(c.Request().Context(), uint(id), req)
	if err != nil {
		return UserResponse{}, echo.NewHTTPError(400, err.Error())
	}

	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Delete handles DELETE /users/:id
func (h *Handler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return echo.NewHTTPError(400, "invalid user ID")
	}

	if err := h.service.Delete(c.Request().Context(), uint(id)); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}

	return c.NoContent(204)
}
`)
}

// generateUserDTO creates the User DTOs (Data Transfer Objects)
func (g *ProjectGenerator) generateUserDTO() string {
	return `package user

import "time"

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name  string ` + "`json:\"name\" validate:\"required,min=2,max=100\" example:\"John Doe\"`" + `
	Email string ` + "`json:\"email\" validate:\"required,email\" example:\"john.doe@example.com\"`" + `
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Name  string ` + "`json:\"name,omitempty\" validate:\"omitempty,min=2,max=100\"`" + `
	Email string ` + "`json:\"email,omitempty\" validate:\"omitempty,email\"`" + `
}

// ListUsersRequest represents the request to list users with pagination
type ListUsersRequest struct {
	Page   int    ` + "`query:\"page\" validate:\"min=1\" example:\"1\"`" + `
	Limit  int    ` + "`query:\"limit\" validate:\"min=1,max=100\" example:\"20\"`" + `
	Search string ` + "`query:\"search\" validate:\"max=100\" example:\"john\"`" + `
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        uint      ` + "`json:\"id\" example:\"1\"`" + `
	Name      string    ` + "`json:\"name\" example:\"John Doe\"`" + `
	Email     string    ` + "`json:\"email\" example:\"john.doe@example.com\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []UserResponse ` + "`json:\"users\"`" + `
	Total      int            ` + "`json:\"total\" example:\"100\"`" + `
	Page       int            ` + "`json:\"page\" example:\"1\"`" + `
	Limit      int            ` + "`json:\"limit\" example:\"20\"`" + `
	TotalPages int            ` + "`json:\"total_pages\" example:\"5\"`" + `
}
`
}

// generateUserTest creates the User service tests
func (g *ProjectGenerator) generateUserTest() string {
	return `package user

import (
	"context"
	"testing"
)

// TestService_Create tests the user creation functionality
func TestService_Create(t *testing.T) {
	// Mock repository
	repo := &mockRepository{}
	service := NewService(repo)

	ctx := context.Background()
	req := CreateUserRequest{
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Test successful creation
	t.Run("success", func(t *testing.T) {
		repo.users = make(map[string]*User) // Reset mock data

		user, err := service.Create(ctx, req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if user.Name != req.Name {
			t.Errorf("Expected name %s, got %s", req.Name, user.Name)
		}

		if user.Email != req.Email {
			t.Errorf("Expected email %s, got %s", req.Email, user.Email)
		}
	})

	// Test duplicate email
	t.Run("duplicate email", func(t *testing.T) {
		repo.users = map[string]*User{
			req.Email: {ID: 1, Email: req.Email, Name: "Existing User"},
		}

		_, err := service.Create(ctx, req)
		if err == nil {
			t.Error("Expected error for duplicate email, got nil")
		}
	})
}

// TestService_GetByID tests getting a user by ID
func TestService_GetByID(t *testing.T) {
	repo := &mockRepository{
		users: map[string]*User{
			"test@example.com": {ID: 1, Email: "test@example.com", Name: "Test User"},
		},
	}
	service := NewService(repo)

	ctx := context.Background()

	t.Run("user exists", func(t *testing.T) {
		user, err := service.GetByID(ctx, 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if user.ID != 1 {
			t.Errorf("Expected ID 1, got %d", user.ID)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := service.GetByID(ctx, 999)
		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}
	})
}

// mockRepository is a simple in-memory repository for testing
type mockRepository struct {
	users   map[string]*User
	nextID  uint
}

func (r *mockRepository) Create(ctx context.Context, user *User) error {
	r.nextID++
	user.ID = r.nextID
	r.users[user.Email] = user
	return nil
}

func (r *mockRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (r *mockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, exists := r.users[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (r *mockRepository) Update(ctx context.Context, user *User) error {
	r.users[user.Email] = user
	return nil
}

func (r *mockRepository) Delete(ctx context.Context, id uint) error {
	for email, user := range r.users {
		if user.ID == id {
			delete(r.users, email)
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

func (r *mockRepository) List(ctx context.Context, req ListUsersRequest) ([]*User, int, error) {
	var users []*User
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, len(users), nil
}
`
}