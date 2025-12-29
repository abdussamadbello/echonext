package graph

import (
	"sync"

	"github.com/abdussamadbello/echonext/examples/graphql-demo/graph/model"
)

// Resolver holds the application state and dependencies
type Resolver struct {
	users []*model.User
	posts []*model.Post
	mu    sync.RWMutex
}

// NewResolver creates a new resolver with sample data
func NewResolver() *Resolver {
	r := &Resolver{
		users: []*model.User{
			{ID: "1", Name: "Alice", Email: "alice@example.com"},
			{ID: "2", Name: "Bob", Email: "bob@example.com"},
			{ID: "3", Name: "Charlie", Email: "charlie@example.com"},
		},
		posts: []*model.Post{
			{ID: "1", Title: "Hello GraphQL", Content: "Getting started with GraphQL and EchoNext", AuthorID: "1"},
			{ID: "2", Title: "EchoNext Tips", Content: "Building type-safe APIs with EchoNext", AuthorID: "2"},
			{ID: "3", Title: "Go Best Practices", Content: "Writing clean Go code", AuthorID: "1"},
		},
	}
	return r
}

// GetUserByID finds a user by ID
func (r *Resolver) GetUserByID(id string) *model.User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.ID == id {
			return u
		}
	}
	return nil
}

// GetPostsByAuthorID finds posts by author ID
func (r *Resolver) GetPostsByAuthorID(authorID string) []*model.Post {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.Post
	for _, p := range r.posts {
		if p.AuthorID == authorID {
			result = append(result, p)
		}
	}
	return result
}
