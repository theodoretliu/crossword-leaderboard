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

## Required Headers

### Authentication Cookie

```
NYT-S=<NYT_SESSION_TOKEN>
```

The `NYT-S` cookie is the primary session token. Format appears to be:
```
0^CB4SNg<BASE64_ENCODED_SESSION_DATA>
```

### App Identification Headers

| Header | Value | Required |
|--------|-------|----------|
| `nyt-app-type` | `NYT-iOS-Crossword` | Yes |
| `nyt-app-version` | `6.5.0` | Yes |
| `nyt-agent-id` | `<DEVICE_UUID>` | Yes |
| `Accept` | `application/json` | Yes |
| `Accept-Language` | `en-US,en;q=0.9` | Recommended |

### Request Signing Headers

The API uses request signing for validation:

| Header | Description |
|--------|-------------|
| `nyt-timestamp` | Unix timestamp (must be recent - server validates time skew) |
| `nyt-signature` | Base64-encoded signature (likely HMAC or RSA) |

## Request Signing

The API validates requests using a signature mechanism:

1. `nyt-timestamp` must be within an acceptable window of server time
2. `nyt-signature` is a Base64-encoded signature (~344 chars)
3. Error `RequestTimeTooSkewed` returned if timestamp is stale

### Signature Format

```
<BASE64_SIGNATURE>
```

The signature appears to be ~256 bytes (2048-bit) when decoded, suggesting RSA or similar asymmetric signing.

## Optional Headers

These headers were observed but may not be required:

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

## Other Cookies (Not Required)

These cookies were captured but don't appear necessary for API access:

| Cookie | Purpose |
|--------|---------|
| `nyt-a` | Analytics/device identifier |
| `nyt-gdpr` | GDPR consent flag |
| `nyt-geo` | Geolocation (e.g., `US`) |
| `regi_cookie` | Registration metadata |
| `nyt-jkidd` | User activity tracking |
| `datadome` | Bot detection |

## Error Responses

| Error | HTTP Status | Description |
|-------|-------------|-------------|
| `RequestTimeTooSkewed` | 403 | Timestamp too old/far from server time |

## Response Format

*TODO: Capture successful response to document shape*

Expected to return user puzzle statistics including:
- Solve times
- Streak information
- Leaderboard position

## Placeholder Summary

For scraping/automation, these values need to be obtained:

| Placeholder | Source |
|-------------|--------|
| `<NYT_SESSION_TOKEN>` | Login flow / browser cookies |
| `<DEVICE_UUID>` | Generate UUID v4 or extract from app |
| `<TIMESTAMP>` | Current Unix timestamp |
| `<SIGNATURE>` | Must be computed - algorithm unknown |
| `<PERSISTED_QUERY_HASH>` | Captured from app traffic |

## Open Questions

1. How is `nyt-signature` computed? Inputs likely include:
   - Timestamp
   - Request path/query
   - Possibly app secret or device-specific key

2. Can the web interface be used instead? (Simpler auth, no signing)

3. What other GraphQL operations are available?
