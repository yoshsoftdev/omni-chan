# Channel Meta Readiness

A channel is Meta-ready when all validation gates pass.

## Validation API

```go
channel.ValidateForMeta(ch *Channel, creds *Credentials, accessToken string) error
```

For campaign/broadcast use, also require:

```go
channel.ValidateForMetaWithOnboarding(ch, ob *Onboarding, creds, accessToken) error
```

## Required State

| Check | Field / Condition |
|-------|-------------------|
| Connected | `channels.status = connected` |
| WABA linked | `channels.waba_id` not empty |
| Phone linked | `channels.phone_number_id` not empty |
| Access token | `channel_credentials.access_token_enc` present and decryptable |
| Token not expired | `token_expires_at` null or in the future |
| Onboarding complete (campaigns) | `channel_onboarding.status = completed` |

## Workspace Readiness

`meta_connected` readiness check uses `channels.CountMetaReady(teamID)`:

- Joins `channels`, `channel_credentials`, `channel_onboarding`
- Requires connected + completed onboarding + waba + phone + valid token

## Health Endpoint

`GET /api/v1/channels/:id/health` runs live Meta API probes:

```json
{
  "healthy": true,
  "checks": {
    "token": true,
    "waba": true,
    "phone": true,
    "templates": true
  }
}
```

## Error Codes

Validation failures return `validation_error` with field `channel_id` or `meta`.
