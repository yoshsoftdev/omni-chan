# Phase 11.5 — Inbox Synchronization Review

**Date:** 2026-06-22  
**Verdict:** APPROVE WITH FINDINGS

## Summary

Inbox synchronization layer connects message ingestion to team inbox UX: lifecycle, unread counters, assignment, SLA timestamps, filtered views, dashboard metrics, and WebSocket realtime fanout.

## Conversation Lifecycle

**PASS**

- States: open, pending, resolved, closed
- Inbound reopen from resolved/closed → open
- `resolution_at` on resolve/close

## Assignment Engine

**PASS**

- manual, round_robin, least_loaded
- Settings in `team_inbox_settings`
- Unassigned fallback

## Realtime Layer

**PASS**

- WebSocket at `/ws` and `/api/v1/ws`
- Events: conversation.created/updated, message.created, assignment.updated
- Redis pub/sub fanout via existing hub

## Unread Counters

**PASS**

- Increment on inbound
- Mark read API `POST /conversations/:id/read`
- Optional reset on agent reply via `reset_unread_on_reply`
- Never negative (clamped in repository)

## SLA Tracking

**PASS**

- Timestamps on conversations
- Dashboard averages on `GET /api/v1/dashboard`

## Message Model Audit

**PASS WITH GAPS**

- `raw_payload` preserves Meta structures
- Reply/interactive/reactions not normalized (documented)

## Build & Test

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./internal/inbox/...` | PASS |
| `go test ./internal/eventconsumer/...` | PASS |

## Findings

1. Assignment mode defaults to `manual` — teams must configure `team_inbox_settings` for auto-assign
2. Presence tracked on WebSocket connect/disconnect only (no heartbeat)
3. Typing indicators explicitly out of scope
4. Migration `000019` required for SLA and assignment tables

## Recommendation

Proceed to **Phase 11.6 — Campaign Delivery Tracking** for status webhooks → `campaign_recipients` updates.
