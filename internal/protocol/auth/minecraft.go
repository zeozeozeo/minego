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
	mcAuthLoginWithXboxURL = "https://api.minecraftservices.com/authentication/login_with_xbox"
	mcEntitlementsURL      = "https://api.minecraftservices.com/entitlements/mcstore"
	mcProfileURL           = "https://api.minecraftservices.com/minecraft/profile"
)

type minecraftLoginRequest struct {
	IdentityToken string `json:"identityToken"`
}

type minecraftLoginResponse struct {
	Username    string `json:"username"`
	Roles       []any  `json:"roles"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func minecraftLoginWithXbox(ctx context.Context, httpClient *http.Client, userHash, xstsToken string) (*minecraftLoginResponse, error) {
	body := minecraftLoginRequest{IdentityToken: fmt.Sprintf("XBL3.0 x=%s;%s", userHash, xstsToken)}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mcAuthLoginWithXboxURL, strings.NewReader(string(buf)))
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
		return nil, fmt.Errorf("minecraft login_with_xbox failed: %s: %s", res.Status, string(data))
	}
	var out minecraftLoginResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

type entitlementsResponse struct {
	Items []struct {
		Name      string `json:"name"`
		Signature string `json:"signature"`
	} `json:"items"`
	Signature string `json:"signature"`
	KeyID     string `json:"keyId"`
}

func checkGameOwnership(ctx context.Context, httpClient *http.Client, mcAccessToken string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mcEntitlementsURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+mcAccessToken)

	res, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("entitlements request failed: %s: %s", res.Status, string(data))
	}
	var out entitlementsResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return false, err
	}
	if len(out.Items) == 0 {
		return false, nil
	}

	// according to wiki, expect product_minecraft and game_minecraft
	// https://minecraft.wiki/w/Microsoft_authentication#Checking_game_ownership
	var hasProduct, hasGame bool
	for _, it := range out.Items {
		switch it.Name {
		case "product_minecraft":
			hasProduct = true
		case "game_minecraft":
			hasGame = true
		}
	}

	return hasProduct && hasGame, nil
}

type profile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Skins any    `json:"skins"`
	Capes any    `json:"capes"`
}

func fetchMinecraftProfile(ctx context.Context, httpClient *http.Client, mcAccessToken string) (*profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mcProfileURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+mcAccessToken)

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("profile request failed: %s: %s", res.Status, string(data))
	}

	var out profile
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
