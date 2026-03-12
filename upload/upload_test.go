package upload

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test request/response types for file uploads
type SingleFileRequest struct {
	File *File `form:"file" validate:"required"`
}

type MultiFileRequest struct {
	Files []*File `form:"files" validate:"required,min=1"`
}

type MixedUploadRequest struct {
	Title       string  `form:"title" validate:"required"`
	Description string  `form:"description"`
	File        *File   `form:"file" validate:"required"`
	Attachments []*File `form:"attachments"`
}

// Helper to create multipart request
func createMultipartRequest(t *testing.T, fields map[string]string, files map[string][]byte) (*http.Request, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, value := range fields {
		err := writer.WriteField(key, value)
		require.NoError(t, err)
	}

	// Add files
	for fieldName, content := range files {
		part, err := writer.CreateFormFile(fieldName, fieldName+".txt")
		require.NoError(t, err)
		_, err = part.Write(content)
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, writer.FormDataContentType()
}

// Helper to create multipart request with multiple files for same field
func createMultiFileRequest(t *testing.T, fieldName string, files [][]byte) (*http.Request, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for i, content := range files {
		filename := fieldName + "_" + string(rune('a'+i)) + ".txt"
		part, err := writer.CreateFormFile(fieldName, filename)
		require.NoError(t, err)
		_, err = part.Write(content)
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, writer.FormDataContentType()
}

func TestBind_SingleFile(t *testing.T) {
	e := echo.New()

	req, _ := createMultipartRequest(t, nil, map[string][]byte{
		"file": []byte("Hello, World!"),
	})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var uploadReq SingleFileRequest
	err := Bind(c, &uploadReq, nil)
	require.NoError(t, err)

	assert.NotNil(t, uploadReq.File)
	assert.Equal(t, "file.txt", uploadReq.File.Filename)
	assert.Equal(t, int64(13), uploadReq.File.Size)
}

func TestBind_MultipleFiles(t *testing.T) {
	e := echo.New()

	req, _ := createMultiFileRequest(t, "files", [][]byte{
		[]byte("File 1 content"),
		[]byte("File 2 content"),
		[]byte("File 3 content"),
	})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var uploadReq MultiFileRequest
	err := Bind(c, &uploadReq, nil)
	require.NoError(t, err)

	assert.Len(t, uploadReq.Files, 3)
}

func TestBind_MixedFormAndFiles(t *testing.T) {
	e := echo.New()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	writer.WriteField("title", "My Document")
	writer.WriteField("description", "A test document")

	// Add main file
	part, _ := writer.CreateFormFile("file", "document.pdf")
	part.Write([]byte("PDF content here"))

	// Add attachments
	att1, _ := writer.CreateFormFile("attachments", "image1.png")
	att1.Write([]byte("PNG1"))
	att2, _ := writer.CreateFormFile("attachments", "image2.png")
	att2.Write([]byte("PNG2"))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var uploadReq MixedUploadRequest
	err := Bind(c, &uploadReq, nil)
	require.NoError(t, err)

	assert.Equal(t, "My Document", uploadReq.Title)
	assert.Equal(t, "A test document", uploadReq.Description)
	assert.NotNil(t, uploadReq.File)
	assert.Equal(t, "document.pdf", uploadReq.File.Filename)
	assert.Len(t, uploadReq.Attachments, 2)
}

func TestBind_MaxFileSizeExceeded(t *testing.T) {
	e := echo.New()

	req, _ := createMultipartRequest(t, nil, map[string][]byte{
		"file": []byte("This content is definitely longer than 10 bytes"),
	})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config := &Config{
		MaxFileSize: 10,
	}

	var uploadReq SingleFileRequest
	err := Bind(c, &uploadReq, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

func TestBind_AllowedMIMETypes(t *testing.T) {
	e := echo.New()

	req, _ := createMultipartRequest(t, nil, map[string][]byte{
		"file": []byte("plain text content"),
	})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config := &Config{
		AllowedMIMETypes: []string{"image/png", "image/jpeg"},
	}

	var uploadReq SingleFileRequest
	err := Bind(c, &uploadReq, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MIME type")
}

func TestBind_AllowedExtensions(t *testing.T) {
	e := echo.New()

	req, _ := createMultipartRequest(t, nil, map[string][]byte{
		"file": []byte("text content"),
	})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config := &Config{
		AllowedExtensions: []string{".pdf"},
	}

	var uploadReq SingleFileRequest
	err := Bind(c, &uploadReq, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "extension")
}

func TestBind_MaxFiles(t *testing.T) {
	e := echo.New()

	req, _ := createMultiFileRequest(t, "files", [][]byte{
		[]byte("File 1"),
		[]byte("File 2"),
		[]byte("File 3"),
	})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config := &Config{
		MaxFiles: 2,
	}

	var uploadReq MultiFileRequest
	err := Bind(c, &uploadReq, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many files")
}

func TestFile_Extension(t *testing.T) {
	f := &File{Filename: "document.pdf"}
	assert.Equal(t, ".pdf", f.Extension())

	f2 := &File{Filename: "archive.tar.gz"}
	assert.Equal(t, ".gz", f2.Extension())

	f3 := &File{Filename: "noextension"}
	assert.Equal(t, "", f3.Extension())
}

func TestError_String(t *testing.T) {
	err := &Error{
		Field:   "file",
		Message: "file too large",
	}
	assert.Equal(t, "file: file too large", err.Error())
}

func TestErrors_String(t *testing.T) {
	errs := Errors{
		{Field: "file1", Message: "too large"},
		{Field: "file2", Message: "invalid type"},
	}
	assert.Equal(t, "file1: too large; file2: invalid type", errs.Error())
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, int64(DefaultMaxFileSize), config.MaxFileSize)
	assert.Equal(t, int64(DefaultMaxTotalSize), config.MaxTotalSize)
	assert.Equal(t, DefaultMaxFiles, config.MaxFiles)
	assert.Empty(t, config.AllowedMIMETypes)
	assert.Empty(t, config.AllowedExtensions)
}

func TestIsMultipartRequest(t *testing.T) {
	e := echo.New()

	// Test multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	c := e.NewContext(req, httptest.NewRecorder())
	assert.True(t, IsMultipartRequest(c))

	// Test JSON request
	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req2.Header.Set("Content-Type", "application/json")

	c2 := e.NewContext(req2, httptest.NewRecorder())
	assert.False(t, IsMultipartRequest(c2))
}

func TestHasFileFields(t *testing.T) {
	// Type with file upload field
	type WithFile struct {
		File *File `form:"file"`
	}
	assert.True(t, HasFileFields(reflect.TypeOf(WithFile{})))

	// Type with file upload slice
	type WithFiles struct {
		Files []*File `form:"files"`
	}
	assert.True(t, HasFileFields(reflect.TypeOf(WithFiles{})))

	// Type without file upload
	type NoFile struct {
		Name string `form:"name"`
	}
	assert.False(t, HasFileFields(reflect.TypeOf(NoFile{})))
}

func TestFile_Open_ReturnsError_WhenNilFile(t *testing.T) {
	f := &File{
		Filename: "test.txt",
		file:     nil,
	}

	reader, err := f.Open()
	assert.Error(t, err)
	assert.Nil(t, reader)
	assert.Contains(t, err.Error(), "file is not available")
}

func TestFile_Close_ReturnsNil_WhenNilFile(t *testing.T) {
	f := &File{
		Filename: "test.txt",
		file:     nil,
	}

	err := f.Close()
	assert.NoError(t, err)
}

// MockReadCloser for testing
type mockReadCloser struct {
	*bytes.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestFile_OpenAndRead(t *testing.T) {
	content := []byte("test file content")
	f := &File{
		Filename:    "test.txt",
		Size:        int64(len(content)),
		ContentType: "text/plain",
		file:        &mockReadCloser{bytes.NewReader(content)},
	}

	// Test Open
	reader, err := f.Open()
	require.NoError(t, err)
	defer reader.Close()

	// Read content
	readContent, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, readContent)
}

func TestIsFileType(t *testing.T) {
	// Test File type
	assert.True(t, IsFileType(reflect.TypeOf(&File{})))
	assert.True(t, IsFileType(reflect.TypeOf(File{})))

	// Test non-File type
	assert.False(t, IsFileType(reflect.TypeOf("")))
	assert.False(t, IsFileType(reflect.TypeOf(struct{}{})))
}

func TestIsFileSliceType(t *testing.T) {
	// Test []*File type
	assert.True(t, IsFileSliceType(reflect.TypeOf([]*File{})))

	// Test non-slice types
	assert.False(t, IsFileSliceType(reflect.TypeOf(&File{})))
	assert.False(t, IsFileSliceType(reflect.TypeOf([]string{})))
}

func TestGetFileFields(t *testing.T) {
	type TestStruct struct {
		File        *File   `form:"file" validate:"required"`
		Attachments []*File `form:"attachments"`
		Name        string  `form:"name"`
	}

	fields := GetFileFields(reflect.TypeOf(TestStruct{}))
	assert.Len(t, fields, 2)

	assert.Equal(t, "File", fields[0].Name)
	assert.Equal(t, "file", fields[0].FormTag)
	assert.False(t, fields[0].IsMultiple)
	assert.True(t, fields[0].Required)

	assert.Equal(t, "Attachments", fields[1].Name)
	assert.Equal(t, "attachments", fields[1].FormTag)
	assert.True(t, fields[1].IsMultiple)
	assert.False(t, fields[1].Required)
}

func TestGetNonFileFields(t *testing.T) {
	type TestStruct struct {
		File *File  `form:"file"`
		Name string `form:"name" validate:"required"`
		Age  int    `form:"age"`
	}

	fields := GetNonFileFields(reflect.TypeOf(TestStruct{}))
	assert.Len(t, fields, 2)

	assert.Equal(t, "Name", fields[0].Name)
	assert.Equal(t, "name", fields[0].FormTag)
	assert.True(t, fields[0].Required)

	assert.Equal(t, "Age", fields[1].Name)
	assert.Equal(t, "age", fields[1].FormTag)
	assert.False(t, fields[1].Required)
}

// Benchmark tests
func BenchmarkBind_SingleFile(b *testing.B) {
	e := echo.New()
	content := bytes.Repeat([]byte("x"), 1024) // 1KB file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write(content)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var uploadReq SingleFileRequest
		Bind(c, &uploadReq, nil)
	}
}
