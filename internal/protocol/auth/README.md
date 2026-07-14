# Auth

This package implements the full Microsoft Authentication flow for Minecraft: Java Edition, allowing clients to obtain valid Minecraft access tokens for joining online-mode servers.

## Overview

Since Minecraft 1.16, Mojang accounts have been migrated to Microsoft accounts. Authentication now requires a multi-step process involving Microsoft OAuth2, Xbox Live, XSTS, and Minecraft services.

## Authentication Flow

```plain
┌────────────────────────────────────────────────────────────────────────────┐
│                         Microsoft Authentication                           │
└────────────────────────────────────────────────────────────────────────────┘

┌─────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  User   │     │  Microsoft   │     │  Xbox Live   │     │    XSTS      │
│ Browser │     │   OAuth2     │     │    (XBL)     │     │   Service    │
└────┬────┘     └──────┬───────┘     └──────┬───────┘     └──────┬───────┘
     │                 │                    │                    │
     │  Login Page     │                    │                    │
     │◄────────────────│                    │                    │
     │                 │                    │                    │
     │  Auth Code      │                    │                    │
     │────────────────►│                    │                    │
     │                 │                    │                    │
     │           ┌─────┴─────┐              │                    │
     │           │ Exchange  │              │                    │
     │           │   Code    │              │                    │
     │           └─────┬─────┘              │                    │
     │                 │                    │                    │
     │                 │  MS Access Token   │                    │
     │                 │───────────────────►│                    │
     │                 │                    │                    │
     │                 │              ┌─────┴─────┐              │
     │                 │              │    XBL    │              │
     │                 │              │   Auth    │              │
     │                 │              └─────┬─────┘              │
     │                 │                    │                    │
     │                 │                    │    XBL Token       │
     │                 │                    │───────────────────►│
     │                 │                    │                    │
     │                 │                    │              ┌─────┴─────┐
     │                 │                    │              │   XSTS    │
     │                 │                    │              │ Authorize │
     │                 │                    │              └─────┬─────┘
     │                 │                    │                    │
     │                 │                    │                    ▼
     │                 │                    │              XSTS Token
     │                 │                    │                    │
     └─────────────────┴────────────────────┴────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         Minecraft Authentication                             │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Minecraft   │     │ Entitlements │     │   Profile     │
│   Service    │     │    Check     │     │   Fetch      │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       │  XSTS Token        │                    │
       │◄───────────────────┘                    │
       │                    │                    │
 ┌─────┴─────┐              │                    │
 │  Login    │              │                    │
 │ with Xbox │              │                    │
 └─────┬─────┘              │                    │
       │                    │                    │
       │  MC Access Token   │                    │
       │───────────────────►│                    │
       │                    │                    │
       │              ┌─────┴─────┐              │
       │              │  Verify   │              │
       │              │ Ownership │              │
       │              └─────┬─────┘              │
       │                    │                    │
       │                    │───────────────────►│
       │                    │                    │
       │                    │              ┌─────┴─────┐
       │                    │              │   Get     │
       │                    │              │  Profile   │
       │                    │              └─────┬─────┘
       │                    │                    │
       │                    │                    ▼
       │                    │              UUID + Username
       │                    │                    │
       └────────────────────┴────────────────────┘
```

## Authentication Steps

1. **Microsoft OAuth2** - User logs in via browser, app receives authorization code
2. **Token Exchange** - Code exchanged for Microsoft access token and refresh token
3. **Xbox Live (XBL)** - Microsoft token exchanged for Xbox Live token
4. **XSTS Authorization** - XBL token authorized for Minecraft services
5. **Minecraft Login** - XSTS token exchanged for Minecraft access token
6. **Entitlement Check** - Verify the account owns Minecraft
7. **Profile Fetch** - Retrieve player UUID, username and access token to join servers

## Usage

### Basic Login (Interactive)

```go
client := auth.NewClient(auth.AuthClientConfig{
    ClientID: "your-azure-app-client-id",
})

ctx := context.Background()
loginData, err := client.Login(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Logged in as %s (UUID: %s)\n", loginData.Username, loginData.UUID)
// use loginData.AccessToken to join servers
```

### Login with Session Caching

The `Login()` method automatically handles full session caching, including the Minecraft access token:

1. **Check cached session** - If a valid session exists (access token not expired), returns immediately with zero API calls
2. **Refresh if expired** - If session expired but refresh token exists, refreshes through the full auth chain
3. **Interactive fallback** - If no cache or refresh fails, opens browser for interactive login
4. **Save session** - Caches the full session (access token, refresh token, UUID, username, expiry) for future use

The Minecraft access token is valid for **24 hours**, so rapid start/stop cycles won't trigger re-authentication.

```go
client := auth.NewClient(auth.AuthClientConfig{
    ClientID: "your-azure-app-client-id",
    Username: "PlayerName", // specify which cached account to use
})

// first call: may need interactive login or token refresh
loginData, err := client.Login(ctx)

// subsequent calls within 24 hours: returns immediately from cache
loginData, err = client.Login(ctx)
```

### Login with Refresh Token (Non-Interactive)

If you already have a refresh token:

```go
client := auth.NewClient(auth.AuthClientConfig{
    ClientID: "your-azure-app-client-id",
})

loginData, err := client.LoginWithRefreshToken(ctx, refreshToken)
if err != nil {
    // token expired or invalid, need interactive login
}
```

### Custom Token Storage

By default, tokens are stored in `~/.mclib/credentials_cache.json`. You can customize this:

```go
// custom file path
customPath := filepath.Join(os.Getenv("HOME"), "my_credentials_cache.json")
client := auth.NewClient(auth.AuthClientConfig{
    ClientID: "your-client-id",
    TokenStoreConfig: auth.TokenStoreConfig{
        Path: &customPath,
    },
})

// in-memory only (no persistence)
client := auth.NewClient(auth.AuthClientConfig{
    ClientID: "your-client-id",
    TokenStoreConfig: auth.TokenStoreConfig{
        Path: nil,
    },
})

// custom implementation (e.g. if you want to use database or redis instead, you could??)
// but in 99 % cases, the default file-based store is sufficient
client := auth.NewClient(auth.AuthClientConfig{
    ClientID:   "your-client-id",
    TokenStore: myCustomStore, // implements auth.TokenStore
})
```

### Managing Multiple Accounts

```go
// list cached accounts
accounts, _ := client.ListCachedAccounts()
for _, username := range accounts {
    fmt.Println(username)
}

// switch account
client.SetUsername("OtherPlayer")
loginData, _ := client.Login(ctx)

// clear cached token for current username/account
client.ClearCachedToken(ctx)
```

### Fetching Mojang Certificates

For chat message signing (1.19+), you need Mojang-issued certificates. To do this, you need to have the access token of the account you want to sign messages for.

```go
certData, err := auth.FetchMojangCertificate(loginData.AccessToken)
if err != nil {
    log.Fatal(err)
}

// certData.PrivateKey - RSA private key for signing
// certData.PublicKey - RSA public key
// certData.PublicKeyBytes - DER-encoded public key
// certData.SignatureBytes - raw bytes of Mojang's signature of your public key
// certData.ExpiryTime - when the certificate expires
```

## Configuration

| Field | Description | Default |
| ----- | ------------ | ------ |
| `ClientID` | Azure AD application client ID | Required |
| `RedirectPort` | Local port for OAuth callback | Random available port |
| `Scopes` | OAuth scopes to request | `["XboxLive.signin", "offline_access"]` |
| `HTTPClient` | Custom HTTP client | 20s timeout client |
| `TokenStore` | Custom token storage | File-based store |
| `TokenStoreConfig` | Config for default store | `~/.mclib/credentials_cache.json` |
| `Username` | Account username for caching | From login response |

## API Endpoints Used

| Service | Endpoint | Purpose |
| ------- | -------- | ------- |
| Microsoft | `https://login.live.com/oauth20_authorize.srf` | OAuth authorization |
| Microsoft | `https://login.live.com/oauth20_token.srf` | Token exchange/refresh |
| Xbox Live | `https://user.auth.xboxlive.com/user/authenticate` | XBL authentication |
| XSTS | `https://xsts.auth.xboxlive.com/xsts/authorize` | XSTS authorization |
| Minecraft | `https://api.minecraftservices.com/authentication/login_with_xbox` | MC token |
| Minecraft | `https://api.minecraftservices.com/entitlements/mcstore` | Ownership check |
| Minecraft | `https://api.minecraftservices.com/minecraft/profile` | Profile data |
| Minecraft | `https://api.minecraftservices.com/player/certificates` | Chat signing certs |

## Creating an Azure Application

To use this package, you need an Azure AD application:

1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to Azure Active Directory → App registrations
3. Click "New registration"
4. Set redirect URI to `http://127.0.0.1` (Mobile and desktop applications)
5. Under "API permissions", add `XboxLive.signin` and `offline_access`
6. Copy the "Application (client) ID" - this is your `ClientID`

## Session Caching

The auth package caches full sessions to disk, including:

- **Access token** - The Minecraft JWT used to join servers
- **Refresh token** - Long-lived token for obtaining new access tokens
- **UUID** - Player's unique identifier
- **Username** - Player's display name
- **Expiry time** - When the access token expires (24 hours from issuance)

Sessions are automatically migrated from the old format (refresh token only) to the new format on first load.

## Security Notes

- Sessions are stored with `0600` permissions (owner read/write only)
- Tokens are saved atomically using temp file + rename
- Access tokens expire after 24 hours; refresh tokens are long-lived (90+ days)
- A 5-minute buffer is used when checking expiry to avoid edge cases
- The `offline_access` scope is required for refresh token functionality

## References

- [Microsoft Authentication Scheme - Minecraft Wiki](https://minecraft.wiki/w/Microsoft_authentication)
- [Mojang API - Minecraft Wiki](https://minecraft.wiki/w/Mojang_API)
