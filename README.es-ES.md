

# EchoNext

EchoNext es un envoltorio con seguridad de tipos para el framework web Echo que genera automáticamente especificaciones OpenAPI y proporciona validación de solicitudes. Construye APIs robustas y bien documentadas con seguridad de tipos en tiempo de compilación.

## Características

- 🔒 **Manejadores con Seguridad de Tipos** - Define manejadores con structs fuertemente tipados para solicitudes y respuestas
- 📚 **Generación Automática de OpenAPI** - Genera especificaciones OpenAPI 3.0 desde tu código
- ✅ **Validación Integrada** - Valida solicitudes usando etiquetas de structs
- 📖 **Swagger UI** - Documentación interactiva de la API lista para usar
- 🚀 **Cero Código Repetitivo** - Concéntrate en la lógica de negocio, no en los detalles de HTTP
- 🔌 **Compatible con Echo** - Usa todo el middleware y características de Echo
- 📁 **Cargas de Archivos** - Cargas de archivos con seguridad de tipos y documentación OpenAPI
- 🔌 **Soporte para WebSocket** - Comunicación en tiempo real con el patrón Hub
- 📊 **Integración con GraphQL** - Integración perfecta con gqlgen y compartición de contexto
- 🛠️ **Herramienta CLI** - Generación de código, recarga automática y creación de proyectos

## Instalación

```bash
go get github.com/abdussamadbello/echonext
```

## Inicio Rápido

```go
package main

import (
    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v5"
)

// Define your request/response types
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // Create new EchoNext app
    app := echonext.New()
    
    // Set API info
    app.SetInfo("User API", "1.0.0", "User management service")
    
    // Register typed routes
    app.POST("/users", createUser, echonext.Route{
        Summary:     "Create a new user",
        Description: "Creates a new user with the provided information",
        Tags:        []string{"Users"},
    })
    
    app.GET("/users/:id", getUser, echonext.Route{
        Summary: "Get user by ID",
        Tags:    []string{"Users"},
    })
    
    // Serve OpenAPI spec and Swagger UI
    app.ServeOpenAPISpec("/api/openapi.json")
    app.ServeSwaggerUI("/api/docs", "/api/openapi.json")
    
    // Start server
    app.Start(":8080")
}

// Handlers with typed parameters
func createUser(c *echo.Context, req CreateUserRequest) (UserResponse, error) {
    // Your business logic here
    user := UserResponse{
        ID:    "123",
        Name:  req.Name,
        Email: req.Email,
    }
    return user, nil
}

func getUser(c *echo.Context) (UserResponse, error) {
    id := c.Param("id")
    // Fetch user logic
    return UserResponse{
        ID:    id,
        Name:  "John Doe",
        Email: "john@example.com",
    }, nil
}
```

## Firmas de los Manejadores

EchoNext admite varias firmas de manejadores:

```go
// No request body (GET, DELETE)
func handler(c *echo.Context) (ResponseType, error)

// With request body (POST, PUT, PATCH)
func handler(c *echo.Context, req RequestType) (ResponseType, error)

// No response body
func handler(c *echo.Context) error
```

## Validación

Usa etiquetas de structs para la validación:

```go
type CreatePostRequest struct {
    Title   string   `json:"title" validate:"required,min=3,max=200"`
    Content string   `json:"content" validate:"required,min=10"`
    Tags    []string `json:"tags" validate:"max=5,dive,min=2,max=20"`
    Status  string   `json:"status" validate:"required,oneof=draft published"`
}
```

## Parámetros de Consulta

Para solicitudes GET, usa etiquetas `query`:

```go
type ListUsersRequest struct {
    Page  int    `query:"page" validate:"min=1"`
    Limit int    `query:"limit" validate:"min=1,max=100"`
    Sort  string `query:"sort" validate:"omitempty,oneof=name email created_at"`
}

func listUsers(c *echo.Context, req ListUsersRequest) (ListResponse, error) {
    // Access validated query params from req
}
```

## Manejo de Errores

Retorna errores desde los manejadores para obtener respuestas de error automáticas:

```go
func getUser(c *echo.Context) (UserResponse, error) {
    id := c.Param("id")
    user, err := db.GetUser(id)
    if err != nil {
        return UserResponse{}, echo.NewHTTPError(404, "user not found")
    }
    return user, nil
}
```

## Middleware y Compatibilidad con Echo

EchoNext es totalmente compatible con todo el middleware y características de Echo. Dado que envuelve `*echo.Echo`, tienes acceso a todo lo que Echo proporciona:

```go
import "github.com/labstack/echo/v5/middleware"

app := echonext.New()

// All standard Echo middleware works
app.Use(middleware.Logger())
app.Use(middleware.Recover())
app.Use(middleware.CORS())
app.Use(middleware.Gzip())
app.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
app.Use(middleware.BasicAuth(func(username, password string, c *echo.Context) (bool, error) {
    return username == "admin" && password == "secret", nil
}))
```

### Características de Echo Disponibles

- **Métodos de contexto**: `c.Param()`, `c.QueryParam()`, `c.FormValue()`, etc.
- **Cargas de archivos**: `c.FormFile()`, `c.MultipartForm()`
- **Archivos estáticos**: `app.Static("/static", "assets")`
- **Grupos de rutas**: `api := app.Group("/api")`
- **Enlazadores personalizados**: Lógica personalizada de enlace de solicitudes
- **Manejo de errores**: Manejador de errores centralizado de Echo
- **Opciones del servidor**: TLS, apagado elegante, etc.

### Ejemplo con Características de Echo

```go
app := echonext.New()

// Use Echo middleware
app.Use(middleware.Logger())
app.Use(middleware.CORS())

// Create route groups (standard Echo)
api := app.Group("/api/v1")

// Static files (standard Echo)
app.Static("/assets", "public")

// EchoNext typed routes work within groups
api.POST("/users", createUser, echonext.Route{
    Summary: "Create user",
    Tags:    []string{"Users"},
})

// Mix typed and standard Echo handlers
app.POST("/upload", func(c *echo.Context) error {
    file, err := c.FormFile("upload")
    if err != nil {
        return err
    }
    // Standard Echo file handling
    return c.String(200, "Uploaded: "+file.Filename)
})
```

EchoNext añade seguridad de tipos y generación de OpenAPI **encima de** Echo sin eliminar ninguna funcionalidad.

## Características Avanzadas de OpenAPI

### Esquemas de Seguridad

Define los requisitos de seguridad para tu API:

```go
app := echonext.New()

// Add security schemes
app.AddSecurityScheme("bearerAuth", echonext.Security{
    Type:   "bearer",
    Scheme: "JWT",
})
app.AddSecurityScheme("apiKey", echonext.Security{
    Type: "apiKey",
    Name: "X-API-Key",
    In:   "header",
})

// Apply security to routes
app.POST("/protected", handler, echonext.Route{
    Security: []echonext.Security{
        {Type: "bearer"},
        {Type: "apiKey", Name: "X-API-Key"},
    },
})
```

### Códigos de Estado de Respuesta Personalizados

Usa códigos de estado HTTP apropiados:

```go
app.POST("/users", createUser, echonext.Route{
    SuccessStatus: 201, // Returns 201 Created instead of 200 OK
})

app.DELETE("/users/:id", deleteUser, echonext.Route{
    SuccessStatus: 204, // Returns 204 No Content
})
```

### Encabezados de Solicitud/Respuesta

Documenta encabezados requeridos y opcionales:

```go
app.POST("/upload", uploadHandler, echonext.Route{
    RequestHeaders: map[string]echonext.HeaderInfo{
        "X-Request-ID": {
            Description: "Unique request identifier",
            Required:    true,
            Schema:      "string",
        },
    },
    ResponseHeaders: map[string]echonext.HeaderInfo{
        "X-Upload-ID": {
            Description: "ID of uploaded file",
            Schema:      "string",
        },
    },
})
```

### Tipos de Contenido y Ejemplos

Soporta múltiples tipos de contenido y proporciona ejemplos:

```go
type CreateUserRequest struct {
    Name string `json:"name" example:"John Doe"`
    Age  int    `json:"age" example:"30"`
}

app.POST("/users", createUser, echonext.Route{
    ContentTypes: []string{"application/json", "application/xml"},
    Examples: map[string]interface{}{
        "basic": map[string]interface{}{
            "name": "John Doe",
            "age":  30,
        },
    },
})
```

### Configuración Completa de la API

```go
app := echonext.New()

// Set comprehensive API information
app.SetInfo("My API", "1.0.0", "A comprehensive API example")
app.SetContact("API Team", "https://example.com/support", "api@example.com")
app.SetLicense("MIT", "https://opensource.org/licenses/MIT")
app.SetServers([]echonext.Server{
    {URL: "https://api.example.com/v1", Description: "Production"},
    {URL: "https://staging.example.com/v1", Description: "Staging"},
})
```

## Cargas de Archivos

EchoNext proporciona soporte para cargas de archivos con seguridad de tipos y documentación OpenAPI automática:

```go
import "github.com/abdussamadbello/echonext/upload"

type AvatarRequest struct {
    File *upload.File `form:"avatar" validate:"required"`
}

type AvatarResponse struct {
    URL string `json:"url"`
}

func uploadAvatar(c *echo.Context, req AvatarRequest) (AvatarResponse, error) {
    // Access file metadata
    fmt.Printf("Filename: %s, Size: %d\n", req.File.Filename, req.File.Size)

    // Save the file easily
    if err := req.File.SaveTo("/uploads/" + req.File.Filename); err != nil {
        return AvatarResponse{}, err
    }

    return AvatarResponse{URL: "/uploads/" + req.File.Filename}, nil
}

// Register upload endpoint
app.Upload("/avatar", uploadAvatar, echonext.Route{
    Summary: "Upload avatar image",
})
```

### Archivos Múltiples y Configuración

```go
type DocumentsRequest struct {
    Files []*upload.File `form:"documents" validate:"max=10"`
}

app.Upload("/documents", handler, echonext.Route{
    FileConfig: &echonext.FileUploadConfig{
        MaxFileSize:       10 << 20,  // 10MB per file
        MaxTotalSize:      50 << 20,  // 50MB total
        AllowedMIMETypes:  []string{"image/jpeg", "image/png", "application/pdf"},
        AllowedExtensions: []string{".jpg", ".png", ".pdf"},
        MaxFiles:          5,
    },
})
```

Genera manejadores de carga con la CLI:

```bash
echonext generate upload avatar
```

Consulta [examples/upload-demo/](examples/upload-demo/) para un ejemplo completo.

## Soporte para WebSocket

Manejadores de WebSocket con seguridad de tipos y gestión de conexiones:

```go
import "github.com/abdussamadbello/echonext/websocket"

// Simple handler
func chatHandler(conn *websocket.Connection) error {
    for {
        var msg ChatMessage
        if err := conn.ReadJSON(&msg); err != nil {
            return err
        }
        response := ChatResponse{Text: "Echo: " + msg.Text}
        if err := conn.WriteJSON(response); err != nil {
            return err
        }
    }
}

app.WS("/chat", chatHandler)
```

### Patrón Hub para Difusión

```go
type ChatHandler struct {
    hub *websocket.Hub
}

func (h *ChatHandler) OnConnect(conn *websocket.Connection) error {
    h.hub.Register(conn)
    return nil
}

func (h *ChatHandler) OnMessage(conn *websocket.Connection, msg []byte) error {
    return h.hub.Broadcast(msg)  // Broadcast to all connections
}

func (h *ChatHandler) OnDisconnect(conn *websocket.Connection, err error) {
    h.hub.Unregister(conn)
}

// Usage
hub := websocket.NewHub()
go hub.Run()

app.WS("/ws/chat", &ChatHandler{hub: hub})
```

Genera manejadores de WebSocket con la CLI:

```bash
echonext generate websocket chat
```

Consulta [examples/websocket-demo/](examples/websocket-demo/) para un ejemplo completo.

## Integración con GraphQL

Integración perfecta con gqlgen:

```go
import "github.com/abdussamadbello/echonext/graphql"

app.GraphQL(graphql.Config{
    Path:           "/graphql",
    PlaygroundPath: "/playground",
    Schema:         graph.NewExecutableSchema(graph.Config{
        Resolvers: graph.NewResolver(),
    }),
})
```

### Acceso al Contexto de Echo en Resolvers

```go
func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
    echoCtx := graphql.GetEchoContext(ctx)
    userID := echoCtx.Get("user_id").(string)
    return r.userService.GetByID(userID)
}
```

### Opciones de Configuración de GraphQL

```go
graphql.Config{
    Path:                "/graphql",
    PlaygroundPath:      "/playground",  // Empty to disable
    Schema:              schema,
    ComplexityLimit:     100,
    QueryCacheSize:      1000,
    EnableIntrospection: true,
}
```

Genera el código base de GraphQL con la CLI:

```bash
echonext generate graphql
```

Consulta [examples/graphql-demo/](examples/graphql-demo/) para un ejemplo completo.

## Generación de Código desde OpenAPI

Genera código de EchoNext desde especificaciones OpenAPI existentes:

```bash
# From local file
echonext generate openapi api.yaml

# From URL
echonext generate openapi https://api.example.com/openapi.json

# With options
echonext generate openapi api.yaml --output=./generated --package=api
```

Archivos generados:
- `models/models.go` - Modelos de datos desde componentes del esquema
- `dto/dto.go` - DTOs de solicitud/respuesta
- `handlers/handlers.go` - Plantillas de funciones de manejador
- `routes.go` - Registro de rutas

## Aplicación de Ejemplo

Ejecuta la API de ejemplo Todo:

```bash
go run example/main.go
```

Luego visita:
- Servidor de la API: http://localhost:8080
- Documentación de la API: http://localhost:8080/api/docs
- Especificación OpenAPI: http://localhost:8080/api/openapi.json

## Desarrollo

### Ejecutar Pruebas

```bash
go test ./...                    # Run all tests
go test -v ./...                # Run with verbose output
go test -bench=.                # Run benchmarks
go test -cover                  # Run with coverage
```

### Estructura del Proyecto

```
echonext/
├── echonext.go              # Main package implementation
├── echonext_test.go         # Test suite
├── upload/                  # File upload package
│   └── upload.go
├── websocket/               # WebSocket package
│   └── websocket.go
├── graphql/                 # GraphQL integration
│   └── graphql.go
├── cmd/echonext-cli/        # CLI tool
│   ├── commands.go
│   └── generator/           # Code generation templates
├── examples/
│   ├── graphql-demo/        # GraphQL example
│   ├── websocket-demo/      # WebSocket chat example
│   └── upload-demo/         # File upload example
├── pkg/contrib/             # Optional helper packages
└── example/main.go          # Quick start example
```

## Formato de Respuesta de la API

Todas las respuestas se envuelven en un formato consistente:

```json
{
  "success": true,
  "data": { ... },
  "error": ""
}
```

Respuestas de error:

```json
{
  "success": false,
  "data": null,
  "error": "Validation failed: Name is required"
}
```

## Paquetes Contrib Opcionales

EchoNext proporciona paquetes auxiliares opcionales en `pkg/contrib/` para tareas comunes. Estos son completamente opcionales: puedes usar las bibliotecas subyacentes directamente si lo prefieres.

### 📦 Base de Datos (`pkg/contrib/database`)

Ayudantes de integración con GORM con:
- Gestión de conexiones con lógica de reintento
- Patrón Repository[T] genérico
- Utilidades de transacciones
- Ayudantes de migración
- **Integración con Atlas** para migraciones de esquema

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

cfg := database.DefaultConfig()
db, err := database.Connect(postgres.Open(dsn), cfg)

// Use repository pattern with generics
userRepo := database.NewRepository[User](db)
user, err := userRepo.Find(1)
users, err := userRepo.Where("active = ?", true).FindAll()
```

### 🔄 Migraciones de Base de Datos (Atlas)

EchoNext utiliza [Atlas](https://atlasgo.io) para la gestión de esquemas de base de datos:

```bash
# Initialize Atlas in your project
echonext db init

# Apply migrations
echonext db migrate

# Generate migration from schema changes
echonext db migrate:diff add_users_table

# Check migration status
echonext db migrate:status

# Rollback migrations
echonext db migrate:down --count=1
```

**Esquema Declarativo** - Define tu esquema en `schema.hcl`:

```hcl
table "users" {
  schema = schema.public
  column "id" { type = bigserial }
  column "email" { type = varchar(255) }
  primary_key { columns = [column.id] }
}
```

Consulta [CLAUDE.md](CLAUDE.md#database-migrations-with-atlas) para documentación detallada de Atlas.

### ⚙️ Configuración (`pkg/contrib/config`)

Ayudantes de integración con Viper con:
- Carga de configuración genérica
- Vinculación de variables de entorno
- Soporte para recarga automática
- Estructuras de configuración estándar

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/config"

type MyConfig struct {
    App      config.AppConfig      `mapstructure:"app"`
    Database config.DatabaseConfig `mapstructure:"database"`
}

var cfg MyConfig
config.LoadSimple(&cfg)
```

### 🧪 Pruebas (`pkg/contrib/testing`)

Utilidades de pruebas con:
- APIClient para probar endpoints
- FixtureManager para datos de prueba
- Suite de pruebas con configuración/limpieza
- Patrón Factory para entidades de prueba

```go
import echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"

client := echonexttest.NewAPIClient(app)
resp := client.POST("/users", userRequest)
resp.AssertStatus(t, 201).AssertSuccess(t)
```

Consulta [pkg/contrib/README.md](pkg/contrib/README.md) para documentación detallada.

## 🎓 Proyectos de Ejemplo

¡Aprende con ejemplos! Consulta nuestros proyectos de ejemplo completos:

### [⚡ Inicio Rápido](example/main.go) - Ejemplo Funcionando
API Todo completamente funcional. ¡Ejécútala ahora!

```bash
go run example/main.go
# Visit http://localhost:8080/api/docs
```

### [📝 API de Lista de Tareas](examples/todo-api/) - Principiante
Operaciones CRUD simples que demuestran los fundamentos de EchoNext.

```bash
echonext init todo-api
echonext generate domain todo
go run ./cmd/api
```

### [📰 API de Blog](examples/blog-api/) - Intermedio
Plataforma de blog multidominio con autenticación, búsqueda y relaciones.

```bash
echonext init blog-api
echonext generate domain post
echonext generate domain comment
echonext generate domain user
```

### [🛒 API de Comercio Electrónico](examples/ecommerce-api/) - Avanzado
Plataforma de comercio electrónico completa con pedidos, pagos e inventario.

```bash
echonext init ecommerce-api
echonext generate domain product
echonext generate domain order
echonext generate domain payment
```

### [🔧 Microservicios](examples/microservice/) - Experto
Sistema distribuido con comunicación entre servicios y eventos.

Consulta [examples/README.md](examples/README.md) para guías detalladas y más ejemplos.

## Contribuir

1. Haz un fork del repositorio
2. Crea tu rama de funcionalidad (`git checkout -b feature/nueva-funcionalidad`)
3. Commit de tus cambios (`git commit -m 'Agregar una nueva funcionalidad'`)
4. Haz push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Abre un Pull Request

## Licencia

Licencia MIT - consulta el archivo [LICENSE](LICENSE) para más detalles.

## Hoja de Ruta

**Completado:**
- [x] ✅ Ayudantes de integración con base de datos (ver `pkg/contrib/database`)
- [x] ✅ Ayudantes de gestión de configuración (ver `pkg/contrib/config`)
- [x] ✅ Utilidades de pruebas (ver `pkg/contrib/testing`)
- [x] ✅ Ayudantes de middleware (ver `pkg/contrib/middleware`)
- [x] ✅ Herramienta CLI para generación de proyectos (`echonext init`)
- [x] ✅ Comandos de generación de código (`echonext generate domain/handler/service/model/dto`)
- [x] ✅ Comandos de gestión de base de datos (`echonext db init/migrate/seed`)
- [x] ✅ Integración de migración con Atlas para gestión de esquemas
- [x] ✅ Proyectos de ejemplo completos (Todo, Blog, Comercio Electrónico, Microservicios)
- [x] ✅ Integración con OpenTelemetry (ver `pkg/contrib/middleware`)

**v1.4.0:**
- [x] ✅ Comando de desarrollo con recarga automática (`echonext dev`)
- [x] ✅ Ejecutor de pruebas mejorado (`echonext test`)
- [x] ✅ Automatización de compilación (`echonext build`)
- [x] ✅ Soporte para cargas de archivos en la especificación OpenAPI
- [x] ✅ Soporte para WebSocket con seguridad de tipos
- [x] ✅ Integración con GraphQL
- [x] ✅ Generación de código desde especificación OpenAPI

**Planificado:**
- [ ] Sistema de plugins para generadores personalizados
- [ ] Soporte para gRPC
- [ ] Ayudantes de versionamiento de API
- [ ] Eventos Enviados por el Servidor (SSE)
