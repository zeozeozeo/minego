package session_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeozeozeo/minego/internal/protocol/crypto"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// SessionServerClient represents a session server client
type SessionServerClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewSessionServerClient creates a new session server client
func NewSessionServerClient() *SessionServerClient {
	return &SessionServerClient{
		baseURL: "https://sessionserver.mojang.com",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// JoinRequest represents the request payload for /session/minecraft/join
type JoinRequest struct {
	AccessToken     string `json:"accessToken"`
	SelectedProfile string `json:"selectedProfile"`
	ServerID        string `json:"serverId"`
}

// HasJoinedRequest represents the request for /session/minecraft/hasJoined
type HasJoinedRequest struct {
	Username string `json:"username"`
	ServerID string `json:"serverId"`
	IP       string `json:"ip,omitempty"`
}

// HasJoinedResponse represents the response from /session/minecraft/hasJoined
type HasJoinedResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Properties []Property `json:"properties"`
}

// Property represents a profile property
type Property struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Signature string `json:"signature,omitempty"`
}

// ErrorResponse represents an error response from Mojang
type ErrorResponse struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Path         string `json:"path,omitempty"`
}

func (e ErrorResponse) String() string {
	if e.ErrorMessage != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Error, e.ErrorMessage, e.Path)
	}
	return fmt.Sprintf("%s (path: %s)", e.Error, e.Path)
}

// NewClientWithURL creates a new session server client with a custom base URL
func NewClientWithURL(baseURL string) *SessionServerClient {
	return &SessionServerClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Join authenticates a client session with the session server
func (c *SessionServerClient) Join(accessToken, selectedProfile, serverID string, sharedSecret, publicKey []byte) error {
	if !ValidateAccessToken(accessToken) {
		return fmt.Errorf("invalid access token format")
	}
	if !ns.ValidateUUID(selectedProfile) {
		return fmt.Errorf("invalid selectedProfile UUID format: %s", selectedProfile)
	}

	serverHash := ComputeServerHash(serverID, sharedSecret, publicKey)
	joinReq := JoinRequest{
		AccessToken:     accessToken,
		SelectedProfile: selectedProfile,
		ServerID:        serverHash,
	}

	jsonData, err := json.Marshal(joinReq)
	if err != nil {
		return fmt.Errorf("failed to marshal join request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/session/minecraft/join", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create join request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gomc-lib/protocol (github.com/zeozeozeo/minego/internal/protocol)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send join request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == 204 {
		// success
		return nil
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("authentication failed: %s (status %d)", string(body), resp.StatusCode)
	}

	return fmt.Errorf("authentication failed: %s (status %d)", errResp.String(), resp.StatusCode)
}

// HasJoined checks if a user has joined a server
func (c *SessionServerClient) HasJoined(username, serverID string, ip ...string) (*HasJoinedResponse, error) {
	url := fmt.Sprintf("%s/session/minecraft/hasJoined?username=%s&serverId=%s",
		c.baseURL, username, serverID)

	if len(ip) > 0 && ip[0] != "" {
		url += "&ip=" + ip[0]
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create hasJoined request: %w", err)
	}
	req.Header.Set("User-Agent", "gomc-lib/protocol (github.com/zeozeozeo/minego/internal/protocol)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send hasJoined request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == 204 {
		// user hasn't joined or session expired
		return nil, nil
	}

	if resp.StatusCode != 200 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("hasJoined failed: %s (status %d)", string(body), resp.StatusCode)
		}
		return nil, fmt.Errorf("hasJoined failed: %s (status %d)", errResp.String(), resp.StatusCode)
	}

	var hasJoinedResp HasJoinedResponse
	if err := json.Unmarshal(body, &hasJoinedResp); err != nil {
		return nil, fmt.Errorf("failed to parse hasJoined response: %w", err)
	}

	return &hasJoinedResp, nil
}

func ComputeServerHash(serverID string, sharedSecret, publicKey []byte) string {
	hasher := crypto.NewMinecraftSHA1()

	hasher.Write([]byte(serverID))
	hasher.Write(sharedSecret)
	hasher.Write(publicKey)

	return hasher.HexDigest()
}

func ValidateAccessToken(token string) bool {
	return len(token) > 10 && len(token) < 2048
}
