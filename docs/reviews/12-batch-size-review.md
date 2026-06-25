# Batch Size Strategy Review — Phase 12.0

**Status:** Design review  
**Default in Phase 8 spec:** 100 recipients

## Candidates evaluated

| Batch size | Meta throughput | Memory | Retry impact | Worker concurrency |
|------------|-----------------|--------|--------------|-------------------|
| **50** | Low burst; smooth under rate limit | ~50 × payload in memory | Small blast radius | More NATS messages (2× vs 100) |
| **100** | Balanced for tier-2 WABA (~80 mps) | ~100 template builds in flight per worker | Moderate | **Recommended default** |
| **250** | Higher burst risk | ~250 × variable map + HTTP clients | Large retry unit on crash | Fewer messages |
| **500** | Likely to hit 429 without strict limiter | High heap per batch | Re-send 500 on transient failure | Poor fit v1 |
| **1000** | Exceeds typical per-second limits | OOM risk on variable resolution | Unacceptable retry cost | Not recommended |

## Meta API throughput constraints

WhatsApp Cloud API sends **one template message per HTTP request**. Throughput is bounded by:

| Limit | Typical value |
|-------|---------------|
| WABA messaging tier | 80–1,000+ messages/second (tier-dependent) |
| Practical per `phone_number_id` | Often 80–250 mps before 429 |
| Error 130429 | Rate limit — batch should defer, not fail |

**Implication:** Batch size does not increase Meta throughput; it only affects **worker parallelism granularity** and **retry scope**.

## Memory usage estimate (per send worker)

Per recipient in batch (in memory during processing):

| Component | Approx size |
|-----------|-------------|
| Recipient row + contact fields | ~1–2 KB |
| Template variable resolution | ~2–5 KB |
| HTTP request/response | ~5–10 KB |

| Batch size | Peak per worker (rough) |
|------------|-------------------------|
| 50 | ~0.5 MB |
| 100 | ~1 MB |
| 250 | ~2.5 MB |
| 500 | ~5 MB |
| 1000 | ~10 MB |

Memory is not the primary constraint; **rate limits and retry blast radius** are.

## Retry impact

On worker crash mid-batch, all non-terminal recipients in `sending` must be reclaimed.

| Batch size | Max recipients to re-evaluate |
|------------|--------------------------------|
| 50 | 50 |
| 100 | 100 |
| 250 | 250 |
| 500+ | Unacceptable duplicate-send risk without strict idempotency |

Idempotency guard: `WHERE status IN ('pending','queued','sending')` before Meta call; skip if `wamid` already set.

## Worker concurrency interaction

```
effective_mps ≈ min(waba_limit, workers × (batch_size / batch_duration))
```

Example: 100 recipients/batch, 2s/batch processing, 10 workers → ~500 mps theoretical — **rate limiter must cap below WABA tier**.

| Strategy | Setting |
|----------|---------|
| Default batch size | **100** |
| Tunable range | 50–250 via config `BROADCAST_BATCH_SIZE` |
| Per-tier override | Enterprise workspaces: 250 with dedicated channel limiter |
| Never in v1 | 500, 1000 |

## Tier recommendations

| Workspace tier | Contacts/campaign | Batch size | Notes |
|----------------|-------------------|------------|-------|
| Starter | 50k | 100 | 500 batches — manageable |
| Growth | 300k | 100–150 | 2,000–3,000 batches |
| Enterprise | 1M+ | 100 (default), 250 optional | Requires channel affinity + strict rate limit |

## Decision

| Item | Selection |
|------|-----------|
| **Default batch size** | **100** |
| Config key | `BROADCAST_BATCH_SIZE` (env), validated 50–250 |
| Planner behavior | Last batch may be smaller; `recipient_count` on batch row is authoritative |
