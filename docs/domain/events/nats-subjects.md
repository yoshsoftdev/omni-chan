# NATS Subject Design

JetStream stream: `WEBHOOK_EVENTS`  
Subjects: `events.>` (legacy), `meta.>` (Phase 11.3+)

## Meta Subjects

| Canonical type | NATS subject |
|----------------|--------------|
| `message.received` | `meta.message.received` |
| `message.status.sent` | `meta.message.status.sent` |
| `message.status.delivered` | `meta.message.status.delivered` |
| `message.status.read` | `meta.message.status.read` |
| `message.status.failed` | `meta.message.status.failed` |
| `template.updated` | `meta.template.updated` |
| `channel.updated` | `meta.channel.updated` |

## Dead Letter Subjects

| Category | Subject |
|----------|---------|
| Message | `meta.dlq.message` |
| Status | `meta.dlq.status` |
| Template | `meta.dlq.template` |
| Channel | `meta.dlq.channel` |
| Unknown | `meta.dlq.unknown` |

## Versioning

- Subject names are stable (no version suffix).
- Breaking envelope changes require new subject prefix (e.g. `meta.v2.`) — not used in Phase 11.3.

## Consumer Binding (Phase 11.4)

```
meta.message.>
meta.message.status.>
meta.template.>
meta.channel.>
```

No database assumptions in subject names.
