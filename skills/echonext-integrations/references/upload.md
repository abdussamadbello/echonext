# File Upload — Reference

Package `github.com/abdussamadbello/echonext/upload`. Register upload routes
with `app.Upload(...)`; the handler is an ordinary typed handler whose request
struct has `*upload.File` (or `[]*upload.File`) fields tagged `form:"..."`.

## Types

```go
type File struct {
    Filename    string
    Size        int64
    ContentType string
    Header      textproto.MIMEHeader
    // underlying file handle is internal
}

type Config struct {
    MaxFileSize       int64    // bytes, per file
    MaxTotalSize      int64    // bytes, all files
    AllowedMIMETypes  []string // e.g. {"image/png", "image/jpeg"}
    AllowedExtensions []string // e.g. {".png", ".jpg"}
    MaxFiles          int
    PreserveFile      bool
}

func DefaultConfig() Config
```

`*File` methods:

```go
func (f *File) Open() (io.ReadCloser, error)
func (f *File) Read() ([]byte, error)
func (f *File) Close() error
func (f *File) Extension() string
func (f *File) SaveTo(destPath string) error
```

## Single-file upload

```go
type AvatarRequest struct {
    File *upload.File `form:"file" validate:"required"`
}

type AvatarResponse struct {
    Filename     string `json:"filename"`
    OriginalName string `json:"original_name"`
    Size         int64  `json:"size"`
    ContentType  string `json:"content_type"`
    URL          string `json:"url"`
}

func uploadAvatar(c *echo.Context, req AvatarRequest) (AvatarResponse, error) {
    f := req.File
    name := fmt.Sprintf("%d%s", time.Now().UnixNano(), f.Extension())
    if err := f.SaveTo(filepath.Join("uploads", name)); err != nil {
        return AvatarResponse{}, echo.NewHTTPError(http.StatusInternalServerError, "save failed")
    }
    return AvatarResponse{
        Filename:     name,
        OriginalName: f.Filename,
        Size:         f.Size,
        ContentType:  f.ContentType,
        URL:          "/uploads/" + name,
    }, nil
}
```

Register with limits via `Route.FileConfig`:

```go
app.Upload("/avatar", uploadAvatar, echonext.Route{
    Summary: "Upload avatar",
    Tags:    []string{"Avatar"},
    FileConfig: &upload.Config{
        MaxFileSize:       5 << 20, // 5 MB
        AllowedExtensions: []string{".png", ".jpg", ".jpeg"},
        AllowedMIMETypes:  []string{"image/png", "image/jpeg"},
        MaxFiles:          1,
    },
})
```

## Multiple files

```go
type GalleryRequest struct {
    Files []*upload.File `form:"files" validate:"required,max=10"`
}

func uploadGallery(c *echo.Context, req GalleryRequest) (GalleryResponse, error) {
    for _, f := range req.Files {
        if err := f.SaveTo(filepath.Join("uploads", f.Filename)); err != nil {
            return GalleryResponse{}, echo.NewHTTPError(http.StatusInternalServerError, "save failed")
        }
    }
    // ...
}
```

## Notes

- `Route.FileConfig` enforces size/type/count limits before your handler runs;
  violations return 400 automatically.
- Validation tags (`required`, `max=`) still apply to the request struct.
- Scaffold with `echonext generate upload <name>`.
