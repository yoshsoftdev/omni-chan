# Scaling Review — Phase 12.0

**Status:** Design review  
**Reference:** [`docs/broadcast-execution-architecture.md`](../broadcast-execution-architecture.md) §12

## Tiers

| Tier | Contacts / campaign | Platform notes |
|------|---------------------|----------------|
| **Starter** | 50,000 | Single channel typical |
| **Growth** | 300,000 | Multi-segment audiences |
| **Enterprise** | 1,000,000+ | Partitioned recipients, dedicated rate tier |

**Assumptions:**

- Batch size: **100** (default)
- Average send latency: **250 ms** per Meta API call (incl. network)
- Sequential sends within batch
- Channel limit: **80 mps** (conservative tier)

---

## Starter — 50k contacts

| Metric | Estimate |
|--------|----------|
| Recipients | 50,000 |
| Batches | 500 |
| NATS `batch.send` messages | ~500 (+ retries ~5% → 525) |
| Send time (1 worker, 80 mps) | ~625 seconds (~10 min) |
| Send time (10 workers, limiter-capped) | ~10 min (limit-bound, not worker-bound) |
| Queue depth (peak) | 500 |
| Orchestrator time | Audience resolve ~2–5 min (5k chunks) |

| Resource | Recommendation |
|----------|----------------|
| Send workers | 3–5 pods |
| Orchestrator | 1 pod |
| Redis memory | < 100 MB (rate keys) |
| NATS storage | < 50 MB per campaign |

---

## Growth — 300k contacts

| Metric | Estimate |
|--------|----------|
| Recipients | 300,000 |
| Batches | 3,000 |
| NATS messages | ~3,150 with retries |
| Send time (80 mps limit) | ~62 minutes |
| Send time (250 mps tier) | ~20 minutes |
| Queue depth (peak) | 3,000 |
| Audience resolve | ~15–30 min (streaming inserts) |

| Resource | Recommendation |
|----------|----------------|
| Send workers | 10–20 pods |
| Orchestrator | 2 pods (HA) |
| Postgres | Bulk `COPY` recipients; partition `campaign_recipients` by month |
| Connection pool | PgBouncer; workers use small pool |

**Bottleneck:** Meta WABA tier, not workers.

---

## Enterprise — 1M+ contacts

| Metric | Estimate |
|--------|----------|
| Recipients | 1,000,000+ |
| Batches | 10,000+ |
| Send time (1000 mps tier) | ~17 min theoretical |
| Send time (80 mps) | ~3.5 hours |

| Resource | Recommendation |
|----------|----------------|
| Send workers | 50–100 pods (HPA max) |
| Channel affinity | Consider per-channel subject sharding |
| Audience resolver | Dedicated job; stream from segment cursor |
| Partitioning | Monthly `campaign_recipients` partitions (already in schema) |
| Campaign concurrency | Limit 1 active send per channel |

---

## Worker count formula

```
min_workers = ceil(target_duration_sec × channel_mps / batch_size)
```

Example: 300k contacts, target 30 min, 80 mps, batch 100:

```
needed_mps = 300000 / 1800 = 167 mps → exceeds 80 mps tier
```

**Conclusion:** Duration is **always limit-bound** until WABA tier upgraded. Workers scale to consume queue without adding latency pile-up:

```
workers = min(N_max, ceil(consumer_lag / batches_per_worker_per_min))
```

Start HPA at lag > 100 messages.

---

## Database scaling

| Table | Growth | Mitigation |
|-------|--------|------------|
| `campaign_recipients` | 1M rows/campaign | Range partitions by `created_at` |
| `broadcast_batches` | 10k rows/campaign | Index on `(job_id, status)` |
| `campaigns` counters | Hot row updates | Delta updates (Phase 11.6); avoid per-message trigger |

---

## NATS scaling

| Setting | Value |
|---------|-------|
| Stream replicas | 3 |
| Max age | 7d |
| Max bytes | 10 GB (tune) |
| Consumers | `broadcast-send` queue group |

---

## Summary table

| Tier | Batches | Peak queue | Send workers | Dominant bottleneck |
|------|---------|------------|--------------|---------------------|
| Starter 50k | 500 | 500 | 3–5 | Meta mps |
| Growth 300k | 3,000 | 3,000 | 10–20 | Meta mps |
| Enterprise 1M+ | 10,000+ | 10,000+ | 50–100 | Meta tier + audience resolve |

## Decision

Architecture scales horizontally for **workers** and **NATS**. **Meta WABA tier** is the primary capacity gate. Document tier limits in channel health metadata and surface in campaign send API estimates.
