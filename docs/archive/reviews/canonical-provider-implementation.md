# Canonical Provider Architecture — Implementation Review

## Architecture

The implementation follows ADR 0001–0007:

- `Contact` remains the canonical CRM record.
- `provider.Profile` owns provider-scoped identity and belongs to one Contact.
- `Conversation.provider_profile_id` is required and immutable.
- `Conversation.external_thread_id` is required and immutable.
- Legacy `Conversation.contact_id`, `channel_id`, and `channel_type` remain for API and ingestion compatibility.
- Contact and Inbox projections resolve the current Contact through Provider Profile ownership.
- Linking changes only `provider_profiles.contact_id`.
- Unlinking creates a standalone Contact container and leaves Conversations and Messages untouched.
- Only WhatsApp profiles can be selected as source of truth.
- Link, unlink, and source-of-truth changes create actor- and timestamp-bearing Contact activities.
- Optional Contact-to-Contact links are symmetric and do not change Provider Profile or Conversation ownership.
- Contact Detail keeps each Contact's provider data separate and exposes Contact links only as relationships.

Provider-specific SDK and webhook models do not enter the new domain. No provider messaging adapter, OAuth flow, Meta client behavior, campaign, broadcast, assignment, webhook service, or storage service was changed.

## Files Modified

### Backend

- Added the six-file `internal/provider` domain plus its tests.
- Added `internal/api/v1/handler/provider.go`.
- Wired provider dependencies and v1 routes in `cmd/api/main.go`, `internal/api/v1/deps.go`, and `internal/api/v1/routes.go`.
- Enriched Contact reads and provider-aware search in `internal/contact`.
- Added canonical conversation ownership and provider projections in `internal/conversation`.
- Added migration `000037_canonical_provider` (up/down).
- Added migration `000038_contact_links` (up/down) for normalized, tenant-safe, auditable, soft-deletable Contact pairs.
- Added symmetric Contact link repository/service operations and versioned list/link/unlink endpoints.

### Frontend

- Added shared provider/source-of-truth badges.
- Added the Link Contact dialog and channel aggregation for linked Contacts.
- Added provider APIs and canonical provider types.
- Updated both Contact tables to show Providers.
- Updated Contact Detail to show Channels without a separate Linked Contacts section.
- Changed the frontend-only source-of-truth label to `Main`; API and database names remain unchanged.
- Updated Inbox cards and header with provider, source-of-truth, and linked-account indicators.
- Updated realtime inbox deltas to preserve provider context and unknown future provider keys.
- Added the versioned Provider Profile API contract to `docs/openapi/v1.yaml`.

## Migration

Migration `000037_canonical_provider.up.sql`:

1. Creates `provider_profiles` with tenant/provider/channel/external-ID uniqueness.
2. Backfills demonstrably observed WhatsApp and legacy Instagram identities.
3. Adds `contacts.source_of_truth_profile_id` and assigns WhatsApp source profiles.
4. Creates `contact_provider_activities`.
5. Adds and backfills `conversations.provider_profile_id`.
6. Installs a compatibility trigger for legacy conversation writers.
7. Rejects attempts to change an existing Conversation's Provider Profile.
8. Makes Provider Profile ownership required.
9. Allows provisional Contacts without a canonical phone.
10. Removes legacy `ig:{id}` phone shims after profile backfill.
11. Enforces tenant/contact consistency for source-of-truth references.

Existing Conversations, Messages, Activities, and Statistics are not moved or rewritten.

Migration `000038_contact_links.up.sql` creates normalized Contact pairs scoped to one team. Linking and unlinking these pairs never updates `provider_profiles`, `conversations`, or messages.

## API

Added:

- `GET /api/v1/providers`
- `GET /api/v1/providers/:id`
- `GET /api/v1/contacts/:id/providers`
- `GET /api/v1/contacts/:id/activities`
- `POST /api/v1/providers/:id/link`
- `POST /api/v1/providers/:id/unlink`
- `PATCH /api/v1/providers/:id/source-of-truth`
- `GET /api/v1/contacts/:id/links`
- `POST /api/v1/contacts/:id/links`
- `DELETE /api/v1/contacts/:id/links/:linkedContactId`

Existing Contact, Conversation, Inbox, and WhatsApp endpoints remain available.

## Compatibility

- Existing WhatsApp Contacts are automatically represented by WhatsApp Provider Profiles.
- Manual/import/API phone records are not assumed to be observed WhatsApp accounts.
- Existing conversation creation paths may continue sending legacy Contact/channel fields; the database resolves and stores canonical ownership.
- Existing API fields are retained.
- Inbox `contact_id` resolves to the Provider Profile's current Contact after linking.
- Contact search now includes provider phone, username, display name, and email.
- WhatsApp remains the only permitted source of truth.
- Contact links are returned from either side of the pair.
- Linked Contacts do not contribute provider details or badges to each other; Provider Profile ownership remains singular and unchanged.

## Breaking Changes

No v1 response fields were removed or renamed. New response fields and endpoints are additive.

At the database level, `conversations.provider_profile_id` is now required and immutable. `contacts.phone` is nullable for provisional non-WhatsApp Contacts. Legacy writers are covered by the migration trigger.

## Regression Analysis

Verified:

- Canonical provider package tests pass.
- Contact, Conversation, Inbox, and v1 API package tests pass.
- Frontend TypeScript check passes.
- Frontend tests pass: 9 files, 26 tests.
- IDE lint diagnostics report no errors in changed files.

The full backend `go test ./...` run reaches and passes the changed packages, but the repository has unrelated pre-existing test-double compile failures in Campaign, Config, Event Consumer, Report, Template, and Workspace tests. Those failures concern missing methods such as `CreateRole`, `FindByIGUserID`, and `ListCursor`; they are outside Canonical Provider scope and were not modified.

Primary regression controls:

- A database trigger preserves compatibility for all existing WhatsApp conversation writers.
- Conversation ownership cannot be rewritten after creation.
- Link/unlink tests verify ownership-only behavior.
- Bidirectional Contact-link tests verify idempotency, tenant isolation, self-link rejection, symmetric unlinking, and unchanged Provider/Conversation ownership.
- Frontend tests verify linked-contact normalization, link payloads, and bidirectional channel aggregation.
- Migration contract tests verify backfill, ownership, search, activity, immutability, and legacy-field retention.
