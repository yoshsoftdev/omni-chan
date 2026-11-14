# Message wamid dedup in the database

## Status

Accepted

## Date

2026-11-15

## Context

`messages` is partitioned by `created_at` (migration `000015`) but has only the DEFAULT partition. A unique index on a partitioned table must include the partition key, so the existing `UNIQUE (wamid, created_at)` never rejects a second row with the same wamid, because `created_at` differs. Dedup relied on check-then-insert (`FindByWAMID` → `Create`) plus a 24-hour Redis key. Concurrent redeliveries, replicas, or a Redis outage could store the same message twice (risk D7 in `architecture/data-flow.md`).

An integration test confirmed this: 16 concurrent inserts of one wamid produced 16 rows.

Options considered in the roadmap (item 6.1):

1. Monthly partitions with `pg_partman`, plus a claim table `message_wamids(team_id, wamid)`.
2. Remove partitioning and put a plain `UNIQUE (team_id, wamid)` on `messages`.
3. A claim table only, keeping the current partitioning as is.

## Decision

Option 3. Add `message_wamids (team_id, wamid) PRIMARY KEY` (migration `000053`). `message.Repository.Create` inserts the claim and the message **in one transaction**. If the claim already exists, `Create` returns `repository.ErrConflict` and writes nothing. Messages without a wamid are not claimed.

Callers treat `ErrConflict` as "already stored":

- Inbound/echo ingest re-reads the existing row (behaviour unchanged).
- History import (onboarding, channelsync) skips the message.
- Outbound and automation sends, which are already delivered to Meta, return or keep the existing row instead of failing.

## Consequences

- The database guarantees one message per (team, wamid) regardless of Redis, restarts, or replica count. Redis remains a fast-path cache only.
- Additive and reversible: no rewrite of `messages`, and the down migration drops the table. Partitioning stays available if volume ever needs it.
- `pg_partman` is not adopted now. With one DEFAULT partition, monthly partitions bring an extension to operate but no benefit at current volume. Revisit when `messages` reaches tens of millions of rows.
- Duplicates that already existed before `000053` stay in `messages`. The backfill claims the earliest row per (team, wamid).
- Hard-deleting a message leaves its claim behind, so the same wamid cannot be re-inserted. This is intended: Meta never reuses wamids.
