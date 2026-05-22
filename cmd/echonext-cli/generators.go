package main

import (
	"fmt"
	"strings"
)

// DomainGenerator generates domain files
type DomainGenerator struct {
	Name   string // lowercase name like "user"
	Module string // module path like "github.com/user/myapp"
}

// PascalCase converts name to PascalCase
func (g *DomainGenerator) PascalCase() string {
	parts := strings.Split(g.Name, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}

// Pluralize returns plural form
func (g *DomainGenerator) Pluralize() string {
	if strings.HasSuffix(g.Name, "s") {
		return g.Name
	}
	return g.Name + "s"
}

func (g *DomainGenerator) generateModel() string {
	entityName := g.PascalCase()
	return fmt.Sprintf(`package %s

import (
	"time"

	"gorm.io/gorm"
)

// %s represents a %s entity
type %s struct {
	ID        uint           %sgorm:"primaryKey"%s
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt %sgorm:"index"%s

	// TODO: Add your fields here
	// Name  string %sgorm:"not null"%s
	// Email string %sgorm:"unique;not null"%s
}

// TableName specifies the table name for the %s model
func (%s) TableName() string {
	return "%s"
}
`, g.Name, entityName, g.Name, entityName, "`", "`", "`", "`", "`", "`", "`", "`", entityName, entityName, g.Pluralize())
}

func (g *DomainGenerator) generateService() string {
	entityName := g.PascalCase()
	return fmt.Sprintf(`package %s

import (
	"fmt"

	"gorm.io/gorm"
)

// Service handles business logic for %s
type Service struct {
	db *gorm.DB
}

// NewService creates a new %s service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Create creates a new %s
func (s *Service) Create(%s *%s) error {
	if err := s.db.Create(%s).Error; err != nil {
		return fmt.Errorf("failed to create %s: %%w", err)
	}
	return nil
}

// GetByID retrieves a %s by ID
func (s *Service) GetByID(id uint) (*%s, error) {
	var %s %s
	if err := s.db.First(&%s, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%s not found")
		}
		return nil, fmt.Errorf("failed to get %s: %%w", err)
	}
	return &%s, nil
}

// List retrieves all %s
func (s *Service) List(limit, offset int) ([]*%s, error) {
	var %s []*%s
	query := s.db.Limit(limit).Offset(offset)

	if err := query.Find(&%s).Error; err != nil {
		return nil, fmt.Errorf("failed to list %s: %%w", err)
	}

	return %s, nil
}

// Update updates a %s
func (s *Service) Update(%s *%s) error {
	if err := s.db.Save(%s).Error; err != nil {
		return fmt.Errorf("failed to update %s: %%w", err)
	}
	return nil
}

// Delete deletes a %s by ID
func (s *Service) Delete(id uint) error {
	if err := s.db.Delete(&%s{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete %s: %%w", err)
	}
	return nil
}

// Count returns the total count of %s
func (s *Service) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&%s{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count %s: %%w", err)
	}
	return count, nil
}
`, g.Name, g.Pluralize(), g.Name, g.Name, g.Name, entityName, g.Name, g.Name,
		g.Name, entityName, g.Name, entityName, g.Name, g.Name, g.Name, g.Name,
		g.Pluralize(), entityName, g.Pluralize(), entityName, g.Pluralize(), g.Pluralize(), g.Pluralize(),
		g.Name, g.Name, entityName, g.Name, g.Name, g.Name, entityName, g.Name,
		g.Pluralize(), entityName, g.Pluralize())
}

func (g *DomainGenerator) generateHandler() string {
	entityName := g.PascalCase()
	return fmt.Sprintf(`package %s

import (
	"net/http"
	"strconv"

	"github.com/abdussamadbello/echonext"
	"github.com/labstack/echo/v5"
)

// Handler handles HTTP requests for %s
type Handler struct {
	service *Service
}

// NewHandler creates a new %s handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the %s routes
func (h *Handler) RegisterRoutes(app *echonext.App) {
	app.POST("/%s", h.Create, echonext.Route{
		Summary:     "Create a new %s",
		Description: "Creates a new %s with the provided data",
		Tags:        []string{"%s"},
	})

	app.GET("/%s/:id", h.Get, echonext.Route{
		Summary:     "Get %s by ID",
		Description: "Retrieves a %s by its ID",
		Tags:        []string{"%s"},
	})

	app.GET("/%s", h.List, echonext.Route{
		Summary:     "List %s",
		Description: "Retrieves a list of %s",
		Tags:        []string{"%s"},
	})

	app.PUT("/%s/:id", h.Update, echonext.Route{
		Summary:     "Update %s",
		Description: "Updates an existing %s",
		Tags:        []string{"%s"},
	})

	app.DELETE("/%s/:id", h.Delete, echonext.Route{
		Summary:     "Delete %s",
		Description: "Deletes a %s by ID",
		Tags:        []string{"%s"},
	})
}

// Create handles creating a new %s
func (h *Handler) Create(c *echo.Context, req Create%sRequest) (%sResponse, error) {
	// TODO: Map request to model
	%s := &%s{
		// Map fields from req
	}

	if err := h.service.Create(%s); err != nil {
		return %sResponse{}, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return To%sResponse(%s), nil
}

// Get handles retrieving a %s by ID
func (h *Handler) Get(c *echo.Context) (%sResponse, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return %sResponse{}, echo.NewHTTPError(http.StatusBadRequest, "invalid ID")
	}

	%s, err := h.service.GetByID(uint(id))
	if err != nil {
		return %sResponse{}, echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return To%sResponse(%s), nil
}

// List handles listing %s
func (h *Handler) List(c *echo.Context, req List%sRequest) (List%sResponse, error) {
	%s, err := h.service.List(req.Limit, req.Offset)
	if err != nil {
		return List%sResponse{}, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	items := make([]%sResponse, len(%s))
	for i, item := range %s {
		items[i] = To%sResponse(item)
	}

	total, _ := h.service.Count()

	return List%sResponse{
		Items:  items,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// Update handles updating a %s
func (h *Handler) Update(c *echo.Context, req Update%sRequest) (%sResponse, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return %sResponse{}, echo.NewHTTPError(http.StatusBadRequest, "invalid ID")
	}

	%s, err := h.service.GetByID(uint(id))
	if err != nil {
		return %sResponse{}, echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	// TODO: Update fields from request
	// %s.Name = req.Name

	if err := h.service.Update(%s); err != nil {
		return %sResponse{}, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return To%sResponse(%s), nil
}

// Delete handles deleting a %s
func (h *Handler) Delete(c *echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ID")
	}

	if err := h.service.Delete(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
`, g.Name, g.Pluralize(), g.Name, g.Name,
		g.Pluralize(), g.Name, g.Name, entityName,
		g.Pluralize(), g.Name, g.Name, entityName,
		g.Pluralize(), g.Pluralize(), g.Pluralize(), entityName,
		g.Pluralize(), g.Name, g.Name, entityName,
		g.Pluralize(), g.Name, g.Name, entityName,
		g.Name, entityName, entityName, g.Name, entityName, g.Name, entityName, entityName, g.Name,
		g.Name, entityName, entityName, g.Name, entityName, entityName, g.Name,
		g.Pluralize(), entityName, entityName, g.Pluralize(), entityName, entityName, g.Pluralize(), g.Pluralize(), entityName,
		entityName, g.Name, entityName, entityName, entityName, g.Name, entityName, g.Name, g.Name, entityName, entityName,
		g.Name, g.Name)
}

func (g *DomainGenerator) generateDTO() string {
	entityName := g.PascalCase()
	return fmt.Sprintf(`package %s

import "time"

// Create%sRequest represents the request to create a %s
type Create%sRequest struct {
	// TODO: Add your request fields
	// Name  string %sjson:"name" validate:"required,min=2"%s
	// Email string %sjson:"email" validate:"required,email"%s
}

// Update%sRequest represents the request to update a %s
type Update%sRequest struct {
	// TODO: Add your request fields
	// Name  string %sjson:"name" validate:"omitempty,min=2"%s
	// Email string %sjson:"email" validate:"omitempty,email"%s
}

// List%sRequest represents the request to list %s
type List%sRequest struct {
	Limit  int %squery:"limit" validate:"min=1,max=100"%s
	Offset int %squery:"offset" validate:"min=0"%s
}

// %sResponse represents a %s in responses
type %sResponse struct {
	ID        uint      %sjson:"id"%s
	CreatedAt time.Time %sjson:"created_at"%s
	UpdatedAt time.Time %sjson:"updated_at"%s

	// TODO: Add your response fields
	// Name  string %sjson:"name"%s
	// Email string %sjson:"email"%s
}

// List%sResponse represents a list of %s
type List%sResponse struct {
	Items  []%sResponse %sjson:"items"%s
	Total  int64          %sjson:"total"%s
	Limit  int            %sjson:"limit"%s
	Offset int            %sjson:"offset"%s
}

// To%sResponse converts a %s model to a response
func To%sResponse(%s *%s) %sResponse {
	return %sResponse{
		ID:        %s.ID,
		CreatedAt: %s.CreatedAt,
		UpdatedAt: %s.UpdatedAt,
		// TODO: Map your fields
		// Name:  %s.Name,
		// Email: %s.Email,
	}
}
`, g.Name, entityName, g.Name, entityName, "`", "`", "`", "`",
		entityName, g.Name, entityName, "`", "`", "`", "`",
		entityName, g.Pluralize(), entityName, "`", "`", "`", "`",
		entityName, g.Name, entityName, "`", "`", "`", "`", "`", "`", "`", "`", "`", "`",
		entityName, g.Pluralize(), entityName, entityName, "`", "`", "`", "`", "`", "`", "`", "`",
		entityName, entityName, entityName, g.Name, entityName, entityName, entityName, g.Name, g.Name, g.Name, g.Name, g.Name)
}

// MiddlewareGenerator generates middleware files
type MiddlewareGenerator struct {
	Name string // lowercase name like "auth"
}

func (g *MiddlewareGenerator) PascalCase() string {
	parts := strings.Split(g.Name, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}

func (g *MiddlewareGenerator) generate() string {
	name := g.PascalCase()
	return fmt.Sprintf(`package middleware

import (
	"github.com/labstack/echo/v5"
)

// %s is a custom middleware for %s
func %s() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// TODO: Implement your middleware logic here

			// Example: Add a custom header
			c.Response().Header().Set("X-%s", "enabled")

			// Call the next handler
			return next(c)
		}
	}
}

// %sConfig holds configuration for %s middleware
type %sConfig struct {
	// TODO: Add your configuration fields
	// Enabled bool
}

// %sWithConfig returns a %s middleware with custom configuration
func %sWithConfig(config %sConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// TODO: Implement your middleware logic with config

			return next(c)
		}
	}
}
`, name, g.Name, name, name, name, g.Name, name, name, name, name, name)
}
