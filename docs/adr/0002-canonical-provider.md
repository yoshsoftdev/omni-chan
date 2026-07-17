# Canonical Provider Architecture

## Status

Accepted

## Date

2026-07-15

## Context

Omni-chan supports multiple messaging providers with different payloads, identifiers, capabilities, delivery states, and webhook formats. The product domain includes Contacts, Conversations, Messages, Inbox, assignment, CRM, segments, and automation. Those concepts must remain stable when providers evolve or new providers are added.

The architectural flow is:

```text
Provider
    ↓
Canonical Provider Layer
    ↓
Domain
    ↓
Inbox
```

## Problem

If domain and Inbox logic consume WhatsApp, Instagram, or Messenger models directly, provider-specific assumptions spread across the system. This creates branching business logic, duplicated behavior, incompatible data, and high risk when adding or changing providers.

## Decision

Omni-chan uses a provider-independent architecture.

- Every inbound provider payload enters through a provider adapter.
- The adapter validates and converts the payload into a canonical event.
- Only canonical events cross into the domain layer.
- The domain resolves Provider Profiles, Contacts, Conversations, and Messages without consuming provider payload models.
- Inbox reads domain projections and does not interpret provider webhooks.
- Outbound domain commands target a canonical provider capability; the selected provider adapter translates the command into the provider API request.
- Provider-specific identifiers and payloads may be preserved as metadata or `raw_payload`, but they do not define domain behavior.
- Provider capability differences are represented explicitly as capabilities, not scattered provider-name conditionals.
- Omni-chan domain models must not import or expose WhatsApp, Instagram, Messenger, Telegram, or email SDK models.

## Alternatives Considered

### Direct provider integration in domain services

Rejected because it couples business logic to external models and causes provider conditionals throughout the application.

### Separate domain implementation per provider

Rejected because Contact, Conversation, Inbox, assignment, and CRM behavior would be duplicated and diverge.

### Normalize only in the frontend

Rejected because backend persistence, webhook processing, automation, and realtime behavior also require canonical semantics.

### Preserve only raw provider payloads

Rejected because raw payloads cannot support stable provider-independent queries and business rules.

## Consequences

### Positive

- Domain behavior is consistent across providers.
- Provider changes remain isolated.
- Canonical events are testable independently of external APIs.
- Inbox and CRM do not require provider-specific models.
- Adding a provider primarily requires an adapter and capability mapping.

### Negative

- Canonical models require careful versioning and governance.
- Some provider features cannot be represented perfectly and must remain in metadata.
- Mapping errors can lose semantics unless `raw_payload` is retained.
- Capability negotiation adds explicit architectural complexity.

## Future Considerations

- Define canonical event versioning and compatibility rules.
- Define capability discovery for text, media, replies, reactions, deletion, and delivery receipts.
- Define observability around adapter failures and unsupported provider features.
- Define replay rules for raw provider events after canonical mappings evolve.

## Related ADR

- [ADR 0003: Provider Profile](0003-provider-profile.md)
- [ADR 0006: Canonical Message](0006-canonical-message.md)
- [ADR 0007: Provider Abstraction](0007-provider-abstraction.md)
