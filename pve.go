package main

import (
	"crypto/tls"
	"encoding/json"
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
	Vmid   int     `json:"vmid"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	Type   string  `json:"type"`
	CPU    float64 `json:"cpu"`
}

type PveVmsResponse struct {
	Data []PveVmData `json:"data"`
}

// REFACTORED: TitleCase naming standard alignment to bridge main.go interfaces
type VMList []PveVmData

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

// Len returns the cardinality bounds dynamically for index calculations
func (v VMList) Len() int {
	return len(v)
}

// Swap exchanges elements out-of-order in-place securely using native tuple decoupling
// MODIFIED: Capitalized 'Swap' to implement sort.Interface properly
func (v VMList) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Less evaluates ordering parameters to achieve descending risk allocation metrics
// MODIFIED: Capitalized 'Less' and switched field pointer to the verified uppercase .CPU
func (v VMList) Less(i, j int) bool {
	return v[i].CPU > v[j].CPU
}

// FetchAndParseVms acts as a high-level domain orchestrator that encapsulates
// the entire raw network transport layer and JSON unmarshalling guardrails.
// It abstracts away the raw string payload and returns a clean, decoupled VMList slice.
func (c *PveClient) FetchAndParseVms() (VMList, error) {
	// 1. Dispatch the underlying low-level resource resource endpoint request
	statusCode, rawJson, err := c.GetVms()
	if err != nil {
		// Using %w to implement modern error wrapping for deep telemetry tracing
		return nil, fmt.Errorf("failed to fetch cluster resources: %w", err)
	}

	// 2. Assert HTTP transport status boundaries securely
	if statusCode != http.StatusOK {
		return nil, &PveApiError{
			StatusCode: statusCode,
			Message:    "PVE cluster infrastructure returned an unexpected status code",
		}
	}

	// 3. Allocate a structured memory buffer acting as the target data receiver
	var vmsResp PveVmsResponse

	// 4. Invoke the unmarshal worker by passing the memory address pointer (&vmsResp)
	// This dynamically decodes the raw byte slice into the allocated struct fields
	err = json.Unmarshal([]byte(rawJson), &vmsResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal internal PVE VMs data payload: %w", err)
	}

	// 5. Extract and propagate the cleanly parsed data slice up to the main logic pipeline
	return vmsResp.Data, nil
}
