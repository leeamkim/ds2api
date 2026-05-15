// Package synology provides a client for interacting with the Synology Download Station API.
package synology

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client holds the configuration and HTTP client for communicating with a Synology NAS.
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	SessionID  string
	httpClient *http.Client
}

// APIResponse is the generic wrapper returned by the Synology API.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

// APIError represents an error returned by the Synology API.
type APIError struct {
	Code int `json:"code"`
}

// LoginData holds the session ID returned after a successful login.
type LoginData struct {
	SID string `json:"sid"`
}

// NewClient creates a new Synology API client.
// If skipTLSVerify is true, TLS certificate verification is disabled (useful for self-signed certs).
func NewClient(baseURL, username, password string, skipTLSVerify bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify, //nolint:gosec // intentional, controlled by config
		},
	}

	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// Login authenticates with the Synology API and stores the session ID.
func (c *Client) Login() error {
	params := url.Values{}
	params.Set("api", "SYNO.API.Auth")
	params.Set("version", "3")
	params.Set("method", "login")
	params.Set("account", c.Username)
	params.Set("passwd", c.Password)
	params.Set("session", "DownloadStation")
	params.Set("format", "sid")

	resp, err := c.get("/webapi/auth.cgi", params)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	var loginData LoginData
	if err := json.Unmarshal(resp.Data, &loginData); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if loginData.SID == "" {
		return fmt.Errorf("login succeeded but no session ID was returned")
	}

	c.SessionID = loginData.SID
	return nil
}

// Logout invalidates the current session.
func (c *Client) Logout() error {
	params := url.Values{}
	params.Set("api", "SYNO.API.Auth")
	params.Set("version", "1")
	params.Set("method", "logout")
	params.Set("session", "DownloadStation")

	_, err := c.get("/webapi/auth.cgi", params)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}

	c.SessionID = ""
	return nil
}

// get performs a GET request to the given path with the provided query parameters.
// It automatically appends the session ID if one is available.
func (c *Client) get(path string, params url.Values) (*APIResponse, error) {
	if c.SessionID != "" {
		params.Set("_sid", c.SessionID)
	}

	rawURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	resp, err := c.httpClient.Get(rawURL) //nolint:noctx // context support to be added
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal api response: %w", err)
	}

	if !apiResp.Success {
		code := 0
		if apiResp.Error != nil {
			code = apiResp.Error.Code
		}
		return nil, fmt.Errorf("api returned error code: %d", code)
	}

	return &apiResp, nil
}
