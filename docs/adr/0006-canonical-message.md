# Canonical Message

## Status

Accepted

## Date

2026-07-15

## Context

Messaging providers express content and lifecycle events differently. WhatsApp, Instagram, Messenger, Telegram, email, and future providers use different payload shapes for text, media, replies, reactions, locations, interactive content, delivery state, and deletion.

The domain and Inbox require one stable message representation while retaining provider evidence.

## Problem

Persisting or exposing provider message models directly makes domain behavior provider-dependent. A model based on one provider either loses other-provider semantics or grows provider-specific fields and conditionals.

## Decision

All provider messages map into a canonical Message before entering domain behavior.

The canonical Message envelope contains, at minimum:

- stable internal message identifier;
- owning Conversation;
- direction: inbound, outbound, or system when applicable;
- canonical message type;
- sender and recipient references;
- provider external message identifier;
- reply or relationship reference when present;
- canonical timestamp fields;
- canonical delivery or processing state;
- deletion state;
- structured canonical content;
- provider metadata;
- original provider payload in `raw_payload`.

Canonical content supports these variants:

- **Text** — text body and optional formatting metadata.
- **Media** — media category, MIME type, caption, filename, size, and canonical storage reference where available.
- **Reaction** — reaction value and target message reference.
- **Reply** — referenced canonical or provider message identity plus the new message content.
- **Location** — latitude, longitude, name, and address where supplied.
- **Interactive** — canonical representation of buttons, selections, postbacks, forms, or provider interaction responses.
- **Metadata** — extensible structured attributes that do not control core business behavior.

Rules:

- Provider payloads are normalized before domain processing.
- Unknown provider fields remain available in `raw_payload`.
- Provider-specific fields that have no canonical equivalent remain in metadata or `raw_payload`.
- Canonical fields must have provider-independent semantics.
- Missing provider capabilities produce absent canonical fields, not fabricated values.
- Message ownership and Conversation identity are immutable.
- Replies and reactions reference existing messages when resolvable and retain the external reference when not resolvable.
- Outbound canonical commands are mapped by a provider adapter into supported provider operations.

## Alternatives Considered

### One message model per provider

Rejected because Inbox, search, reporting, realtime, and automation would require provider-specific branches.

### WhatsApp message model as the universal model

Rejected because WhatsApp semantics are not canonical and cannot represent every future provider cleanly.

### Store only raw payload

Rejected because domain queries and behavior need stable fields.

### Flatten every field into one large table or object

Rejected because mutually exclusive content shapes become ambiguous and provider-specific growth remains unbounded.

## Consequences

### Positive

- Inbox and domain logic consume one message contract.
- Common behavior works across providers.
- Raw payload preserves audit and replay evidence.
- New message types can be added deliberately.
- Provider capability gaps remain explicit.

### Negative

- Mapping requires careful semantic decisions.
- Canonical evolution needs compatibility rules.
- Some provider details are only available through metadata or raw payload.
- Resolving replies and reactions may require deferred reconciliation.
- Raw payload retention has privacy and storage implications.

## Future Considerations

- Define canonical schema versioning.
- Define edit, unsend, expiration, forwarding, and ephemeral-message semantics.
- Define group participants and email threading.
- Define content security scanning and media retention.
- Define replay and remapping of historical raw payloads.
- Define canonical delivery-state ordering across providers.

## Related ADR

- [ADR 0002: Canonical Provider Architecture](0002-canonical-provider.md)
- [ADR 0005: Conversation Ownership](0005-conversation-model.md)
- [ADR 0007: Provider Abstraction](0007-provider-abstraction.md)
