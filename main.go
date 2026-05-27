package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

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

func main() {
	// Retrieve Proxmox VE connection details from environment variables for security
	apiUrl := os.Getenv("PVE_API_URL")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_SECRET")

	// Validate that all required environment variables are set
	if apiUrl == "" || tokenID == "" || secret == "" {
		fmt.Println("[ERROR] Missing required environment variables: PVE_API_URL, PVE_TOKEN_ID, or PVE_SECRET")
		return
	}

	fmt.Println("Initializing Proxmox VE API client...")
	client := NewPveClient(apiUrl, tokenID, secret)

	// Build the target URL for testing API connectivity (Get PVE Version)
	reqUrl := client.ApiUrl + "/version"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		fmt.Printf("Failed to create HTTP request: %v\n", err)
		return
	}

	// Inject PVE specific API Token authentication into HTTP Header
	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", client.TokenID, client.Secret)
	req.Header.Add("Authorization", authHeader)

	// Execute the HTTP request
	resp, err := client.HttpCli.Do(req)
	if err != nil {
		fmt.Printf("Failed to connect to PVE API: %v\n", err)
		return
	}
	// Ensure the response body is closed to prevent resource leaks
	defer resp.Body.Close()

	// Read and print the API response
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("HTTP Status Code: %d\n", resp.StatusCode)
	fmt.Printf("PVE API Raw Response: %s\n", string(body))
}
