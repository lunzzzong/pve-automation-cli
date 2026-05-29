package main

import (
	"crypto/tls"
	"encoding/json" // 💡 Added for JSON unmarshalling
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

// PveVersionData represents the inner fields of the PVE version response
type PveVersionData struct {
	Version string `json:"version"`
	Release string `json:"release"`
}

// PveVersionResponse represents the top-level structure of the PVE version API response
type PveVersionResponse struct {
	Data PveVersionData `json:"data"`
}

// NewPveClient initializes and returns a new PveClient instance
func NewPveClient(apiUrl, tokenID, secret string) *PveClient {
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
func (c *PveClient) GetVersion() (int, string, error) {
	reqUrl := c.ApiUrl + "/version"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %v", err)
	}

	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %v", err)
	}

	return resp.StatusCode, string(body), nil
}

// ==========================================
// 3. MAIN EXECUTION ENTRYPOINT
// ==========================================

func main() {
	apiUrl := os.Getenv("PVE_API_URL")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_SECRET")

	if apiUrl == "" || tokenID == "" || secret == "" {
		fmt.Println("[ERROR] Missing required environment variables: PVE_API_URL, PVE_TOKEN_ID, or PVE_SECRET")
		return
	}

	fmt.Println("Initializing Proxmox VE API client...")
	client := NewPveClient(apiUrl, tokenID, secret)

	fmt.Println("Fetching PVE node version via Method...")
	statusCode, rawJson, err := client.GetVersion()
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	// Milestone 3: Parse the raw JSON string into Go struct
	var versionResp PveVersionResponse
	err = json.Unmarshal([]byte(rawJson), &versionResp)
	if err != nil {
		fmt.Printf("[ERROR] Failed to parse JSON: %v\n", err)
		return
	}

	// Output the beautifully formatted results
	fmt.Printf("HTTP Status Code: %d\n", statusCode)
	fmt.Println("\n[Success] Connected to Proxmox VE successfully!")
	fmt.Printf("-> PVE Version: %s\n", versionResp.Data.Version)
	fmt.Printf("-> PVE Release: %s\n", versionResp.Data.Release)
}
