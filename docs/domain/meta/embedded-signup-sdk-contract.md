# Embedded Signup SDK Contract

Frontend ↔ Backend contract for Meta Embedded Signup (Coexistence).

## Prerequisites

- `GET /api/v1/meta/embedded-config` returns `app_id`, `config_id`, `api_version`, `enabled`
- Channel row exists (`POST /api/v1/channels` if needed)

## Flow

```text
1. GET  /meta/embedded-config
2. POST /channels/:channelId/embedded-signup/start
3. FB.login (Meta JS SDK) + WA_EMBEDDED_SIGNUP postMessage
4. POST /channels/:channelId/embedded-signup/:sessionId/complete
5. POST /workspace/readiness/refresh (optional — backend also refreshes on complete)
```

## SDK Initialization

```javascript
FB.init({ appId, version: api_version, xfbml: true, autoLogAppEvents: true });

FB.login(callback, {
  config_id: config_id,
  response_type: "code",
  override_default_response_type: true,
  extras: {
    setup: {},
    featureType: "whatsapp_business_app_onboarding",
    sessionInfoVersion: "3",
  },
});
```

## PostMessage Payload (Meta → Frontend)

Listen for `WA_EMBEDDED_SIGNUP` from `https://www.facebook.com`:

```json
{
  "type": "WA_EMBEDDED_SIGNUP",
  "data": {
    "waba_id": "...",
    "phone_number_id": "...",
    "business_id": "..."
  }
}
```

## Start Session

```http
POST /api/v1/channels/{channelId}/embedded-signup/start
```

Response:

```json
{
  "session_id": "uuid",
  "oauth_state": "hex",
  "status": "pending",
  "expires_at": "2026-06-22T12:00:00Z"
}
```

## Complete Signup

```http
POST /api/v1/channels/{channelId}/embedded-signup/{sessionId}/complete
```

Request:

```json
{
  "code": "authorization_code_from_FB.login",
  "oauth_state": "hex_from_start",
  "business_id": "optional_from_postMessage",
  "waba_id": "from_postMessage",
  "phone_number_id": "from_postMessage"
}
```

Success response (no access token):

```json
{
  "channel_id": "uuid",
  "status": "connected",
  "business_id": "...",
  "waba_id": "...",
  "phone_number_id": "...",
  "phone_number": "+..."
}
```

Failure response:

```json
{
  "success": false,
  "message": "invalid authorization code"
}
```

## Status Polling

```http
GET /api/v1/channels/{channelId}/embedded-signup/{sessionId}
```

## Security Rules

- Frontend never receives or stores Meta access tokens
- Always send `oauth_state` from start response on complete
- Session expires after 1 hour
