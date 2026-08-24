package feed

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/version"

	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
)

// HandleTestIMAPConnection tests IMAP connection settings
// @Summary      Test IMAP connection
// @Description  Test IMAP server connection with provided credentials
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "IMAP connection test (email_imap_server, email_imap_port, email_username, email_password, email_folder)"
// @Success      200  {object}  map[string]string  "Connection successful (message)"
// @Failure      400  {object}  map[string]string  "Bad request (missing required fields)"
// @Failure      401  {object}  map[string]string  "Authentication failed"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /email/imap/test [post]
func HandleTestIMAPConnection(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("[IMAP Test] Handler called, method: %s", r.Method)

	if r.Method != http.MethodPost {
		log.Printf("[IMAP Test] Method not allowed: %s", r.Method)
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IMAPServer string `json:"email_imap_server"`
		IMAPPort   int    `json:"email_imap_port"`
		Username   string `json:"email_username"`
		Password   string `json:"email_password"`
		Folder     string `json:"email_folder"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[IMAP Test] JSON decode error: %v", err)
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	log.Printf("[IMAP Test] Request received: server=%s, port=%d, username=%s, folder=%s",
		req.IMAPServer, req.IMAPPort, req.Username, req.Folder)

	// Validate required fields
	if req.IMAPServer == "" || req.Username == "" || req.Password == "" {
		response.JSON(w, map[string]string{"error": "IMAP server, username, and password are required"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Set default port if not specified
	if req.IMAPPort == 0 {
		req.IMAPPort = 993
	}

	// Set default folder if not specified
	if req.Folder == "" {
		req.Folder = "INBOX"
	}

	// Try to connect to IMAP server
	server := req.IMAPServer
	if req.IMAPPort != 0 {
		server = fmt.Sprintf("%s:%d", req.IMAPServer, req.IMAPPort)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         req.IMAPServer,
	}

	// Try TLS first
	log.Printf("[IMAP Test] Attempting TLS connection to %s", server)
	c, err := client.DialTLS(server, tlsConfig)
	if err != nil {
		log.Printf("[IMAP Test] TLS failed, trying non-TLS: %v", err)
		// Fallback to non-TLS
		c, err = client.Dial(server)
		if err != nil {
			log.Printf("[IMAP Test] Connection failed: %v", err)
			response.JSON(w, map[string]string{"error": "Failed to connect to IMAP server: " + err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	defer c.Logout()
	log.Printf("[IMAP Test] Connected successfully")

	// Send ID command before login (RFC 2971)
	// This is required by some email providers like NetEase (163, 126)
	sendIMAPIDCommand(c)

	// Login
	log.Printf("[IMAP Test] Attempting login for user: %s", req.Username)
	if err := c.Login(req.Username, req.Password); err != nil {
		log.Printf("[IMAP Test] Login failed: %v", err)
		response.JSON(w, map[string]string{"error": "Authentication failed: " + err.Error()})
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[IMAP Test] Login successful")

	// Try to select the folder
	log.Printf("[IMAP Test] Selecting folder: %s", req.Folder)
	_, err = c.Select(req.Folder, false)
	if err != nil {
		log.Printf("[IMAP Test] Folder selection failed: %v", err)
		response.JSON(w, map[string]string{"error": "Failed to select folder '" + req.Folder + "': " + err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("[IMAP Test] Folder selected successfully")

	// Success!
	log.Printf("[IMAP Test] All checks passed!")
	response.JSON(w, map[string]string{"message": "Connection successful!"})
}

// sendIMAPIDCommand sends the IMAP ID command to identify the client
// This is required by some email providers (e.g., NetEase 163/126) as per RFC 2971
func sendIMAPIDCommand(c *client.Client) {
	idClient := id.NewClient(c)

	// Check if server supports ID extension
	supported, err := idClient.SupportID()
	if err != nil || !supported {
		// Server doesn't support ID extension, skip
		log.Printf("[IMAP ID] Server does not support ID extension")
		return
	}

	// Send client identification
	clientID := id.ID{
		id.FieldName:    "MrRSS",
		id.FieldVersion: version.Version,
		id.FieldVendor:  "MrRSS",
		id.FieldOS:      runtime.GOOS,
	}

	serverID, err := idClient.ID(clientID)
	if err != nil {
		log.Printf("[IMAP ID] Failed to send ID command: %v", err)
		return
	}
	log.Printf("[IMAP ID] Server ID: %v", serverID)
}
