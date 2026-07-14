package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// AuthClient is the main struct for the auth client
type AuthClient struct {
	cfg        AuthClientConfig
	httpClient *http.Client
	tokenStore TokenStore
	username   string
}

// AuthClientConfig is the configuration for the auth client
type AuthClientConfig struct {
	ClientID         string
	RedirectPort     int
	Scopes           []string
	HTTPClient       *http.Client
	TokenStore       TokenStore
	TokenStoreConfig TokenStoreConfig
	Username         string
	DeviceCode       func(DeviceCode)
}

// DeviceCode contains the instructions displayed during Microsoft device login.
type DeviceCode struct {
	UserCode        string
	VerificationURI string
	Message         string
	ExpiresAt       time.Time
}

// LoginData is the data returned from a login
type LoginData struct {
	AccessToken  string
	RefreshToken string
	UUID         string
	Username     string
	ExpiresAt    time.Time
}

// ToSession converts LoginData to a CachedSession for storage.
func (d LoginData) ToSession() *CachedSession {
	return &CachedSession{
		AccessToken:  d.AccessToken,
		RefreshToken: d.RefreshToken,
		UUID:         d.UUID,
		Username:     d.Username,
		ExpiresAt:    d.ExpiresAt,
	}
}

// FromSession creates LoginData from a CachedSession.
func FromSession(s *CachedSession) LoginData {
	return LoginData{
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		UUID:         s.UUID,
		Username:     s.Username,
		ExpiresAt:    s.ExpiresAt,
	}
}

// NewClient creates a new AuthClient with the given configuration
func NewClient(cfg AuthClientConfig) *AuthClient {
	if cfg.RedirectPort == 0 {
		cfg.RedirectPort = tryPort()
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"XboxLive.signin", "offline_access"}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}

	var store TokenStore
	if cfg.TokenStore != nil {
		store = cfg.TokenStore
	} else {
		// try to create store from config
		if s, err := NewTokenStore(cfg.TokenStoreConfig); err == nil {
			store = s
		}
	}

	return &AuthClient{cfg: cfg, httpClient: httpClient, tokenStore: store, username: cfg.Username}
}

// AuthorizeDevice starts Microsoft's device-code flow and waits for approval.
func (c *AuthClient) AuthorizeDevice(ctx context.Context) (string, error) {
	if c.cfg.ClientID == "" {
		return "", errors.New("missing client_id in Config")
	}
	code, err := requestDeviceCode(ctx, c.httpClient, c.cfg.ClientID, c.cfg.Scopes)
	if err != nil {
		return "", err
	}
	if c.cfg.DeviceCode == nil {
		return "", errors.New("device-code callback is required")
	}
	c.cfg.DeviceCode(DeviceCode{UserCode: code.UserCode, VerificationURI: code.VerificationURI, Message: code.Message, ExpiresAt: time.Now().Add(time.Duration(code.ExpiresIn) * time.Second)})
	token, err := pollDeviceToken(ctx, c.httpClient, c.cfg.ClientID, code)
	if err != nil {
		return "", err
	}
	if token.RefreshToken == "" {
		return "", errors.New("no refresh_token in token response")
	}
	return token.RefreshToken, nil
}

// AuthorizeWithLocalServer is retained for source compatibility and now uses
// the headless-friendly device-code flow.
func (c *AuthClient) AuthorizeWithLocalServer(ctx context.Context) (string, error) {
	return c.AuthorizeDevice(ctx)
}

// LoginWithRefreshToken performs a login with a refresh token.
func (c *AuthClient) LoginWithRefreshToken(ctx context.Context, refreshToken string) (LoginData, error) {
	if c.cfg.ClientID == "" {
		return LoginData{}, errors.New("missing client_id in Config")
	}
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", c.cfg.RedirectPort)

	// refresh Microsoft access token
	tokRes, err := refreshAccessToken(ctx, c.httpClient, c.cfg.ClientID, redirectURL, refreshToken)
	if err != nil {
		return LoginData{}, err
	}
	msAccessToken := tokRes.AccessToken
	refreshToken = tokRes.RefreshToken

	// XBL authenticate
	xblRes, err := xblAuthenticate(ctx, c.httpClient, msAccessToken)
	if err != nil {
		return LoginData{}, err
	}

	// XSTS authorize
	xstsRes, err := xstsAuthorize(ctx, c.httpClient, xblRes.Token)
	if err != nil {
		return LoginData{}, err
	}

	// Minecraft login with Xbox
	mcAuth, err := minecraftLoginWithXbox(ctx, c.httpClient, xblRes.DisplayClaims.XUI[0].UHS, xstsRes.Token)
	if err != nil {
		return LoginData{}, err
	}

	// verify entitlements
	owns, err := checkGameOwnership(ctx, c.httpClient, mcAuth.AccessToken)
	if err != nil {
		return LoginData{}, err
	}
	if !owns {
		return LoginData{}, errors.New("account does not own Minecraft (no entitlements)")
	}

	// fetch profile
	profile, err := fetchMinecraftProfile(ctx, c.httpClient, mcAuth.AccessToken)
	if err != nil {
		return LoginData{}, err
	}
	if profile == nil || profile.ID == "" {
		return LoginData{}, errors.New("minecraft profile not found for account")
	}

	// calculate expiry time from expires_in (seconds)
	expiresAt := time.Now().Add(time.Duration(mcAuth.ExpiresIn) * time.Second)

	return LoginData{
		AccessToken:  mcAuth.AccessToken,
		RefreshToken: refreshToken,
		UUID:         profile.ID,
		Username:     profile.Name,
		ExpiresAt:    expiresAt,
	}, nil
}

// Login performs a cached login. It first checks for a valid cached session
// (with unexpired access token). If no valid session exists, it attempts to
// refresh using the stored refresh token. If that fails, it falls back to
// interactive auth via a local HTTP callback and browser.
//
// The username for caching is determined by:
// 1. The Username field in AuthClientConfig if set
// 2. The username from the LoginData response (after successful login)
func (c *AuthClient) Login(ctx context.Context) (LoginData, error) {
	store := c.tokenStore

	// try cached session if username is specified
	if store != nil && c.username != "" {
		if session, err := store.LoadSession(c.username); err == nil && session != nil {
			// check if session is still valid (access token not expired)
			if session.IsValid() {
				c.username = session.Username
				return FromSession(session), nil
			}

			// session expired but we have a refresh token - try to refresh
			if session.RefreshToken != "" {
				if data, err := c.LoginWithRefreshToken(ctx, session.RefreshToken); err == nil {
					c.username = data.Username
					_ = store.SaveSession(data.ToSession())
					return data, nil
				}
			}
		}
	}

	// no cache or cache failed; must reauthenticate
	rt, err := c.AuthorizeDevice(ctx)
	if err != nil {
		return LoginData{}, err
	}

	data, err := c.LoginWithRefreshToken(ctx, rt)
	if err != nil {
		return LoginData{}, err
	}

	// save session with the username from login response
	c.username = data.Username
	if store != nil {
		_ = store.SaveSession(data.ToSession())
	}

	return data, nil
}

// ClearCachedToken removes the stored refresh token for the current username.
// If username is not set, this returns an error.
func (c *AuthClient) ClearCachedToken(_ context.Context) error {
	if c.tokenStore == nil {
		return nil
	}

	if c.username == "" {
		return errors.New("no username set, cannot clear cached token")
	}

	return c.tokenStore.Clear(c.username)
}

// SetUsername updates the username for this client. This affects which cached
// token is loaded/saved.
func (c *AuthClient) SetUsername(username string) {
	c.username = username
}

// GetUsername returns the current username for this client.
func (c *AuthClient) GetUsername() string {
	return c.username
}

// ListCachedAccounts returns a list of all usernames that have cached tokens.
func (c *AuthClient) ListCachedAccounts() ([]string, error) {
	if c.tokenStore == nil {
		return nil, nil
	}
	return c.tokenStore.ListAccounts()
}

// tryPort tries to find an open port
func tryPort() int {
	randomPort := rand.Intn(65535-1024) + 1024

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", randomPort))
	if err != nil {
		return tryPort()
	}
	defer listener.Close()

	return randomPort
}
