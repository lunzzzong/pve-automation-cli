package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// ==========================================
// 1. DATA STRUCTS & CUSTOM ERROR
// ==========================================

// PveApiError wraps HTTP fault boundaries for advanced diagnostic state introspection
type PveApiError struct {
	StatusCode int
	Message    string
}

// Error converts the internal telemetry tracking state into a standard error string payload
func (e *PveApiError) Error() string {
	return fmt.Sprintf("PVE API Fault (Status: %d): %s", e.StatusCode, e.Message)
}

// PveClient abstracts the cluster transport pipeline connection parameters
type PveClient struct {
	ApiUrl  string
	TokenID string
	Secret  string
	HttpCli *http.Client
}

type PveNodeData struct {
	Node   string `json:"node"`
	Status string `json:"status"`
}

type PveNodesResponse struct {
	Data []PveNodeData `json:"data"`
}

type PveNodeMemory struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
}

type PveNodeStatusData struct {
	Cpu    float64       `json:"cpu"`
	MaxCpu int           `json:"cpus"`
	Memory PveNodeMemory `json:"memory"`
	Uptime int64         `json:"uptime"`
}

type PveNodeStatusResponse struct {
	Data PveNodeStatusData `json:"data"`
}

type PveVmData struct {
	Vmid   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

type PveVmsResponse struct {
	Data []PveVmData `json:"data"`
}

// ==========================================
// 2. CONSTRUCTORS & METHOD DEFINITIONS
// ==========================================

// NewPveClient acts as a low-level constructor allocating a custom transport pipeline with TLS bypass
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

// NewPveClientFromEnv automates environment configuration parsing with native boundary sanity checks
func NewPveClientFromEnv() (*PveClient, error) {
	_ = godotenv.Load()
	apiUrl := os.Getenv("PVE_API_URL")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_SECRET")

	// Defensive engineering: Validate that all critical identity variables are present
	if apiUrl == "" || tokenID == "" || secret == "" {
		return nil, fmt.Errorf("missing required environment variables: PVE_API_URL, PVE_TOKEN_ID, or PVE_SECRET")
	}

	// Dynamic delegation: Propagate verified configs to the underlying connection constructor
	return NewPveClient(apiUrl, tokenID, secret), nil
}

// GetNodes fetches the cluster top-level machine nodes index map securely
func (c *PveClient) GetNodes() (int, string, error) {
	reqUrl := c.ApiUrl + "/nodes"
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

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, "", &PveApiError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %v", err)
	}
	return resp.StatusCode, string(body), nil
}

// GetNodeStatus tracks real-time compute hardware consumption metrics on specific targets
func (c *PveClient) GetNodeStatus(nodeName string) (int, string, error) {
	reqUrl := fmt.Sprintf("%s/nodes/%s/status", c.ApiUrl, nodeName)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request for node %s: %v", nodeName, err)
	}

	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE node status API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read node status response body: %v", err)
	}
	return resp.StatusCode, string(body), nil
}

// GetVms evaluates system hypervisor compute resource assets distributed inside the infrastructure
func (c *PveClient) GetVms() (int, string, error) {
	reqUrl := c.ApiUrl + "/cluster/resources"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request for VMs: %v", err)
	}

	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE resource API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read resource response body: %v", err)
	}
	return resp.StatusCode, string(body), nil
}
