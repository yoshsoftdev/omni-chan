# Runbook: NATS JetStream cluster

Decision: [ADR 0009](../adr/0009-event-bus.md). The event bus is a **3-node** NATS JetStream cluster. Both streams (`WEBHOOK_EVENTS`, `BROADCAST`) use `Replicas: 3`, so one node can be lost without stopping ingest or broadcast and without losing acknowledged events. This is covered by `shared/nats/cluster_test.go` and was checked on the compose cluster.

## Configuration

| Env (all services) | Value | Effect |
|---|---|---|
| `NATS_URL` | `nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222` | Clients fail over between nodes. |
| `NATS_STREAM_REPLICAS` | `3` | Applied to both streams at startup (`EnsureStream`). `0` leaves replicas unchanged. |
| `NATS_STREAM_MAX_AGE` | e.g. `168h` | Webhook stream retention. When set, it is applied to the existing stream too. Default 7 days. |
| `NATS_BROADCAST_STREAM_MAX_AGE` | e.g. `168h` | Broadcast stream retention, same rules. |

Every service must use the **same** values: each one reconciles the stream at startup.

- Local: `docker compose up` runs `nats-1..3`.
- Servers: use `deploy/nats/nats-server.conf` and `deploy/nats/nats-server.service` on three hosts.

## Health

- `GET :8222/healthz` on each node returns `ok`.
- `GET :8222/jsz` must show `meta_cluster.cluster_size = 3` and a leader.
- `nats stream info WEBHOOK_EVENTS`: one leader, two replicas with `current: true`.
- Grafana "Infra": consumer lag, redelivery and DLQ, as before.

## One node down

1. Nothing to do immediately: a follower becomes leader within seconds, and clients reconnect.
2. Bring the node back. It catches up on its own; confirm with `nats stream info` that replicas are `current` with lag 0.
3. **Two nodes down means no quorum**: publishes fail and webhooks get 5xx (Meta retries them). Restore one node first.

## Rolling restart or upgrade

One node at a time. `systemctl stop` sends SIGUSR2 (lame duck), so clients move away first. Wait for `/healthz` to return `ok` and the replicas to be current before the next node.

## Moving from the single node to the cluster

A standalone JetStream server does not turn its streams into replicated ones. Start the cluster fresh and drain the old node:

1. Bring up the three cluster nodes. Do not point services at them yet.
2. Check the old node is drained: consumer lag is 0 for the `WEBHOOK_EVENTS` consumers, and `broadcast_outbox` has no unpublished rows (see [broadcast recovery](broadcast-recovery.md)).
3. Switch every service to the cluster `NATS_URL` with `NATS_STREAM_REPLICAS=3`. Do it in one deploy, webhook service last. At startup the streams are created with 3 replicas.
4. During the switch, webhooks that fail with 5xx are retried by Meta, and the outbox republishes broadcast events. Consumers are idempotent (Postgres `processed_events`), so a redelivery is harmless.
5. Keep the old node read-only for 7 days (stream retention) in case a replay is needed, then remove it.

## When to revisit Kafka

See the triggers in ADR 0009: a consumer outside the platform or CDC, retention over about 30 days, or more than about 10k events/s.
