# Session Server

This package implements a client for Mojang's Session Server API, which is used to authenticate players during the Minecraft: Java Edition login process.

## Overview

When a player joins an online-mode Minecraft server, authentication happens in three steps:

1. **Client authenticates with Mojang** - The client calls `/session/minecraft/join` to register its intent to join a specific server
2. **Server verifies the client** - The server calls `/session/minecraft/hasJoined` to confirm the client actually authenticated
3. **Connection proceeds** - If verification succeeds, the server allows the client to continue connecting

This prevents unauthorized players from connecting with stolen usernames and ensures account ownership.

## Authentication Flow

```plain
┌────────┐                  ┌────────┐                  ┌─────────────────┐
│ Client │                  │ Server │                  │ Session Server  │
└───┬────┘                  └───┬────┘                  └────────┬────────┘
    │                           │                                │
    │◄──Encryption Request──────│                                │
    │   (Server ID, Public Key, │                                │
    │    Verify Token)          │                                │
    │                           │                                │
    │  Generate Shared Secret   │                                │
    │  Compute Server Hash      │                                │
    │                           │                                │
    │──POST /join──────────────────────────────────────────────► │
    │  {accessToken,            │                                │
    │   selectedProfile,         │                                │
    │   serverId: hash}         │                                │
    │                           │                                │
    │◄─────────────────────────────────────────────204 No Content│
    │                           │                                │
    │──Encryption Response─────►│                                │
    │  (Encrypted Secret,       │                                │
    │   Encrypted Token)        │                                │
    │                           │                                │
    │                           │  Compute Server Hash           │
    │                           │                                │
    │                           │──GET /hasJoined───────────────►│
    │                           │  ?username=...&serverId=hash   │
    │                           │                                │
    │                           │◄──────────────Player Profile─── │
    │                           │  {id, name, properties}        │
    │                           │                                │
    │◄──Login Success───────────│                                │
    │                           │                                │
```

## Server Hash Computation

Minecraft uses a non-standard SHA-1 hexdigest algorithm. The hash is computed by:

1. Concatenating: Server ID (empty string since 1.7) + Shared Secret + Server's Public Key (DER encoded)
2. Computing SHA-1 of the concatenated bytes
3. Interpreting the result as a signed big-endian integer (two's complement)
4. Converting to hexadecimal with a minus sign prefix if negative

This means approximately half of all hashes will be negative (start with `-`).

**Examples:**

- `sha1("Notch")` → `4ed1f46bbe04bc756bcb17c0c7ce3e4632f06a48`
- `sha1("jeb_")` → `-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1` (negative!)

## API Endpoints

### POST `/session/minecraft/join`

Called by the client to authenticate with a server.

**Request Body:**

```json
{
  "accessToken": "minecraft_access_token",
  "selectedProfile": "player_uuid_without_dashes",
  "serverId": "computed_server_hash"
}
```

**Response:**

- `204 No Content` - Success
- `403 Forbidden` - Invalid token or session

### GET `/session/minecraft/hasJoined`

Called by the server to verify a client authenticated.

**Query Parameters:**

- `username` - Player name (case insensitive)
- `serverId` - The computed server hash
- `ip` - (Optional) Client IP for additional verification

**Response on Success (200):**

```json
{
  "id": "player_uuid_without_dashes",
  "name": "PlayerName",
  "properties": [
    {
      "name": "textures",
      "value": "base64_encoded_texture_data",
      "signature": "base64_encoded_signature"
    }
  ]
}
```

**Response on Failure:**

- `204 No Content` - Player hasn't joined or session expired

## Usage

### Client-Side (Joining a Server)

```go
client := session_server.NewSessionServerClient()
if err := client.Join(
    accessToken,      // from Microsoft/Minecraft auth
    playerUUID,       // player's UUID (no dashes)
    serverID,         // usually empty string
    sharedSecret,     // generated AES key
    serverPublicKey,  // server's RSA public key (DER)
); err != nil {
    // authentication failed
}
```

### Server-Side (Verifying a Client)

```go
client := session_server.NewSessionServerClient()

// compute the same hash the client used
serverHash := session_server.ComputeServerHash(
    "",               // server ID (empty since 1.7)
    sharedSecret,     // decrypted from client
    publicKeyDER,     // your server's public key
)

profile, err := client.HasJoined(username, serverHash)
if err != nil {
    // error occurred
}
if profile == nil {
    // player didn't authenticate - reject connection
}

// success! profile.ID contains the verified UUID
```

### Custom Session Server

For private servers or testing, you can use a custom mock session server URL:

```go
client := session_server.NewClientWithURL("https://my-auth-server.example.com")
```

## Rate Limits

Mojang's session server is rate limited to approximately 600 requests per 10 minutes. Servers should cache results when possible.

## References

- [Mojang API - Minecraft Wiki](https://minecraft.wiki/w/Mojang_API)
- [Protocol Encryption - Minecraft Wiki](https://minecraft.wiki/w/Java_Edition_protocol/Encryption)
