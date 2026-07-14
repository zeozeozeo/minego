package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	xblUserAuthenticateURL = "https://user.auth.xboxlive.com/user/authenticate"
	xstsAuthorizeURL       = "https://xsts.auth.xboxlive.com/xsts/authorize"
)

type xblRequest struct {
	Properties struct {
		AuthMethod string `json:"AuthMethod"`
		SiteName   string `json:"SiteName"`
		RpsTicket  string `json:"RpsTicket"`
	} `json:"Properties"`
	RelyingParty string `json:"RelyingParty"`
	TokenType    string `json:"TokenType"`
}

type xblResponse struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		XUI []struct {
			UHS string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}

func xblAuthenticate(ctx context.Context, httpClient *http.Client, msAccessToken string) (*xblResponse, error) {
	body := xblRequest{
		RelyingParty: "http://auth.xboxlive.com",
		TokenType:    "JWT",
	}
	body.Properties.AuthMethod = "RPS"
	body.Properties.SiteName = "user.auth.xboxlive.com"
	body.Properties.RpsTicket = "d=" + msAccessToken
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, xblUserAuthenticateURL, strings.NewReader(string(buf)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("xbl authenticate failed: %s: %s", res.Status, string(data))
	}

	var out xblResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

type xstsRequest struct {
	Properties struct {
		SandboxID  string   `json:"SandboxId"`
		UserTokens []string `json:"UserTokens"`
	} `json:"Properties"`
	RelyingParty string `json:"RelyingParty"`
	TokenType    string `json:"TokenType"`
}

type xstsResponse struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		XUI []struct {
			UHS string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}

func xstsAuthorize(ctx context.Context, httpClient *http.Client, xblToken string) (*xstsResponse, error) {
	body := xstsRequest{
		RelyingParty: "rp://api.minecraftservices.com/",
		TokenType:    "JWT",
	}
	body.Properties.SandboxID = "RETAIL"
	body.Properties.UserTokens = []string{xblToken}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, xstsAuthorizeURL, strings.NewReader(string(buf)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("xsts authorize failed: %s: %s", res.Status, string(data))
	}

	var out xstsResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
