# Ponytail Review — Multi-Channel Inbox Indicators & Onboarding Navigation

## Scope

UI-only preparation for multi-channel support:

- Rename sidebar navigation label WhatsApp → Onboarding (route `/devices` unchanged).
- Split Onboarding page into independent WhatsApp and Instagram cards.
- Show unread counts on the existing inbox channel selector.
- Show a small platform indicator on conversation cards.
- Keep a single Inbox menu with total unread badge/label.
- Prefix inbound message notifications with platform.

No backend API changes. No WhatsApp Embedded Signup logic changes. No Instagram OAuth/API implementation. No inbox business-logic changes (assignment, contacts, send/receive).

## Modified Files (6)

1. `frontend/src/components/layout/sidebar.tsx`
2. `frontend/src/features/devices/components/devices-page.tsx`
3. `frontend/src/features/inbox/components/conversation-list.tsx`
4. `frontend/src/features/inbox/components/conversation-card.tsx`
5. `frontend/src/features/inbox/components/inbox-page.tsx`
6. `docs/reviews/ponytail-onboarding-navigation.md` (this review)

## Navigation Changes

| Before | After |
| --- | --- |
| Sidebar label `WhatsApp` → `/devices` | Sidebar label `Onboarding` → `/devices` |
| Page title `Connect WhatsApp` | Page title `Onboarding` |
| Single WhatsApp connect surface | Two independent cards: WhatsApp Business + Instagram Business |

WhatsApp card:

- Status read from existing `channelQueryService.getChannels()` (`Connected` if any channel `status === "connected"`).
- Existing Embedded Signup + Manual Connect remain unchanged; revealed via **Connect WhatsApp**.

Instagram card:

- Always **Not Connected** (no Instagram backend yet; no mock connected state).
- **Connect Instagram** button is present but disabled (no OAuth/API).

## Inbox Indicator Changes

| Area | Change |
| --- | --- |
| Channel selector | Labels include unread counts, e.g. `All Channels (12)`, `WhatsApp (7)` |
| Conversation card | Small platform row (`🟢 WhatsApp` / `🟣 Instagram` / …) above contact name |
| Sidebar Inbox | Label shows `Inbox (N)` plus existing badge from total unread |
| Inbound toast | When message arrives for a non-selected conversation: `🟢 WhatsApp · Budi · Halo Admin` |

`All Channels` unread uses existing dashboard total (`useInboxUnreadCount`). Per-channel unread is summed from currently loaded inbox conversations (no new API).

## Regression Check

- [x] `/devices` route still works; Embedded Signup links unchanged (`/devices/embedded-signup?returnTo=/devices`).
- [x] WhatsApp Embedded Signup / Manual Connect components not edited.
- [x] Single Inbox menu retained (no WhatsApp/Instagram inbox split).
- [x] Channel selector still filters via existing `channel_type` query param.
- [x] Conversation open/send/assign/presence flows untouched aside from inbound notify toast.
- [x] No mock Instagram connection status introduced.

## Remaining Findings

1. **`channel_type` not wired into inbox mapper** — `mapConversationToInbox` still defaults `channel: "whatsapp"`. Platform indicators and per-channel unread for Instagram/Messenger/Email will stay at WhatsApp until a follow-up maps API `channel_type` (would require additional mapper/type files beyond this Ponytail file budget).
2. **Per-channel unread accuracy** — counts for WhatsApp/Instagram/etc. are derived from the currently loaded page of conversations. Exact workspace-wide per-channel unread needs a dedicated backend aggregate (not added; APIs unchanged).
3. **Instagram Connect** — intentionally disabled until Instagram Messaging backend exists.
4. **Toast formatting** — notification uses a single-line `·` separator because the shared toast component does not support multiline layout (toast component left unchanged).

## Confirmation

No existing WhatsApp onboarding flow internals were modified. No inbox assignment/contact/send business logic was modified. Changes are limited to navigation labels, onboarding page composition, and inbox channel indicators/notifications.
