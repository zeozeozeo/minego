package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeviceCodeFlow(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("client_id") != "client" {
			t.Errorf("client id = %q", r.Form.Get("client_id"))
		}
		_ = json.NewEncoder(w).Encode(msDeviceCodeResponse{DeviceCode: "device", UserCode: "ABCD", VerificationURI: "https://example.test", ExpiresIn: 10, Interval: 1})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(msTokenResponse{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 3600})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	oldDevice, oldToken := msDeviceCodeURL, msTokenURL
	msDeviceCodeURL, msTokenURL = server.URL+"/device", server.URL+"/token"
	defer func() { msDeviceCodeURL, msTokenURL = oldDevice, oldToken }()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	code, err := requestDeviceCode(ctx, server.Client(), "client", []string{"offline_access"})
	if err != nil {
		t.Fatal(err)
	}
	if code.UserCode != "ABCD" {
		t.Fatalf("code = %q", code.UserCode)
	}
	token, err := pollDeviceToken(ctx, server.Client(), "client", code)
	if err != nil {
		t.Fatal(err)
	}
	if token.RefreshToken != "refresh" {
		t.Fatalf("refresh = %q", token.RefreshToken)
	}
}
