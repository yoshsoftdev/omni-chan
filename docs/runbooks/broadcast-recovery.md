# Broadcast Recovery Runbook

Operational guide for broadcast pipeline failures. Complements `GET /api/v1/operations/broadcast` and `GET /api/v1/operations/workers`.

---

## Architecture Quick Reference

```text
API (send/cancel) → broadcast_outbox → NATS → orchestrator → NATS → send-worker → Meta
                                                      ↓
                                              event-consumer (delivery webhooks)
```

**Authoritative state:** PostgreSQL  
**Coordination:** Redis (rate limits, cancel flags, worker heartbeats, idempotency)  
**Transport:** NATS JetStream (`BROADCAST` stream)

---

## Scenario A — Orchestrator Restart

### Symptoms

- Campaigns stuck in `queued` with no progress
- `queued_jobs` rising in `/operations/broadcast`
- `orchestrators: 0` in `/operations/workers`

### Verify

```bash
# Check orchestrator pods/processes
curl -s http://<orchestrator-host>:/health  # if exposed

# API ops (requires workspace.manage JWT)
curl -H "Authorization: Bearer $TOKEN" https://<api>/api/v1/operations/broadcast
curl -H "Authorization: Bearer $TOKEN" https://<api>/api/v1/operations/workers
```

### Recovery

1. Restart `cmd/broadcast-orchestrator`
2. Confirm NATS connectivity (`NATS_URL`)
3. Confirm PostgreSQL connectivity
4. Outbox relay resumes automatically — pending `broadcast_outbox` rows publish to JetStream
5. Scheduler tick resumes — due `scheduled` campaigns dispatch

### Expected outcome

- `queued_jobs` decreases as jobs progress to `sending`
- `orchestrators` count matches replica count within 30s

---

## Scenario B — Send Worker Restart

### Symptoms

- Batches stuck in `processing`
- `stale_batches` > 0 or `recovery_pending` > 0
- `send_workers: 0`

### Verify

```bash
curl http://<worker-host>:8091/health
# {"healthy":true,"nats":true,"postgres":true,"redis":true}
```

### Recovery

1. Restart `cmd/broadcast-worker` replicas
2. **Batch recovery (automatic):** Orchestrator `Recovery` requeues batches with expired lease (10m default)
3. **Recipient recovery (automatic):** Worker `RecipientRecovery` requeues `sending` rows without WAMID after 15m

### Manual intervention (if stale > 30m)

- Inspect `broadcast_batches` for `status = 'processing'` and old `locked_at`
- Inspect `campaign_recipients` for `status = 'sending'` and `wamid IS NULL`
- Safe to restart workers; do **not** delete batch rows

---

## Scenario C — Event Consumer / Webhook Restart

### Symptoms

- Sends succeed but delivery/read counts stall
- Campaigns remain `sending` with all recipients `sent`

### Verify

- `event_consumers` count in `/operations/workers`
- NATS consumer lag (Phase 13 metrics)

### Recovery

1. Restart `cmd/api` with `NATS_ENABLED=true`
2. JetStream replays unacked messages to durable consumers
3. Idempotency keys prevent duplicate counter increments

---

## Scenario D — PostgreSQL Restart

### Symptoms

- All services log connection errors
- `/health` endpoints unhealthy

### Recovery

1. Restore PostgreSQL availability
2. Restart in order: **PostgreSQL → NATS → Redis → orchestrator → workers → API**
3. No broadcast state is lost — jobs/batches/recipients persist in PG
4. In-flight `processing` batches recover via lease expiry

---

## Scenario E — Redis Unavailable

### Symptoms

- Worker health shows `redis: false` but `healthy: true`
- Rate limiting falls back to per-process limiter
- Cancel flags unavailable at worker (PG cancel still works)
- Worker heartbeats missing (`operations/workers` shows 0)

### Recovery

1. Restore Redis
2. Workers/orchestrators resume heartbeats automatically
3. **No data loss** — broadcast continues at reduced coordination fidelity
4. Re-cancel campaign if cancel was issued during Redis outage and sends continue (edge case)

---

## Campaign Cancel During Active Send

1. `POST /api/v1/campaigns/{id}/cancel`
2. PostgreSQL: campaign + cancellable recipients → `cancelled`
3. Redis: `cancel:campaign:{id}` set (24h TTL)
4. Workers skip remaining `queued` recipients on next poll
5. In-flight Meta requests may still return WAMID → recipient becomes `sent` (partial send expected)

---

## Operational Endpoints

| Endpoint | Permission | Fields |
|----------|------------|--------|
| `GET /api/v1/operations/broadcast` | `workspace.manage` | `queued_jobs`, `active_batches`, `stale_batches`, `recovery_pending` |
| `GET /api/v1/operations/workers` | `workspace.manage` | `orchestrators`, `send_workers`, `event_consumers` |

Read-only. No side effects.

---

## Escalation

| Condition | Action |
|-----------|--------|
| `recovery_pending` > 0 for > 30m | Check worker logs, batch leases, Meta API errors |
| Duplicate WAMID errors | Check `idx_campaign_recipients_wamid`; review worker scaling |
| Jobs not created | Check `idx_broadcast_jobs_one_active_per_campaign` conflicts |
| Outbox growing | Check NATS publisher; orchestrator outbox relay logs |

Phase 13 adds Prometheus metrics, Grafana dashboards, and alerting for these conditions.
