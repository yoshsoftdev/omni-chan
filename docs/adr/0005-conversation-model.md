# Conversation Ownership

## Status

Accepted

## Date

2026-07-15

## Context

A Contact may communicate through several external accounts and providers. Every conversation occurs in a specific provider context with provider-specific identifiers, capabilities, policies, participants, and delivery semantics.

## Problem

If Conversations belong directly to Contact without Provider Profile ownership, their external identity becomes ambiguous. Moving a conversation when Contacts are linked can rewrite history, break message references, and falsely imply that a conversation occurred on another provider.

## Decision

Conversation ownership follows these rules:

- A Conversation belongs to exactly one Provider Profile.
- A Provider Profile belongs to exactly one Contact at a time.
- A Contact can own many Provider Profiles.
- A Contact can therefore have many Conversations through its Provider Profiles.
- A Provider Profile may have multiple Conversations when the provider's thread model requires it.
- A Conversation has exactly one provider for its entire lifetime.
- A Conversation never changes Provider Profile, provider, provider account scope, or external thread identity.
- Messages belong to their original Conversation and are never moved.
- Linking or unlinking a Provider Profile changes Contact association only; it does not alter Conversation or Message ownership.
- Inbox may aggregate Conversations by Contact, but aggregation is a projection and does not rewrite ownership.
- Closing, assigning, or labeling a Conversation changes operational state, not historical provider ownership.

Conversation history is immutable with respect to identity and ownership. Corrections are represented through links, status changes, or audit records rather than rewriting where a conversation occurred.

## Alternatives Considered

### Conversation belongs directly to Contact

Rejected because provider context and external account identity become secondary or ambiguous.

### Conversation can change provider

Rejected because a WhatsApp thread cannot become an Instagram thread without falsifying history and breaking provider identifiers.

### Move Conversations during Contact merge

Rejected because linking Provider Profiles is safer and preserves ownership.

### One global Conversation per Contact

Rejected because provider capabilities, message identifiers, policy windows, and thread semantics differ.

## Consequences

### Positive

- Provider history and message identity remain trustworthy.
- Linking is possible without moving history.
- Inbox can show a unified customer view while preserving channel boundaries.
- Provider-specific policy and capability checks have a stable scope.
- Audit and replay behavior are deterministic.

### Negative

- A customer may have several simultaneous Conversations.
- Cross-channel Inbox views require aggregation.
- Assignment and unread state must be explicit per Conversation or deliberately projected.
- Reporting must avoid treating linked Conversations as one physical thread.

## Future Considerations

- Define canonical thread keys for providers with multiple thread concepts.
- Define cross-channel customer timelines as read-only projections.
- Define assignment policy across several Conversations owned by one Contact.
- Define archival and retention without changing immutable ownership.
- Define group conversation and multi-participant ownership semantics.

## Related ADR

- [ADR 0003: Provider Profile](0003-provider-profile.md)
- [ADR 0004: Link Account instead of Merge Contact](0004-contact-linking.md)
- [ADR 0006: Canonical Message](0006-canonical-message.md)
