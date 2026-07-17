# Provider Abstraction

## Status

Accepted

## Date

2026-07-15

## Context

Each external provider has its own authentication, webhook payloads, API requests, identifiers, status values, rate limits, and capabilities. Omni-chan needs to isolate those differences while keeping business behavior canonical.

The adapter boundary is:

```text
Provider Payload
       ↓
Provider Adapter
       ↓
Canonical Event
       ↓
Domain
```

Outbound processing follows the reverse boundary: a canonical domain command is translated by the selected provider adapter into a provider API operation.

## Problem

When adapters contain business rules or know downstream modules, provider integrations become coupled to Inbox and CRM behavior. This duplicates logic, makes testing difficult, and causes new providers to require changes throughout the product.

## Decision

Every provider is integrated through a provider adapter.

An adapter may:

- authenticate provider requests;
- verify webhook signatures;
- parse provider payloads;
- map provider identifiers and timestamps;
- normalize payloads into versioned canonical events;
- translate canonical outbound commands into provider API requests;
- map provider responses and errors into canonical outcomes;
- declare provider capabilities;
- preserve the original provider payload as `raw_payload`;
- implement provider-specific retry and rate-limit interpretation at the integration boundary.

An adapter must not:

- contain CRM or customer identity policy;
- create business rules for Contact linking;
- know Inbox presentation or filtering;
- assign Conversations;
- manage campaigns;
- evaluate segments;
- apply tags;
- run automation;
- decide canonical Contact ownership;
- directly mutate unrelated domain modules.

Adapters depend on canonical contracts. Domain services depend on adapter interfaces or canonical events, never concrete provider SDK models.

Provider differences are expressed through explicit capabilities and canonical unsupported-operation errors. Provider-name conditionals outside adapter selection and capability enforcement are not permitted.

## Alternatives Considered

### Provider SDK calls directly from domain services

Rejected because SDK types and behavior would leak into business logic.

### One shared adapter with provider conditionals

Rejected because it becomes a central provider switch and weakens isolation.

### Adapter owns end-to-end message and Inbox behavior

Rejected because business behavior would diverge by provider.

### Generic passthrough adapter

Rejected because unnormalized payloads would move provider coupling downstream.

## Consequences

### Positive

- Provider-specific code has a clear boundary.
- Canonical domain logic is reusable.
- Adapters can be tested with provider fixtures.
- New providers do not require Inbox, CRM, assignment, campaign, or segment knowledge.
- Capability differences are visible and enforceable.

### Negative

- Adapter interfaces require disciplined ownership and versioning.
- Integration and domain failures must be translated consistently.
- Provider features without canonical support remain unavailable until the canonical contract evolves.
- Common infrastructure must not accidentally become a business-logic backdoor.

## Future Considerations

- Define adapter conformance tests and fixture requirements.
- Define capability registry and negotiation.
- Define canonical error taxonomy and retry classification.
- Define adapter lifecycle, version support, and deprecation.
- Define outbound idempotency and reconciliation contracts.
- Define sandbox and provider certification requirements.

## Related ADR

- [ADR 0002: Canonical Provider Architecture](0002-canonical-provider.md)
- [ADR 0003: Provider Profile](0003-provider-profile.md)
- [ADR 0006: Canonical Message](0006-canonical-message.md)
