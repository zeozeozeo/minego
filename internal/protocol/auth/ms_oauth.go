package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	msAuthorizeURL  = "https://login.live.com/oauth20_authorize.srf"
	msDeviceCodeURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode"
	msTokenURL      = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"
)

func buildAuthorizeURL(clientID, redirectURL string, scopes []string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURL)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("prompt", "select_account")
	return msAuthorizeURL + "?" + q.Encode()
}

func startLocalServer(port int, codeCh chan<- string) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		values := r.URL.Query()
		code := values.Get("code")
		if code == "" {
			_, _ = io.WriteString(w, "Cannot authenticate.")
			select {
			case codeCh <- "": // no code
			default: // prevent deadlock when channel is full
			}
			return
		}
		_, _ = io.WriteString(w, "You may now close this page.")
		select {
		case codeCh <- code: // send code
		default:
		}
	})

	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		_ = srv.Serve(ln)
	}()
	return srv, nil
}

func stopLocalServer(srv *http.Server) error {
	if srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

type msTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type msDeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

func requestDeviceCode(ctx context.Context, httpClient *http.Client, clientID string, scopes []string) (*msDeviceCodeResponse, error) {
	form := url.Values{"client_id": {clientID}, "scope": {strings.Join(scopes, " ")}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, msDeviceCodeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("device authorization failed: %s: %s", res.Status, data)
	}
	var out msDeviceCodeResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.DeviceCode == "" || out.UserCode == "" {
		return nil, errors.New("device authorization returned an incomplete response")
	}
	return &out, nil
}

func pollDeviceToken(ctx context.Context, httpClient *http.Client, clientID string, code *msDeviceCodeResponse) (*msTokenResponse, error) {
	interval := time.Duration(code.Interval) * time.Second
	if interval < time.Second {
		interval = 5 * time.Second
	}
	deadline := time.NewTimer(time.Duration(code.ExpiresIn) * time.Second)
	defer deadline.Stop()
	for {
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-deadline.C:
			timer.Stop()
			return nil, errors.New("device code expired")
		case <-timer.C:
		}
		form := url.Values{"client_id": {clientID}, "device_code": {code.DeviceCode}, "grant_type": {"urn:ietf:params:oauth:grant-type:device_code"}}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, msTokenURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		var body struct {
			msTokenResponse
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			return &body.msTokenResponse, nil
		}
		switch body.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
			continue
		case "authorization_declined":
			return nil, errors.New("device authorization declined")
		case "expired_token", "bad_verification_code":
			return nil, errors.New("device code expired")
		default:
			return nil, fmt.Errorf("device token exchange failed: %s", body.Description)
		}
	}
}

func exchangeAuthCodeForTokens(ctx context.Context, httpClient *http.Client, clientID, redirectURL, code string) (*msTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, msTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("token exchange failed: %s: %s", res.Status, string(data))
	}

	var tr msTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}

	return &tr, nil
}

func refreshAccessToken(ctx context.Context, httpClient *http.Client, clientID, redirectURL, refreshToken string) (*msTokenResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("empty refresh token")
	}
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")
	if redirectURL != "" {
		form.Set("redirect_uri", redirectURL)
	}
	form.Set("scope", "XboxLive.signin offline_access")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, msTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("refresh token request failed: %s: %s", res.Status, string(data))
	}

	var tr msTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}

	return &tr, nil
}
