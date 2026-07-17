# Link Account instead of Merge Contact

## Status

Accepted

## Date

2026-07-15

## Context

The same customer may contact Omni-chan through multiple provider accounts. Operators need to associate those accounts with one CRM Contact without corrupting message history or duplicating business records.

Provider Profiles own provider identity, while Contacts own canonical CRM data. Conversations belong to Provider Profiles.

## Problem

Traditional contact merge operations often move conversations and messages, rewrite ownership, combine conflicting fields, and delete one record. These actions are difficult to reverse, weaken auditability, and can create cross-provider identity errors.

## Decision

Omni-chan links Provider Profiles to Contacts instead of merging Contacts or conversation history.

- Never merge Conversations.
- Never move Messages.
- Never rewrite a Conversation's Provider Profile.
- Never create a duplicate Contact as part of linking.
- Linking changes only the Contact ownership reference of a Provider Profile.
- All Conversations continue to belong to their original Provider Profile.
- Existing message identifiers, timestamps, provider, channel, direction, and raw payload remain unchanged.
- CRM data on the destination Contact is not automatically overwritten by Provider Profile data.
- Linking requires an explicit actor or a separately approved deterministic identity rule.
- Linking is tenant-scoped and must not associate records across tenants.
- A linking operation must be auditable, including actor, previous Contact, new Contact, reason, and timestamp.

Unlink behavior:

- Unlinking removes the Provider Profile from its current Contact ownership.
- Existing Conversations and Messages remain attached to the Provider Profile and remain unchanged.
- If product behavior requires every Provider Profile to have a Contact, unlink creates or selects a valid Contact container without copying or moving Conversations or Messages.
- Unlinking does not restore CRM fields that may have been manually edited after linking.
- Unlinking must be auditable and reversible by relinking the same Provider Profile.

## Alternatives Considered

### Destructive Contact merge

Rejected because it combines unrelated concerns, is hard to reverse, and risks rewriting historical ownership.

### Move Conversations to the destination Contact

Rejected because Conversations belong to Provider Profiles and their provider context must remain immutable.

### Copy Messages into a new Conversation

Rejected because it creates duplicate history and breaks provider message identity and audit trails.

### Automatic linking by display name

Rejected because display names and usernames are not reliable proof of identity.

### Keep every provider account permanently separate

Rejected because CRM, tags, segments, and automation would remain fragmented.

## Consequences

### Positive

- Linking is small, reversible, and auditable.
- Provider and message history remain intact.
- CRM consolidation does not alter conversation evidence.
- Identity mistakes can be corrected without data migration.
- Linking works uniformly for current and future providers.

### Negative

- Queries must resolve Contact through Provider Profile ownership.
- UI and reporting must distinguish account linking from history merging.
- Unlinking requires a defined Contact container policy.
- Incorrect links still affect CRM aggregation until corrected.

## Future Considerations

- Define permissions and approval requirements for link and unlink operations.
- Define suggested links using identity confidence without automatic destructive action.
- Define conflict presentation when Provider Profiles disagree.
- Define audit retention and restoration behavior.
- Define a policy for duplicate Contacts discovered independently of provider linking.

## Related ADR

- [ADR 0001: WhatsApp as Source of Truth](0001-whatsapp-source-of-truth.md)
- [ADR 0003: Provider Profile](0003-provider-profile.md)
- [ADR 0005: Conversation Ownership](0005-conversation-model.md)
