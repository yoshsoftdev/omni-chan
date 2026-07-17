# Provider Profile

## Status

Accepted

## Date

2026-07-15

## Context

A customer can interact through multiple external accounts. Each provider exposes identifiers and profile fields with different scopes and lifecycles. These values describe an external account, not the complete business relationship with the customer.

Examples include:

- WhatsApp: normalized phone number, WhatsApp profile name, and provider identifiers.
- Instagram: Instagram-scoped user ID, username, display name, and profile image.
- Messenger: Page-scoped user ID, display name, and profile image.
- Telegram: user or chat ID, username, and display name.
- Email: normalized email address, sender name, and mailbox-specific metadata.

## Problem

Storing provider-specific data directly on Contact mixes external account state with CRM state. It creates provider columns on the canonical customer model, encourages provider data to overwrite business data, and makes additional providers require repeated domain changes.

## Decision

A Provider Profile represents exactly one external customer account within one provider context.

A Provider Profile:

- belongs to exactly one Contact at a time;
- has a stable internal identifier;
- identifies its provider and provider account or channel scope;
- stores the provider-scoped external user identifier;
- stores provider-specific profile data such as username, display name, avatar, phone, email, and metadata when applicable;
- records synchronization and observation timestamps;
- may be active, unavailable, or unlinked according to provider lifecycle rules;
- is unique within the combination of tenant, provider account scope, and external user identifier.

Contact stores business-owned CRM data:

- canonical customer identity;
- manual name overrides;
- notes and custom fields;
- tags and segment membership;
- consent and marketing state;
- automation state;
- business classifications.

Provider Profile data is descriptive evidence from an external system. Contact data is the business record. Synchronizing a Provider Profile does not overwrite Contact except through explicit source-of-truth rules.

## Alternatives Considered

### Add provider columns to Contact

Rejected because every new provider would expand and couple the Contact model.

### Store provider profiles as unstructured Contact metadata

Rejected because identity uniqueness, linking, querying, lifecycle, and ownership would be ambiguous.

### Make Provider Profile the CRM record

Rejected because CRM data would fragment across external accounts.

### Create one Contact for every external account

Rejected because this duplicates customers and splits business state.

## Consequences

### Positive

- Provider data has clear ownership and lifecycle.
- One Contact can represent one customer across many external accounts.
- New provider fields do not alter canonical CRM semantics.
- Linking and unlinking can occur without moving messages.
- Provider identity can be synchronized independently.

### Negative

- Contact resolution requires an additional Provider Profile lookup.
- Duplicate external accounts require uniqueness and collision handling.
- Profile data can disagree with canonical Contact data by design.
- Provider account scope must be modeled correctly because external IDs may not be globally unique.

## Future Considerations

- Define profile verification, confidence, and provenance.
- Define profile deactivation and provider account deletion behavior.
- Define retention policy for profile metadata and avatars.
- Define whether aliases or historical provider identifiers are retained.
- Define capability and policy metadata associated with each Provider Profile.

## Related ADR

- [ADR 0001: WhatsApp as Source of Truth](0001-whatsapp-source-of-truth.md)
- [ADR 0004: Link Account instead of Merge Contact](0004-contact-linking.md)
- [ADR 0005: Conversation Ownership](0005-conversation-model.md)
