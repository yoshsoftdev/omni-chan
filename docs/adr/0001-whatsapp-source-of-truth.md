# WhatsApp as Source of Truth

## Status

Accepted

## Date

2026-07-15

## Context

Omni-chan receives customer interactions from WhatsApp, Instagram, Messenger, and future providers. A single person may appear under multiple external accounts, while CRM workflows require one stable customer identity.

WhatsApp provides a phone-number-based identity that is broadly useful for customer operations. Other providers usually expose provider-scoped identifiers that cannot safely replace a customer's canonical identity.

## Problem

Without a defined source of truth, provider data can overwrite CRM data, create duplicate contacts, fragment tags and automation state, or change customer identity whenever another channel supplies a different name.

## Decision

WhatsApp is the source of truth for canonical customer identity.

- The normalized phone number is the primary customer identifier.
- A customer's canonical name follows the WhatsApp profile name unless a user has manually overridden the name in Omni-chan.
- A manual name override always takes precedence over provider-supplied names.
- CRM data belongs to `Contact`, including customer attributes, notes, consent, and business-owned fields.
- Tags belong to `Contact`.
- Segment membership belongs to `Contact`.
- Automation state and eligibility belong to `Contact`.
- Instagram, Messenger, Telegram, email, and future providers are represented through Provider Profiles and never overwrite canonical customer identity.
- Provider-specific names, usernames, avatars, and external identifiers remain on their Provider Profile.
- A non-WhatsApp Provider Profile may be linked to an existing Contact, but linking does not promote its provider identity to the canonical identity.
- When no WhatsApp identity is known, a Contact may exist with a provisional business identity, but it becomes phone-based when an authorized WhatsApp identity is linked. Existing manual overrides remain unchanged.

## Alternatives Considered

### Last provider update wins

Rejected because provider data has different scopes and reliability. It would allow an Instagram username or Messenger display name to unexpectedly replace CRM identity.

### Provider-neutral synthetic identity only

Rejected as the sole identity rule because it does not provide a practical customer identifier and makes deduplication and customer operations less predictable.

### Separate Contact per provider

Rejected because it fragments CRM data, tags, segments, automation, reporting, and customer history.

### Email as the primary identifier

Rejected because email is not guaranteed to be available in messaging channels and may be shared, missing, or unverified.

## Consequences

### Advantages

- Customer identity remains stable as new providers are added.
- CRM ownership is clear and independent of provider payloads.
- Tags, segments, and automation operate on one Contact.
- Manual corrections are protected from provider synchronization.
- Phone-based lookup and deduplication remain deterministic.

### Limitations

- Customers without a known phone number require provisional Contact identity.
- A person using multiple phone numbers may require explicit business-controlled linking.
- Phone number recycling and shared numbers require operational safeguards.
- Provider-scoped identity can indicate an account, but cannot by itself prove that two accounts belong to the same person.
- WhatsApp profile changes do not override manual edits.

## Future Considerations

- Define provenance and timestamps for canonical and manual identity fields.
- Define verification policy before attaching a newly discovered phone number to an existing Contact.
- Define handling for shared business numbers, recycled numbers, and multiple WhatsApp profiles.
- Consider an explicit identity confidence model without changing the source-of-truth rule.

## Related ADR

- [ADR 0003: Provider Profile](0003-provider-profile.md)
- [ADR 0004: Link Account instead of Merge Contact](0004-contact-linking.md)
- [ADR 0005: Conversation Ownership](0005-conversation-model.md)
