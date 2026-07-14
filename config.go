package minego

import (
	"log/slog"
	"net/http"
	"time"

	pauth "github.com/zeozeozeo/minego/internal/protocol/auth"
)

type ResourcePackPolicy uint8

const (
	ResourcePackDecline ResourcePackPolicy = iota
	ResourcePackAccept
)

// Config controls a bot connection. Zero values receive practical defaults.
type Config struct {
	Address string
	// Version is a release name such as "26.2". An empty value probes the
	// server status endpoint and selects the matching compiled-in protocol.
	Version            string
	Auth               Authentication
	Locale             string
	ViewDistance       int8
	Logger             *slog.Logger
	DialTimeout        time.Duration
	ResourcePackPolicy ResourcePackPolicy
	Brand              string
	HTTPClient         *http.Client
	Plugins            []Plugin
}

type AuthMode uint8

const (
	AuthOffline AuthMode = iota
	AuthMicrosoft
)

// Authentication selects offline or Microsoft device-code login.
type Authentication struct {
	Mode       AuthMode
	Username   string
	ClientID   string
	TokenStore TokenStore
	DeviceCode func(DeviceCode)
}

func Offline(username string) Authentication {
	return Authentication{Mode: AuthOffline, Username: username}
}
func Microsoft(clientID, username string, callback func(DeviceCode)) Authentication {
	return Authentication{Mode: AuthMicrosoft, ClientID: clientID, Username: username, DeviceCode: callback}
}

type DeviceCode = pauth.DeviceCode
type TokenStore = pauth.TokenStore
type CachedSession = pauth.CachedSession

// MemoryTokenStore returns a concurrency-safe, non-persistent token store.
func MemoryTokenStore() TokenStore { return newMemoryStore() }
