// Package upload provides type-safe file upload handling with OpenAPI support
package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
)

// Default file upload limits
const (
	DefaultMaxFileSize  = 32 << 20 // 32MB
	DefaultMaxTotalSize = 64 << 20 // 64MB
	DefaultMaxFiles     = 10
)

// File represents an uploaded file with metadata
type File struct {
	// Filename is the original name of the uploaded file
	Filename string `json:"filename"`
	// Size is the file size in bytes
	Size int64 `json:"size"`
	// ContentType is the MIME type of the file
	ContentType string `json:"content_type"`
	// Header contains the MIME header of the file part
	Header textproto.MIMEHeader `json:"-"`
	// file is the underlying multipart file (not exported)
	file multipart.File
}

// Files is a convenience type for multiple file uploads
type Files []*File

// Open returns an io.ReadCloser for reading the file content
func (f *File) Open() (io.ReadCloser, error) {
	if f.file == nil {
		return nil, errors.New("file is not available")
	}
	// Seek to beginning in case it was already read
	if seeker, ok := f.file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek file: %w", err)
		}
	}
	return f.file, nil
}

// Read reads the entire file content into memory
func (f *File) Read() ([]byte, error) {
	reader, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// Close closes the underlying file
func (f *File) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

// Extension returns the file extension (with dot)
func (f *File) Extension() string {
	return filepath.Ext(f.Filename)
}

// Config configures file upload handling
type Config struct {
	// MaxFileSize is the maximum size for a single file (default: 32MB)
	MaxFileSize int64
	// MaxTotalSize is the maximum total size for all files (default: 64MB)
	MaxTotalSize int64
	// AllowedMIMETypes lists allowed MIME types (empty = all allowed)
	AllowedMIMETypes []string
	// AllowedExtensions lists allowed file extensions (e.g., ".jpg", ".pdf")
	AllowedExtensions []string
	// MaxFiles is the maximum number of files for multi-upload (default: 10)
	MaxFiles int
	// PreserveFile keeps the multipart.File open for later reading
	// If false (default), file content is read into memory immediately
	PreserveFile bool
}

// Error represents a file upload validation error
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Errors is a collection of file upload errors
type Errors []Error

func (e Errors) Error() string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

// DefaultConfig returns the default file upload configuration
func DefaultConfig() Config {
	return Config{
		MaxFileSize:  DefaultMaxFileSize,
		MaxTotalSize: DefaultMaxTotalSize,
		MaxFiles:     DefaultMaxFiles,
	}
}

// Bind binds multipart form data to a request struct containing File fields
func Bind(c echo.Context, req interface{}, config *Config) error {
	if config == nil {
		defaultConfig := DefaultConfig()
		config = &defaultConfig
	}

	// Set max memory for multipart parsing
	maxMemory := config.MaxTotalSize
	if maxMemory == 0 {
		maxMemory = DefaultMaxTotalSize
	}

	// Parse multipart form
	if err := c.Request().ParseMultipartForm(maxMemory); err != nil {
		return &Error{
			Field:   "",
			Message: fmt.Sprintf("failed to parse multipart form: %v", err),
		}
	}

	reqValue := reflect.ValueOf(req)
	if reqValue.Kind() == reflect.Ptr {
		reqValue = reqValue.Elem()
	}
	reqType := reqValue.Type()

	var errors Errors
	var totalSize int64

	// Iterate through struct fields
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		fieldValue := reqValue.Field(i)

		// Get form field name from tag
		formTag := field.Tag.Get("form")
		if formTag == "" {
			formTag = strings.ToLower(field.Name)
		}
		formTag = strings.Split(formTag, ",")[0] // Remove omitempty etc.

		// Check if this is a File field
		isFile := IsFileType(field.Type)
		isFileSlice := IsFileSliceType(field.Type)

		if !isFile && !isFileSlice {
			// Bind non-file form fields
			if err := bindFormField(c, fieldValue, formTag); err != nil {
				errors = append(errors, Error{
					Field:   formTag,
					Message: err.Error(),
				})
			}
			continue
		}

		// Handle file upload fields
		if isFile {
			// Single file upload
			file, header, err := c.Request().FormFile(formTag)
			if err != nil {
				if err == http.ErrMissingFile {
					// Check if field is required (handled by validator later)
					continue
				}
				errors = append(errors, Error{
					Field:   formTag,
					Message: fmt.Sprintf("failed to get file: %v", err),
				})
				continue
			}

			// Validate file
			uploadedFile, err := processFile(file, header, config)
			if err != nil {
				errors = append(errors, Error{
					Field:   formTag,
					Message: err.Error(),
				})
				continue
			}

			totalSize += uploadedFile.Size

			// Set field value
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue.Set(reflect.ValueOf(uploadedFile))
			} else {
				fieldValue.Set(reflect.ValueOf(*uploadedFile))
			}
		} else if isFileSlice {
			// Multiple file upload
			form := c.Request().MultipartForm
			if form == nil || form.File == nil {
				continue
			}

			headers := form.File[formTag]
			if len(headers) == 0 {
				// Try with brackets (common for multiple files)
				headers = form.File[formTag+"[]"]
			}

			if config.MaxFiles > 0 && len(headers) > config.MaxFiles {
				errors = append(errors, Error{
					Field:   formTag,
					Message: fmt.Sprintf("too many files: got %d, max %d", len(headers), config.MaxFiles),
				})
				continue
			}

			var uploads []*File
			for _, header := range headers {
				file, err := header.Open()
				if err != nil {
					errors = append(errors, Error{
						Field:   formTag,
						Message: fmt.Sprintf("failed to open file %s: %v", header.Filename, err),
					})
					continue
				}

				uploadedFile, err := processFile(file, header, config)
				if err != nil {
					errors = append(errors, Error{
						Field:   formTag,
						Message: fmt.Sprintf("file %s: %v", header.Filename, err),
					})
					continue
				}

				totalSize += uploadedFile.Size
				uploads = append(uploads, uploadedFile)
			}

			// Set slice field
			sliceValue := reflect.MakeSlice(fieldValue.Type(), len(uploads), len(uploads))
			for i, upload := range uploads {
				sliceValue.Index(i).Set(reflect.ValueOf(upload))
			}
			fieldValue.Set(sliceValue)
		}
	}

	// Check total size
	if config.MaxTotalSize > 0 && totalSize > config.MaxTotalSize {
		errors = append(errors, Error{
			Field:   "",
			Message: fmt.Sprintf("total file size %d exceeds maximum %d", totalSize, config.MaxTotalSize),
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// processFile creates a File from a multipart file with validation
func processFile(file multipart.File, header *multipart.FileHeader, config *Config) (*File, error) {
	// Validate file size
	if config.MaxFileSize > 0 && header.Size > config.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum %d", header.Size, config.MaxFileSize)
	}

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect content type from file content
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		contentType = http.DetectContentType(buffer[:n])
		// Seek back to beginning
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
	}

	// Validate MIME type
	if len(config.AllowedMIMETypes) > 0 {
		allowed := false
		for _, mime := range config.AllowedMIMETypes {
			if strings.EqualFold(contentType, mime) || strings.HasPrefix(contentType, mime) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("MIME type %s not allowed", contentType)
		}
	}

	// Validate extension
	ext := filepath.Ext(header.Filename)
	if len(config.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range config.AllowedExtensions {
			if strings.EqualFold(ext, allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("file extension %s not allowed", ext)
		}
	}

	return &File{
		Filename:    header.Filename,
		Size:        header.Size,
		ContentType: contentType,
		Header:      header.Header,
		file:        file,
	}, nil
}

// IsFileType checks if a type is *File or File
func IsFileType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name() == "File" && t.PkgPath() == "github.com/abdussamadbello/echonext/upload"
}

// IsFileSliceType checks if a type is []*File
func IsFileSliceType(t reflect.Type) bool {
	if t.Kind() != reflect.Slice {
		return false
	}
	elemType := t.Elem()
	return IsFileType(elemType)
}

// bindFormField binds a non-file form field
func bindFormField(c echo.Context, fieldValue reflect.Value, formTag string) error {
	formValue := c.FormValue(formTag)
	if formValue == "" {
		return nil
	}

	if !fieldValue.CanSet() {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(formValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if _, err := fmt.Sscanf(formValue, "%d", &i); err != nil {
			return fmt.Errorf("invalid integer value: %s", formValue)
		}
		fieldValue.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var i uint64
		if _, err := fmt.Sscanf(formValue, "%d", &i); err != nil {
			return fmt.Errorf("invalid unsigned integer value: %s", formValue)
		}
		fieldValue.SetUint(i)
	case reflect.Float32, reflect.Float64:
		var f float64
		if _, err := fmt.Sscanf(formValue, "%f", &f); err != nil {
			return fmt.Errorf("invalid float value: %s", formValue)
		}
		fieldValue.SetFloat(f)
	case reflect.Bool:
		fieldValue.SetBool(formValue == "true" || formValue == "1")
	}

	return nil
}

// IsMultipartRequest checks if the request is multipart/form-data
func IsMultipartRequest(c echo.Context) bool {
	contentType := c.Request().Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "multipart/form-data")
}

// HasFileFields checks if a type contains File fields
func HasFileFields(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if IsFileType(field.Type) || IsFileSliceType(field.Type) {
			return true
		}
	}
	return false
}

// GetFileFields returns information about File fields in a type
func GetFileFields(t reflect.Type) []FieldInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var fields []FieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if IsFileType(field.Type) || IsFileSliceType(field.Type) {
			formTag := field.Tag.Get("form")
			if formTag == "" {
				formTag = strings.ToLower(field.Name)
			}
			formTag = strings.Split(formTag, ",")[0]

			fields = append(fields, FieldInfo{
				Name:       field.Name,
				FormTag:    formTag,
				IsMultiple: IsFileSliceType(field.Type),
				Required:   strings.Contains(field.Tag.Get("validate"), "required"),
			})
		}
	}
	return fields
}

// FieldInfo contains information about a file upload field
type FieldInfo struct {
	Name       string
	FormTag    string
	IsMultiple bool
	Required   bool
}

// GetNonFileFields returns information about non-file form fields
func GetNonFileFields(t reflect.Type) []FormFieldInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var fields []FormFieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if IsFileType(field.Type) || IsFileSliceType(field.Type) {
			continue
		}

		formTag := field.Tag.Get("form")
		jsonTag := field.Tag.Get("json")
		if formTag == "" && jsonTag != "" {
			formTag = strings.Split(jsonTag, ",")[0]
		}
		if formTag == "" {
			formTag = strings.ToLower(field.Name)
		}
		formTag = strings.Split(formTag, ",")[0]

		fields = append(fields, FormFieldInfo{
			Name:       field.Name,
			FormTag:    formTag,
			Type:       field.Type,
			Required:   strings.Contains(field.Tag.Get("validate"), "required"),
			Validation: field.Tag.Get("validate"),
		})
	}
	return fields
}

// FormFieldInfo contains information about a form field
type FormFieldInfo struct {
	Name       string
	FormTag    string
	Type       reflect.Type
	Required   bool
	Validation string
}
