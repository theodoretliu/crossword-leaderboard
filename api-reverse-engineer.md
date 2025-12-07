# NYT Crossword API - Reverse Engineering Notes

## Endpoint

```
GET https://samizdat-graphql.nytimes.com/graphql/v2
```

## Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `operationName` | string | GraphQL operation name (e.g., `UserDetails`) |
| `variables` | JSON string | Query variables (e.g., `{"printDate": "2025-12-02"}`) |
| `extensions` | JSON string | Contains persisted query info |

### Extensions Format

```json
{
  "persistedQuery": {
    "sha256Hash": "<PERSISTED_QUERY_HASH>",
    "version": 1
  }
}
```

### Known Operations

| Operation | Hash | Description |
|-----------|------|-------------|
| `UserDetails` | `3f462df6ff876e20c737369faf1d3d65725fa54e62ff4039ee13c3e22ecc14f5` | Get user details for a specific puzzle date |

## Web Interface (Recommended)

The web/Phoenix interface is simpler than the iOS app - **no request signing required**.

### Required Headers

| Header | Value | Required |
|--------|-------|----------|
| `Accept` | `application/json` | Yes |
| `Accept-Language` | `en-US,en;q=0.9` | Recommended |
| `nyt-app-type` | `games-phoenix` | Yes |
| `nyt-app-version` | `1.0.0` | Yes |
| `nyt-token` | `<RSA_PUBLIC_KEY>` | Yes |

### nyt-token Header

The `nyt-token` is a base64-encoded RSA public key (2048-bit, DER format):

```
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAiKjdfob/ixNCvLETwnQ3AalkGSm9NX4gcRbOudrtHmBmIJbWb8Xgu3QH516Edr1qD7A+w+5d0p/WsNCpWDLrqfjTIwMft+jtOQG44l7akD9yi9Gaq/6hS3cuntkY25AYR3WtQPqrtxClX+qQdhMmzlA0sRAXKM8dSbIpsNV9uUOclt3JwB4omwFGj4J+pqzsfYZfB/tlx+BPGjCYGNcZ9O9UvtCpLRLgCJmTugL6V/U581gY8mqp+22aVjbEJik+F0j8xTNSxCOV2PLMpNrRSiDZ8FaKtq8ap/HPey5M7qYZQqclfqsEJMXG/KE3PiaTIbO37caFa80FvzfV8MZw1wIDAQAB
```

This appears to be a static public key used for token validation.

### Required Cookies

| Cookie | Purpose |
|--------|---------|
| `NYT-S` | Primary session token (required for auth) |
| `nyt-a` | Analytics/device identifier |
| `nyt-m` | User metadata/preferences |
| `nyt-gdpr` | GDPR consent flag |
| `nyt-purr` | Unknown (consent related?) |
| `nyt-geo` | Geolocation (e.g., `US`) |
| `regi_cookie` | Registration metadata |
| `nyt-traceid` | Request tracing |
| `nyt-jkidd` | User activity tracking |

The `NYT-S` cookie is the primary session token. Format:
```
0^CB4SNg<BASE64_ENCODED_SESSION_DATA>
```

## iOS App Interface (Complex)

The iOS app uses request signing, making it more complex to replicate.

### App Identification Headers

| Header | Value | Required |
|--------|-------|----------|
| `nyt-app-type` | `NYT-iOS-Crossword` | Yes |
| `nyt-app-version` | `6.5.0` | Yes |
| `nyt-agent-id` | `<DEVICE_UUID>` | Yes |
| `Accept` | `application/json` | Yes |
| `Accept-Language` | `en-US,en;q=0.9` | Recommended |

### Request Signing Headers

The iOS API uses request signing for validation:

| Header | Description |
|--------|-------------|
| `nyt-timestamp` | Unix timestamp (must be recent - server validates time skew) |
| `nyt-signature` | Base64-encoded signature (RSA 2048-bit) |

**Signing mechanism:**
1. `nyt-timestamp` must be within an acceptable window of server time
2. `nyt-signature` is a Base64-encoded signature (~344 chars, ~256 bytes decoded)
3. Error `RequestTimeTooSkewed` returned if timestamp is stale

### Optional iOS Headers

| Header | Example Value |
|--------|---------------|
| `User-Agent` | `Crossword/<BUILD_NUMBER> CFNetwork/<VERSION> Darwin/<VERSION>` |
| `nyt-build-type` | `release` |
| `nyt-os-version` | `18.6.2` |
| `nyt-device-model` | `iPhone` |

### Tracing Headers (Optional)

| Header | Description |
|--------|-------------|
| `x-datadog-trace-id` | Datadog distributed tracing |
| `x-datadog-parent-id` | Datadog parent span |
| `x-embrace-id` | Embrace SDK tracking |
| `traceparent` | W3C trace context |
| `tracestate` | W3C trace state |

## Error Responses

| Error | HTTP Status | Description |
|-------|-------------|-------------|
| `RequestTimeTooSkewed` | 403 | Timestamp too old/far from server time (iOS only) |

## Response Format

The `UserDetails` query returns friend leaderboard data:

```json
{
  "data": {
    "user": {
      "friends": {
        "edges": [
          {
            "node": {
              "gameScores": {
                "connections": { "numMistakes": 1 },
                "crosswordMini": { "solveTimeSeconds": 27 },
                "spellingBee": { "rank": "Genius", "score": 173 },
                "wordle": { "numGuesses": 3, "win": true }
              },
              "profile": { "gamesUsername": "FriendName" },
              "regiId": "12345678",
              "setting": {
                "games": { "avatar": { "value": "/Characters-11.png" } }
              }
            },
            "relationshipStatus": "FRIEND"
          }
        ]
      },
      "gameScores": {
        "connections": null,
        "crosswordMini": null,
        "spellingBee": null,
        "wordle": null
      },
      "profile": {
        "gamesUsername": "YourUsername",
        "username": null
      },
      "regiId": "86493738",
      "setting": {
        "games": { "avatar": { "value": "/Games-4.png" } }
      }
    }
  }
}
```

### TypeScript Interface

```typescript
interface UserDetailsResponse {
  data: {
    user: User;
  };
}

interface User {
  friends: {
    edges: FriendEdge[];
  };
  gameScores: GameScores;
  profile: {
    gamesUsername: string;
    username: string | null;
  };
  regiId: string;
  setting: UserSetting | null;
}

interface FriendEdge {
  node: {
    gameScores: GameScores;
    profile: {
      gamesUsername: string;
    };
    regiId: string;
    setting: UserSetting | null;
  };
  relationshipStatus: "FRIEND";
}

interface GameScores {
  connections: { numMistakes: number } | null;
  crosswordMini: { solveTimeSeconds: number } | null;
  spellingBee: { rank: string; score: number } | null;
  wordle: { numGuesses: number; win: boolean } | null;
}

interface UserSetting {
  games: {
    avatar: {
      value: string; // e.g., "/Characters-11.png"
    };
  };
}
```

Fields are `null` if the user hasn't played that game for the given date.

## Placeholder Summary

For web interface (recommended):

| Placeholder | Source |
|-------------|--------|
| `<NYT_SESSION_TOKEN>` | Login flow / browser cookies |
| `<RSA_PUBLIC_KEY>` | Static key (see above) |
| `<PERSISTED_QUERY_HASH>` | Captured from app/web traffic |

For iOS interface (not recommended):

| Placeholder | Source |
|-------------|--------|
| `<DEVICE_UUID>` | Generate UUID v4 or extract from app |
| `<TIMESTAMP>` | Current Unix timestamp |
| `<SIGNATURE>` | Must be computed - algorithm unknown |

## Open Questions

1. ~~Can the web interface be used instead?~~ **Yes! The Phoenix web interface works without signing.**

2. How is `nyt-signature` computed for iOS? (No longer needed since web works)

3. What other GraphQL operations are available?

4. Is the `nyt-token` static across all users, or per-user?
