# Quickstart Example

A simple working Todo API demonstrating all core EchoNext features.

## Running the Example

```bash
# From the repository root
go run example/main.go

# Or from this directory
go run ../../example/main.go
```

The server will start on http://localhost:8080

## Available Endpoints

- `POST /todos` - Create a new todo
- `GET /todos/:id` - Get todo by ID
- `GET /todos` - List todos with pagination
- `PUT /todos/:id` - Update a todo
- `DELETE /todos/:id` - Delete a todo
- `/api/docs` - Swagger UI
- `/api/openapi.json` - OpenAPI spec

## Example Request

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy groceries", "description": "Milk, eggs, bread", "priority": "high"}'
```

## What's Demonstrated

- ✅ Type-safe handlers
- ✅ Automatic validation
- ✅ OpenAPI generation
- ✅ Swagger UI
- ✅ Security schemes
- ✅ Custom status codes

## Next Steps

- 📝 [Todo API Guide](../todo-api/) - Build your own
- 📰 [Blog API](../blog-api/) - More complex example
- 🛒 [E-commerce API](../ecommerce-api/) - Advanced patterns
