package avatar

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/abdussamadbello/echonext"
	"github.com/abdussamadbello/echonext/upload"
	"github.com/labstack/echo/v4"
)

// Config for avatar uploads
var Config = upload.Config{
	MaxFileSize:      10 << 20, // 10 MB
	MaxTotalSize:     50 << 20, // 50 MB
	AllowedMIMETypes: []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		// Add more MIME types as needed
	},
	AllowedExtensions: []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp",
	},
}

// UploadDir is the directory where files are stored
var UploadDir = "uploads/avatar"

// Handler handles avatar file uploads
func Handler(c echo.Context, req UploadRequest) (UploadResponse, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return UploadResponse{}, echo.NewHTTPError(500, "Failed to create upload directory")
	}

	// Process the uploaded file
	file := req.File

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFilename(file.Filename), ext)
	destPath := filepath.Join(UploadDir, newFilename)

	// Save the file
	if err := file.SaveTo(destPath); err != nil {
		return UploadResponse{}, echo.NewHTTPError(500, fmt.Sprintf("Failed to save file: %v", err))
	}

	return UploadResponse{
		Filename:    newFilename,
		OriginalName: file.Filename,
		Size:        file.Size,
		ContentType: file.ContentType,
		URL:         fmt.Sprintf("/%s/%s", UploadDir, newFilename),
	}, nil
}

// MultiHandler handles multiple avatar file uploads
func MultiHandler(c echo.Context, req MultiUploadRequest) (MultiUploadResponse, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return MultiUploadResponse{}, echo.NewHTTPError(500, "Failed to create upload directory")
	}

	var uploaded []FileInfo
	for _, file := range req.Files {
		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		newFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFilename(file.Filename), ext)
		destPath := filepath.Join(UploadDir, newFilename)

		// Save the file
		if err := file.SaveTo(destPath); err != nil {
			return MultiUploadResponse{}, echo.NewHTTPError(500, fmt.Sprintf("Failed to save file %s: %v", file.Filename, err))
		}

		uploaded = append(uploaded, FileInfo{
			Filename:    newFilename,
			OriginalName: file.Filename,
			Size:        file.Size,
			ContentType: file.ContentType,
			URL:         fmt.Sprintf("/%s/%s", UploadDir, newFilename),
		})
	}

	return MultiUploadResponse{
		Files:   uploaded,
		Count:   len(uploaded),
		Message: fmt.Sprintf("Successfully uploaded %d files", len(uploaded)),
	}, nil
}

// RegisterRoutes registers upload routes
func RegisterRoutes(app *echonext.App) {
	// Single file upload
	app.Upload("/avatar/upload", Handler, echonext.Route{
		Summary:     "Upload a avatar",
		Description: "Upload a single avatar file",
		Tags:        []string{"Avatar"},
		FileConfig:  &Config,
	})

	// Multiple file upload
	app.Upload("/avatar/upload-multiple", MultiHandler, echonext.Route{
		Summary:     "Upload multiple avatars",
		Description: "Upload multiple avatar files at once",
		Tags:        []string{"Avatar"},
		FileConfig:  &Config,
	})

	// Serve uploaded files
	app.Echo.Static("/uploads/avatar", UploadDir)
}

// sanitizeFilename removes potentially dangerous characters from filename
func sanitizeFilename(name string) string {
	// Remove extension and path components
	name = filepath.Base(name)
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]

	// Keep only alphanumeric and some safe characters
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result = append(result, r)
		}
	}

	if len(result) == 0 {
		return "file"
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return string(result)
}
