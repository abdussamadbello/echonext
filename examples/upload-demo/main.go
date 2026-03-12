package main

import (
	"html/template"
	"log"

	"github.com/abdussamadbello/echonext"
	uploadAvatar "github.com/abdussamadbello/echonext/examples/upload-demo/internal/upload/avatar"
	"github.com/labstack/echo/v5"
)

func main() {
	app := echonext.New()

	// Register upload routes using the generated handler
	uploadAvatar.RegisterRoutes(app)

	// Serve a simple upload UI
	app.Echo.GET("/", serveHome)

	// Health check
	app.GET("/health", healthCheck, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"Health"},
	})

	// OpenAPI spec and docs
	app.Echo.GET("/api/spec", func(c echo.Context) error {
		return c.JSON(200, app.GenerateOpenAPISpec())
	})
	app.ServeSwaggerUI("/api/docs", "/api/spec")

	log.Println("Server starting on :8080")
	log.Println("Upload UI: http://localhost:8080")
	log.Println("Swagger UI: http://localhost:8080/api/docs")

	if err := app.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

func healthCheck(c echo.Context) (HealthResponse, error) {
	return HealthResponse{Status: "healthy"}, nil
}

func serveHome(c echo.Context) error {
	tmpl := template.Must(template.New("upload").Parse(uploadHTML))
	return tmpl.Execute(c.Response(), nil)
}

const uploadHTML = `<!DOCTYPE html>
<html>
<head>
    <title>EchoNext File Upload Demo</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .upload-form { border: 2px dashed #ccc; padding: 40px; text-align: center; margin: 20px 0; border-radius: 10px; }
        .upload-form:hover { border-color: #4CAF50; }
        input[type="file"] { margin: 10px 0; }
        button { background: #4CAF50; color: white; padding: 10px 20px; border: none; cursor: pointer; border-radius: 5px; }
        button:hover { background: #45a049; }
        .result { margin-top: 20px; padding: 15px; border-radius: 5px; }
        .success { background: #e8f5e9; border: 1px solid #4CAF50; }
        .error { background: #ffebee; border: 1px solid #f44336; }
        .preview { max-width: 200px; margin-top: 10px; border-radius: 5px; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>EchoNext File Upload Demo</h1>

    <div class="info">
        <strong>Allowed file types:</strong> JPEG, PNG, GIF, WebP<br>
        <strong>Max file size:</strong> 10 MB<br>
        <a href="/api/docs" target="_blank">View API Documentation</a>
    </div>

    <h2>Single File Upload</h2>
    <div class="upload-form">
        <form id="singleForm">
            <input type="file" name="file" accept="image/*" required />
            <br><br>
            <button type="submit">Upload Single File</button>
        </form>
    </div>
    <div id="singleResult"></div>

    <h2>Multiple File Upload</h2>
    <div class="upload-form">
        <form id="multiForm">
            <input type="file" name="files" accept="image/*" multiple required />
            <br><br>
            <button type="submit">Upload Multiple Files</button>
        </form>
    </div>
    <div id="multiResult"></div>

    <script>
        document.getElementById('singleForm').onsubmit = async function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            try {
                const response = await fetch('/avatar/upload', {
                    method: 'POST',
                    body: formData
                });
                const result = await response.json();

                const div = document.getElementById('singleResult');
                if (result.success) {
                    div.className = 'result success';
                    div.innerHTML = '<strong>Upload successful!</strong><br>' +
                        'Filename: ' + result.data.filename + '<br>' +
                        'Size: ' + (result.data.size / 1024).toFixed(2) + ' KB<br>' +
                        '<img src="' + result.data.url + '" class="preview" />';
                } else {
                    div.className = 'result error';
                    div.innerHTML = '<strong>Upload failed:</strong> ' + result.error;
                }
            } catch (err) {
                document.getElementById('singleResult').className = 'result error';
                document.getElementById('singleResult').innerHTML = '<strong>Error:</strong> ' + err.message;
            }
        };

        document.getElementById('multiForm').onsubmit = async function(e) {
            e.preventDefault();
            const formData = new FormData(this);

            try {
                const response = await fetch('/avatar/upload-multiple', {
                    method: 'POST',
                    body: formData
                });
                const result = await response.json();

                const div = document.getElementById('multiResult');
                if (result.success) {
                    div.className = 'result success';
                    let html = '<strong>Upload successful!</strong><br>' +
                        'Files uploaded: ' + result.data.count + '<br><br>';

                    result.data.files.forEach(file => {
                        html += file.original_name + ' (' + (file.size / 1024).toFixed(2) + ' KB)<br>';
                        html += '<img src="' + file.url + '" class="preview" /><br><br>';
                    });
                    div.innerHTML = html;
                } else {
                    div.className = 'result error';
                    div.innerHTML = '<strong>Upload failed:</strong> ' + result.error;
                }
            } catch (err) {
                document.getElementById('multiResult').className = 'result error';
                document.getElementById('multiResult').innerHTML = '<strong>Error:</strong> ' + err.message;
            }
        };
    </script>
</body>
</html>
`
