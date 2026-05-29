package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ==========================================
// 1. STRUCT & CONSTRUCTOR DEFINITIONS
// ==========================================

// PveClient stores connection details for Proxmox VE API
type PveClient struct {
	ApiUrl  string
	TokenID string
	Secret  string
	HttpCli *http.Client
}

// NewPveClient initializes and returns a new PveClient instance
func NewPveClient(apiUrl, tokenID, secret string) *PveClient {
	// InsecureSkipVerify: true is required for self-signed certificates in lab environments
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &PveClient{
		ApiUrl:  apiUrl,
		TokenID: tokenID,
		Secret:  secret,
		HttpCli: &http.Client{Transport: tr, Timeout: 10 * time.Second},
	}
}

// ==========================================
// 2. METHOD DEFINITION (The "Smart Button")
// ==========================================

// GetVersion connects to PVE API and returns status code, raw JSON string, and error.
// This method is bound to the *PveClient struct using a pointer receiver (c *PveClient).
func (c *PveClient) GetVersion() (int, string, error) {
	// Step A: Build the target URL using the struct's ApiUrl field
	reqUrl := c.ApiUrl + "/version"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		// Return default values and the formatted error if setup fails
		return 0, "", fmt.Errorf("failed to create request: %v", err)
	}

	// Step B: Inject PVE specific API Token authentication into HTTP Header
	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	// Step C: Execute the HTTP request using the client's internal http client
	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE API: %v", err)
	}
	// Crucial: Ensure the response body is closed to prevent system resource leaks
	defer resp.Body.Close()

	// Step D: Read the raw binary data flow from the server response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Step E: Return HTTP status code, raw JSON string, and nil (no error)
	return resp.StatusCode, string(body), nil
}

// ==========================================
// 3. MAIN EXECUTION ENTRYPOINT
// ==========================================

func main() {
	// Retrieve Proxmox VE connection details from environment variables for security
	apiUrl := os.Getenv("PVE_API_URL")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_SECRET")

	// Validate that all required environment variables are set before proceeding
	if apiUrl == "" || tokenID == "" || secret == "" {
		fmt.Println("[ERROR] Missing required environment variables: PVE_API_URL, PVE_TOKEN_ID, or PVE_SECRET")
		return
	}

	fmt.Println("Initializing Proxmox VE API client...")
	client := NewPveClient(apiUrl, tokenID, secret)

	fmt.Println("Fetching PVE node version via Method...")
	// Trigger the method and receive three variables (statusCode, rawJson, err)
	statusCode, rawJson, err := client.GetVersion()
	if err != nil {
		// If the "alarm light" is on (err != nil), print error and exit
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	// Output the final formatted results
	fmt.Printf("HTTP Status Code: %d\n", statusCode)
	fmt.Printf("PVE API Raw Response:\n%s\n", rawJson)
}
