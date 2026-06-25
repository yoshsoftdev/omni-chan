# Tenant Resolution

Consumers resolve tenant context from canonical events because `team_id` and `channel_id` are empty at webhook ingestion.

## Flow

```
phone_number_id (metadata / payload)
        ↓
channels.FindByPhoneNumberID
        ↓
channel.team_id + channel.id
```

Template events without `phone_number_id` resolve via `waba_id`:

```
waba_id → channels.FindByWABAID → team + channel
```

## Failure Handling

| Condition | Action |
|-----------|--------|
| Unknown `phone_number_id` | Terminal error → `meta.consumer.dlq.channel` |
| Unknown `waba_id` (template) | Terminal error → `meta.consumer.dlq.channel` |
| Malformed payload | Terminal error → category-specific consumer DLQ |

Consumers never guess tenant ownership.

## Envelope Enrichment

After resolution, handlers use `team_id` and `channel_id` internally. Published events are not mutated.
