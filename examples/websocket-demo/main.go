package main

import (
	"html/template"
	"log"

	"github.com/abdussamadbello/echonext"
	"github.com/abdussamadbello/echonext/examples/websocket-demo/internal/ws/chat"
	"github.com/labstack/echo/v4"
)

func main() {
	app := echonext.New()

	// Create WebSocket handler
	chatHandler := chat.NewHandler()

	// Register WebSocket endpoint using EchoNext's WS method
	app.WS("/ws/chat", chatHandler)

	// Serve a simple chat UI
	app.Echo.GET("/", serveHome)

	// Health check
	app.GET("/health", healthCheck, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"Health"},
	})

	log.Println("Server starting on :8080")
	log.Println("Chat UI: http://localhost:8080")
	log.Println("WebSocket: ws://localhost:8080/ws/chat")

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
	tmpl := template.Must(template.New("chat").Parse(chatHTML))
	return tmpl.Execute(c.Response(), nil)
}

const chatHTML = `<!DOCTYPE html>
<html>
<head>
    <title>EchoNext WebSocket Chat Demo</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        #messages { border: 1px solid #ccc; height: 300px; overflow-y: scroll; padding: 10px; margin-bottom: 10px; }
        .message { margin: 5px 0; padding: 5px; border-radius: 5px; }
        .system { background: #e3f2fd; color: #1565c0; }
        .error { background: #ffebee; color: #c62828; }
        .user { background: #e8f5e9; }
        #input { width: calc(100% - 80px); padding: 10px; }
        #send { width: 70px; padding: 10px; }
        .status { margin-bottom: 10px; padding: 5px; border-radius: 3px; }
        .connected { background: #c8e6c9; }
        .disconnected { background: #ffcdd2; }
    </style>
</head>
<body>
    <h1>EchoNext WebSocket Chat Demo</h1>
    <div id="status" class="status disconnected">Disconnected</div>
    <div id="messages"></div>
    <div>
        <input type="text" id="input" placeholder="Type a message..." />
        <button id="send">Send</button>
    </div>

    <script>
        const messagesDiv = document.getElementById('messages');
        const input = document.getElementById('input');
        const sendBtn = document.getElementById('send');
        const statusDiv = document.getElementById('status');

        let ws;

        function connect() {
            ws = new WebSocket('ws://' + window.location.host + '/ws/chat');

            ws.onopen = function() {
                statusDiv.textContent = 'Connected';
                statusDiv.className = 'status connected';
            };

            ws.onmessage = function(e) {
                const msg = JSON.parse(e.data);
                const div = document.createElement('div');
                div.className = 'message ' + msg.type;

                if (msg.sender) {
                    div.textContent = msg.sender + ': ' + msg.payload;
                } else {
                    div.textContent = msg.payload;
                }

                messagesDiv.appendChild(div);
                messagesDiv.scrollTop = messagesDiv.scrollHeight;
            };

            ws.onclose = function() {
                statusDiv.textContent = 'Disconnected - Reconnecting...';
                statusDiv.className = 'status disconnected';
                setTimeout(connect, 2000);
            };

            ws.onerror = function(err) {
                console.error('WebSocket error:', err);
            };
        }

        function sendMessage() {
            if (input.value && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    type: 'message',
                    content: input.value
                }));
                input.value = '';
            }
        }

        sendBtn.onclick = sendMessage;
        input.onkeypress = function(e) {
            if (e.key === 'Enter') sendMessage();
        };

        connect();
    </script>
</body>
</html>
`
