# Architecture Decision Records

## Purpose

Architecture Decision Records (ADRs) are the source of truth for significant Omni-chan architecture decisions. They record the context, problem, chosen decision, rejected alternatives, and consequences so future work can follow an explicit and durable foundation.

Implementation must conform to accepted ADRs. An ADR explains architectural intent; it is not an implementation task, migration plan, or source-code specification.

## ADR Index

- [ADR 0001: WhatsApp as Source of Truth](0001-whatsapp-source-of-truth.md)
- [ADR 0002: Canonical Provider Architecture](0002-canonical-provider.md)
- [ADR 0003: Provider Profile](0003-provider-profile.md)
- [ADR 0004: Link Account instead of Merge Contact](0004-contact-linking.md)
- [ADR 0005: Conversation Ownership](0005-conversation-model.md)
- [ADR 0006: Canonical Message](0006-canonical-message.md)
- [ADR 0007: Provider Abstraction](0007-provider-abstraction.md)

## ADR Numbering

- ADRs use a four-digit, sequential number beginning with `0001`.
- File names use `NNNN-short-kebab-case-title.md`.
- Numbers are never reused, even when an ADR is superseded or rejected.
- The number identifies the decision record, not an implementation phase.
- ADRs remain in this directory to preserve architectural history.

## ADR Status

Use one of these statuses:

- **Proposed** — under discussion and not yet binding.
- **Accepted** — approved and binding for future implementation.
- **Rejected** — considered but not approved.
- **Deprecated** — retained for history but no longer recommended.
- **Superseded by ADR NNNN** — replaced by a newer accepted decision.

Accepted ADRs are not edited to silently reverse a decision. A material change requires a new ADR that references and supersedes the old ADR. Small clarifications that do not alter the decision may be added with review.

## How to Create a Future ADR

1. Select the next unused sequential number.
2. Create `docs/adr/NNNN-short-title.md`.
3. Use the required structure:
   - `# Title`
   - `Status`
   - `Date`
   - `Context`
   - `Problem`
   - `Decision`
   - `Alternatives Considered`
   - `Consequences`
   - `Future Considerations`
   - `Related ADR`
4. State enforceable rules in the Decision section.
5. Explain why alternatives were rejected.
6. Include both positive and negative consequences.
7. Link related or superseded ADRs.
8. Add the ADR to the index in this README.
9. Obtain architecture review before changing the status to Accepted.

## When an ADR Is Required

Create an ADR before implementation when a decision:

- changes domain ownership or boundaries;
- introduces or changes a canonical model;
- changes the source of truth for business data;
- changes identity, linking, or deduplication policy;
- affects multiple services or product modules;
- introduces a provider abstraction or integration boundary;
- changes data immutability, retention, or audit guarantees;
- has meaningful alternatives or difficult-to-reverse consequences;
- establishes a convention that future providers or modules must follow;
- supersedes an accepted architecture decision.

An ADR is usually unnecessary for local refactoring, routine bug fixes, formatting, or an implementation detail that does not change an architectural contract.

## Architecture Principles

### 1. Provider Agnostic

Business logic must never depend on a specific provider. Provider identity may select an adapter or capability, but it must not redefine Contact, Conversation, Message, Inbox, CRM, assignment, segment, or automation behavior.

### 2. Source of Truth

WhatsApp is the canonical customer identity. Phone number is the primary identifier, WhatsApp supplies the default customer name, and manual overrides take precedence. Other providers never overwrite canonical identity.

### 3. Immutable Conversation

Conversation and Message history must never be rewritten to change provider, Provider Profile, or external thread ownership. Operational state may change; historical identity does not.

### 4. Link Instead of Merge

External accounts are linked through Provider Profiles. Linking changes Contact ownership of a Provider Profile only. Conversations are never merged and Messages are never moved.

### 5. Canonical Domain

Every provider payload is normalized into a canonical event before entering the domain layer. Domain and Inbox consume canonical contracts, not provider payload models.

### 6. Provider Isolation

Provider-specific authentication, payload parsing, API translation, capabilities, and error mapping stay inside provider adapters. Adapters contain no Inbox, Campaign, Assignment, CRM, Segment, or Automation business logic.

### 7. Extensibility

Adding Telegram, Line, email, or another provider must not require changing existing domain model semantics. New provider details belong in adapters, Provider Profiles, capabilities, metadata, and raw payloads.

## Decision Authority

The accepted ADRs in this directory govern future Canonical Provider Architecture work. If an implementation requirement conflicts with an accepted ADR, the architecture decision must be reviewed and superseded through a new ADR before implementation proceeds.
