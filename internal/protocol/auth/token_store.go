package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CachedSession represents a cached authentication session.
type CachedSession struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	UUID         string    `json:"uuid"`
	Username     string    `json:"username"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// IsValid returns true if the session has a valid access token that hasn't expired.
// Uses a 5-minute buffer to avoid edge cases.
func (s *CachedSession) IsValid() bool {
	if s == nil || s.AccessToken == "" {
		return false
	}
	return time.Now().Add(5 * time.Minute).Before(s.ExpiresAt)
}

// TokenStore defines an interface to persist and retrieve sessions for multiple accounts.
type TokenStore interface {
	SaveSession(session *CachedSession) error
	LoadSession(username string) (*CachedSession, error)
	Clear(username string) error
	ListAccounts() ([]string, error)
}

// TokenStoreConfig holds configuration for creating a TokenStore.
type TokenStoreConfig struct {
	// Path is the file path for the credentials cache.
	// If nil: uses default path "~/.mclib/credentials_cache.json"
	// If points to empty string: use in-memory store (no persistence)
	// If points to a path: use that path
	Path *string
}

// fileTokenStore persists tokens as JSON in a file with 0600 permissions.
// It supports multiple accounts, keyed by username.
type fileTokenStore struct {
	filePath string
}

// NewTokenStore creates a new TokenStore based on the provided configuration.
// If config.Path is nil, uses the default path: ~/.mclib/credentials_cache.json
// If config.Path points to empty string, returns an in-memory store (no persistence).
// If config.Path points to a path, uses that path.
func NewTokenStore(config TokenStoreConfig) (TokenStore, error) {
	var path string

	if config.Path == nil {
		// default
		var err error
		path, err = getDefaultCredentialsFilePath()
		if err != nil {
			return nil, err
		}
	} else if *config.Path == "" {
		// empty string means in-memory store
		return newMemoryTokenStore(), nil
	} else {
		// custom path
		path = *config.Path
	}

	// ensure directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// test if path is writable
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("credentials path is not writable: %w", err)
	}
	_ = f.Close()

	return &fileTokenStore{filePath: path}, nil
}

func (s *fileTokenStore) SaveSession(session *CachedSession) error {
	if session == nil || session.Username == "" {
		return errors.New("session with username is required")
	}

	sessions, err := s.loadSessions()
	if err != nil {
		return err
	}

	sessions[session.Username] = session
	return s.saveSessions(sessions)
}

func (s *fileTokenStore) LoadSession(username string) (*CachedSession, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	sessions, err := s.loadSessions()
	if err != nil {
		return nil, err
	}

	return sessions[username], nil
}

func (s *fileTokenStore) Clear(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	sessions, err := s.loadSessions()
	if err != nil {
		return err
	}

	delete(sessions, username)

	// if no accounts remain, remove the file
	if len(sessions) == 0 {
		return os.Remove(s.filePath)
	}

	return s.saveSessions(sessions)
}

func (s *fileTokenStore) ListAccounts() ([]string, error) {
	sessions, err := s.loadSessions()
	if err != nil {
		return nil, err
	}

	accounts := make([]string, 0, len(sessions))
	for username := range sessions {
		accounts = append(accounts, username)
	}

	return accounts, nil
}

func (s *fileTokenStore) loadSessions() (map[string]*CachedSession, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]*CachedSession), nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return make(map[string]*CachedSession), nil
	}
	data, err = unprotect(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	var sessions map[string]*CachedSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		// try to migrate from old format (map[string]string with just refresh tokens)
		var oldTokens map[string]string
		if jsonErr := json.Unmarshal(data, &oldTokens); jsonErr == nil {
			sessions = make(map[string]*CachedSession)
			for username, refreshToken := range oldTokens {
				sessions[username] = &CachedSession{
					RefreshToken: refreshToken,
					Username:     username,
				}
			}
			return sessions, nil
		}
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	if sessions == nil {
		sessions = make(map[string]*CachedSession)
	}

	return sessions, nil
}

func (s *fileTokenStore) saveSessions(sessions map[string]*CachedSession) error {
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	data, err = protect(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.filePath)
}

// memoryTokenStore is an in-memory implementation of TokenStore.
// It does not persist sessions to disk.
type memoryTokenStore struct {
	sessions map[string]*CachedSession
}

// newMemoryTokenStore creates a new in-memory token store.
func newMemoryTokenStore() TokenStore {
	return &memoryTokenStore{
		sessions: make(map[string]*CachedSession),
	}
}

func (m *memoryTokenStore) SaveSession(session *CachedSession) error {
	if session == nil || session.Username == "" {
		return errors.New("session with username is required")
	}
	m.sessions[session.Username] = session
	return nil
}

func (m *memoryTokenStore) LoadSession(username string) (*CachedSession, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	return m.sessions[username], nil
}

func (m *memoryTokenStore) Clear(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}
	delete(m.sessions, username)
	return nil
}

func (m *memoryTokenStore) ListAccounts() ([]string, error) {
	accounts := make([]string, 0, len(m.sessions))
	for username := range m.sessions {
		accounts = append(accounts, username)
	}
	return accounts, nil
}

func getDefaultCredentialsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".minego", "credentials_cache.json"), nil
}
