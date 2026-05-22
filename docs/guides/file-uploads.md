# File Uploads

EchoNext provides type-safe file upload support with automatic OpenAPI documentation.

## Basic Usage

```go
import (
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/upload"
    "github.com/labstack/echo/v5"
)

type AvatarRequest struct {
    File *upload.File `form:"avatar" validate:"required"`
}

type AvatarResponse struct {
    URL      string `json:"url"`
    Filename string `json:"filename"`
    Size     int64  `json:"size"`
}

func uploadAvatar(c *echo.Context, req AvatarRequest) (AvatarResponse, error) {
    // Access file metadata
    fmt.Printf("Filename: %s\n", req.File.Filename)
    fmt.Printf("Size: %d bytes\n", req.File.Size)
    fmt.Printf("Content-Type: %s\n", req.File.ContentType)

    // Save the file
    destPath := "/uploads/" + req.File.Filename
    if err := req.File.SaveTo(destPath); err != nil {
        return AvatarResponse{}, echo.NewHTTPError(500, "Failed to save file")
    }

    return AvatarResponse{
        URL:      destPath,
        Filename: req.File.Filename,
        Size:     req.File.Size,
    }, nil
}

func main() {
    app := echonext.New()

    app.Upload("/avatar", uploadAvatar, echonext.Route{
        Summary:     "Upload avatar image",
        Description: "Upload a user avatar image",
        Tags:        []string{"Users"},
    })

    app.Start(":8080")
}
```

## File Type

The `upload.File` type provides:

```go
type File struct {
    Filename    string              // Original filename
    Size        int64               // File size in bytes
    ContentType string              // MIME type
    Header      *multipart.FileHeader // Underlying multipart header
}

// Methods
func (f *File) Open() (multipart.File, error)  // Open file for reading
func (f *File) Read() ([]byte, error)          // Read entire file into memory
func (f *File) SaveTo(path string) error       // Save to destination path
```

## Multiple File Uploads

```go
type DocumentsRequest struct {
    Files []*upload.File `form:"documents" validate:"required,max=10"`
}

type DocumentsResponse struct {
    Uploaded int      `json:"uploaded"`
    Files    []string `json:"files"`
}

func uploadDocuments(c *echo.Context, req DocumentsRequest) (DocumentsResponse, error) {
    var filenames []string

    for _, file := range req.Files {
        destPath := "/uploads/documents/" + file.Filename
        if err := file.SaveTo(destPath); err != nil {
            return DocumentsResponse{}, echo.NewHTTPError(500, "Failed to save: "+file.Filename)
        }
        filenames = append(filenames, file.Filename)
    }

    return DocumentsResponse{
        Uploaded: len(req.Files),
        Files:    filenames,
    }, nil
}

app.Upload("/documents", uploadDocuments, echonext.Route{
    Summary: "Upload multiple documents",
    Tags:    []string{"Documents"},
})
```

## File Upload Configuration

Configure file validation using `FileUploadConfig`:

```go
app.Upload("/upload", handler, echonext.Route{
    Summary: "Upload file with validation",
    FileConfig: &echonext.FileUploadConfig{
        // Maximum size per file (10MB)
        MaxFileSize: 10 << 20,

        // Maximum total size for all files (50MB)
        MaxTotalSize: 50 << 20,

        // Allowed MIME types
        AllowedMIMETypes: []string{
            "image/jpeg",
            "image/png",
            "image/gif",
            "application/pdf",
        },

        // Allowed file extensions
        AllowedExtensions: []string{
            ".jpg", ".jpeg", ".png", ".gif", ".pdf",
        },

        // Maximum number of files
        MaxFiles: 5,
    },
})
```

## Mixed Form Data

Combine file uploads with other form fields:

```go
type ProfileUpdateRequest struct {
    Name   string       `form:"name" validate:"required,min=2"`
    Bio    string       `form:"bio" validate:"max=500"`
    Avatar *upload.File `form:"avatar"`
}

type ProfileResponse struct {
    Name      string `json:"name"`
    Bio       string `json:"bio"`
    AvatarURL string `json:"avatar_url,omitempty"`
}

func updateProfile(c *echo.Context, req ProfileUpdateRequest) (ProfileResponse, error) {
    response := ProfileResponse{
        Name: req.Name,
        Bio:  req.Bio,
    }

    // Avatar is optional
    if req.Avatar != nil {
        destPath := "/uploads/avatars/" + req.Avatar.Filename
        if err := req.Avatar.SaveTo(destPath); err != nil {
            return ProfileResponse{}, echo.NewHTTPError(500, "Failed to save avatar")
        }
        response.AvatarURL = destPath
    }

    return response, nil
}
```

## Custom File Processing

Process files before saving:

```go
import (
    "image"
    "image/jpeg"
    _ "image/png"
)

func uploadAndResize(c *echo.Context, req AvatarRequest) (AvatarResponse, error) {
    // Open the uploaded file
    src, err := req.File.Open()
    if err != nil {
        return AvatarResponse{}, echo.NewHTTPError(500, "Failed to open file")
    }
    defer src.Close()

    // Decode image
    img, _, err := image.Decode(src)
    if err != nil {
        return AvatarResponse{}, echo.NewHTTPError(400, "Invalid image format")
    }

    // Resize image (using your preferred library)
    resized := resize(img, 200, 200)

    // Save resized image
    destPath := "/uploads/avatars/" + req.File.Filename
    dst, err := os.Create(destPath)
    if err != nil {
        return AvatarResponse{}, echo.NewHTTPError(500, "Failed to create file")
    }
    defer dst.Close()

    if err := jpeg.Encode(dst, resized, &jpeg.Options{Quality: 85}); err != nil {
        return AvatarResponse{}, echo.NewHTTPError(500, "Failed to encode image")
    }

    return AvatarResponse{
        URL:      destPath,
        Filename: req.File.Filename,
    }, nil
}
```

## OpenAPI Documentation

File uploads are automatically documented in OpenAPI:

```yaml
paths:
  /avatar:
    post:
      summary: Upload avatar image
      requestBody:
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                avatar:
                  type: string
                  format: binary
              required:
                - avatar
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AvatarResponse'
```

## CLI Generator

Generate upload handler boilerplate:

```bash
echonext generate upload avatar
```

This creates:

```
internal/upload/avatar/
├── handler.go     # Upload handler with validation
└── dto.go         # Request/Response types
```

## Best Practices

1. **Validate file types** - Always validate MIME types and extensions
2. **Limit file sizes** - Set appropriate size limits to prevent abuse
3. **Use unique filenames** - Generate UUIDs or hashes to prevent overwrites
4. **Scan for malware** - Consider integrating virus scanning for user uploads
5. **Store securely** - Use cloud storage (S3, GCS) for production
6. **Set proper permissions** - Ensure upload directories have correct permissions

## Error Handling

```go
func uploadWithErrors(c *echo.Context, req AvatarRequest) (AvatarResponse, error) {
    // Check file size
    if req.File.Size > 5<<20 { // 5MB
        return AvatarResponse{}, echo.NewHTTPError(400, "File too large (max 5MB)")
    }

    // Check content type
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
    }
    if !allowedTypes[req.File.ContentType] {
        return AvatarResponse{}, echo.NewHTTPError(400, "Invalid file type")
    }

    // Save file
    if err := req.File.SaveTo("/uploads/" + req.File.Filename); err != nil {
        return AvatarResponse{}, echo.NewHTTPError(500, "Failed to save file")
    }

    return AvatarResponse{URL: "/uploads/" + req.File.Filename}, nil
}
```

## Example Project

See [examples/upload-demo/](../../examples/upload-demo/) for a complete working example with:

- Single and multiple file uploads
- File validation
- Preview interface
- Health check endpoint
- Swagger documentation
