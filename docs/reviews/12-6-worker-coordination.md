# Phase 12.6 — Worker Coordination Audit

**Verdict: APPROVE WITH FINDINGS**

**Date:** 2026-06-22  
**Scope:** Multi-replica safety across broadcast-orchestrator, broadcast-worker, event-consumer, webhook ingress

---

## Services

| Service | Process | Coordination |
|---------|---------|--------------|
| `broadcast-orchestrator` | `cmd/broadcast-orchestrator` | PG locks, unique job index, outbox, JetStream durable consumer |
| `broadcast-worker` | `cmd/broadcast-worker` | Batch lease, recipient status guards, WAMID idempotency, cancel flag |
| `event-consumer` | `cmd/api` (NATS enabled) | Idempotency store (Redis), explicit ACK/NAK |
| `webhook-service` | Meta → `cmd/api` `/webhook/meta` | HTTP ingress; events forwarded to NATS by platform wiring |

---

## Duplicate Processing Guards

| Scenario | broadcast-orchestrator | broadcast-worker | event-consumer |
|----------|------------------------|------------------|----------------|
| Worker restart | Outbox relay resumes pending rows | Batch lease expires → re-claim; recipient recovery | JetStream redelivery + idempotency key |
| NATS redelivery | Job handler idempotent on recipient count | `BatchCompleted` → no-op; `ErrBatchInProgress` → no-op | `provider:status:{wamid}:{status}` |
| Network interruption | Outbox not marked published until ACK | Lease refresh every 10 recipients | NAK on handler error |
| Delayed ACK | Outbox at-least-once publish | Processing lease prevents duplicate claim | Redis/memory idempotency |

---

## Lease Safety

| Entity | Lease | Recovery |
|--------|-------|----------|
| Batch | `locked_by` + `locked_at` (10m) | Orchestrator `Recovery.RunOnce` |
| Recipient `sending` | `updated_at` (15m stale) | Worker `RecipientRecovery` |
| Job | Status machine (no time lease) | Status guards only |

---

## Replay Safety Principle

```text
at-least-once delivery + conditional DB updates + idempotency keys
  → no duplicate side effects (counters, WAMID, completion)
```

Key implementations:

- `FinalizeRecipientSent` / `FinalizeRecipientFailed` — transactional, idempotent
- `delivery.CanTransition` — monotonic recipient status
- `eventconsumer/idempotency` — webhook dedupe

---

## Worker Registration (Phase 12.6)

Redis heartbeat keys for operational visibility:

| Role | Key pattern | TTL |
|------|-------------|-----|
| Orchestrator | `ops:worker:orchestrator:{id}` | 30s |
| Send worker | `ops:worker:send_worker:{id}` | 30s |
| Event consumer | `ops:worker:event_consumer:{id}` | 30s |

Exposed via `GET /api/v1/operations/workers`.

---

## Findings

### LOW

1. **Webhook-service not a separate binary** — Webhooks hit `cmd/api` directly; JetStream replay applies to event-consumer only. Document deployment: API replicas must all run event consumer when NATS enabled.

2. **No distributed lock on orchestrator scheduler tick** — Safe via `SKIP LOCKED`; multiple orchestrators may compete but only one transitions each campaign.

3. **Worker heartbeat requires Redis** — If Redis down, `operations/workers` returns zero counts; broadcast continues with PG-only cancel.

---

## Exit Criteria

- [x] Coordination mechanisms documented per service
- [x] Replay safety mapped to concrete guards
- [x] Worker heartbeats implemented for ops endpoints
