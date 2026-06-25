# Business ID Storage Decision

**Date:** 2026-06-22  
**Decision:** `channels.business_id` (Option A)

## Options Considered

| Option | Location | Pros | Cons |
|--------|----------|------|------|
| **A (chosen)** | `channels.business_id` | Matches Meta asset model (WABA/phone per channel); supports multi-channel teams with different businesses | Duplicated if same business connects twice |
| B | `teams.meta_business_id` | Single business per workspace assumption | Wrong when team has multiple WABAs/businesses; couples team to one Meta business |

## Rationale

Meta Embedded Signup provisions **per channel** assets:

- Business Manager ID
- WABA ID
- Phone Number ID

A workspace may operate multiple WhatsApp channels (different brands, regions, or numbers). Storing `business_id` on `channels` keeps Meta identity colocated with `waba_id` and `phone_number_id`.

## Migration

`000017_embedded_signup.up.sql` adds:

```sql
ALTER TABLE channels ADD COLUMN business_id VARCHAR(64);
```

## Usage

Set on embedded signup completion from Meta SDK `WA_EMBEDDED_SIGNUP` message payload `business_id`.
