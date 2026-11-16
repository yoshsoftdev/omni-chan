# Event bus: stay on NATS JetStream or move to Kafka

## Status

Proposed — waiting for the product owner and architecture review (roadmap item 8.1).

## Date

2026-11-16

## Context

The roadmap (phase 8) plans a migration from NATS JetStream to Apache Kafka (or Redpanda). The tech-stack note S9 gives the reasons: long retention and replay, the Kafka ecosystem (Connect, CDC/Debezium, stream processing), and ordering per partition key. The PRD still lists this as open question 3: is the migration driven by a real need, or by team preference?

What the system runs on today:

| | Today |
|---|---|
| Streams | `WEBHOOK_EVENTS` (`meta.>`, Meta webhooks → backend consumers) and `BROADCAST` (`broadcast.>`, orchestrator → workers). |
| Retention | 7 days, file storage, `LimitsPolicy`. |
| Topology | One NATS 2.10 node, no replication (`docker-compose.yml`, `deploy/`). |
| Delivery guarantees | Durable consumers, `Nats-Msg-Id` server-side dedup, backoff and `MaxDeliver`, DLQ subjects, idempotency recorded in Postgres (`processed_events`, 6.2), wamid dedup in a claim table (ADR 0008). Broadcast publishes through a transactional outbox. |
| Observability | Consumer lag, redelivery, DLQ and stream size in Prometheus; alerts on lag and DLQ (4.3, 4.6). |
| Code footprint | About 15 Go files: `shared/nats` (7), webhook publisher, backend app wiring, broadcast orchestrator and worker. |

The reliability problems found in phase 0 (D2–D6 in `architecture/data-flow.md`) were all **configuration** faults: no retention limit, `Nak` without delay, unbounded redelivery, stream config never reconciled, and no message ids. All are fixed. None of them was a limit of NATS itself.

## Problem

Choose the event bus for the next 12–18 months against these criteria:

1. **Retention and replay.** How far back must events be replayable, and by whom?
2. **Ordering.** Messages of one conversation must be applied in order.
3. **Throughput headroom** for broadcast bursts.
4. **Operational cost** for a small team: brokers to run, upgrade, back up and monitor.
5. **Data integration.** Feeding a warehouse, CDC, or third-party consumers.
6. **Migration risk.** Phase 8's DoD is "no event lost during cutover".

## Decision

**Proposed: stay on NATS JetStream.** Phase 8 becomes "harden NATS" rather than "migrate", and Kafka is revisited only when one of the triggers below is met.

Concretely, if accepted:

1. Run NATS as a 3-node cluster with `Replicas: 3` on both streams. The single node is the real availability risk today, and every option below needs a highly available broker anyway.
2. Put retention per stream in config (it is a constant today) so it can grow without a code change.
3. Do **not** build the `shared/eventbus` abstraction (8.2) yet. It costs a refactor of every consumer, and it only pays off once a second implementation exists.
4. Close items 8.2–8.7 as "not needed now", and record the triggers in `architecture/tech-stack.md` S9.

**Triggers to reopen this ADR** (any one):

- A consumer outside the platform (warehouse, BI, partner) needs events, or CDC from Postgres is required.
- Required replay or retention exceeds about 30 days, or more than a few hundred GB.
- Sustained throughput above roughly 10k events/s, or a need for more than about 50 independent consumer groups on one stream.

## Alternatives Considered

### A. Apache Kafka (KRaft), as planned in phase 8

- **For:** long retention and replay are routine; per-key partition ordering; the largest ecosystem (Connect, Debezium, ksqlDB/Flink); widely known.
- **Against:** a new cluster to run (3 brokers for HA, partition planning, consumer-group rebalancing, JVM tuning), plus kafka-exporter and new alerts. Every consumer is rewritten, there is a dual-publish and shadow period per stream, and the "no event lost" cutover must be proven with daily reconciliation. On the facts above it solves no current problem: 7 days of retention is enough, ordering per conversation is already handled by idempotent consumers, and nothing outside the platform consumes events.
- **Rejected for now** because of the cost and risk, with no requirement behind it. It becomes the default if a trigger fires.

### B. Redpanda (Kafka API compatible)

- **For:** the Kafka API without a JVM, a single binary, simpler operations, good latency.
- **Against:** the same consumer rewrite and cutover risk as A. The vendor licence (BSL for some features) and a smaller community than Kafka.
- **Rejected for now** for the same reason as A. If a trigger fires, compare it against A at that time.

### C. Stay on NATS JetStream (proposed)

- **For:** already running, tuned and monitored. Phase 0 fixed its real faults, and idempotency and dedup now live in Postgres, independent of the bus. It is light to operate. JetStream already offers replay from a sequence or time, per-subject ordering, and server-side dedup.
- **Against:** a weaker ecosystem for analytics and CDC. Long retention is less common practice, and fewer engineers know it.

## Consequences

- **Positive:** no migration risk or freeze; effort goes to availability (cluster, replicas), which every option needs anyway. Phase 9 features can start sooner.
- **Negative:** if a trigger fires later, the migration is still ahead, and an `eventbus` abstraction has to be built then. Some tech-stack and data-flow text that assumes Kafka must be corrected.
- **Neutral:** the outbox, DB idempotency and DLQ conventions carry over unchanged to any future bus.

## Future Considerations

- Run the per-day count reconciliation (needed anyway for the phase 8 DoD) as a monitoring check now. It is useful on NATS too, and it becomes the cutover proof if Kafka comes later.
- If analytics needs arrive before a full migration, consider a one-way NATS → Kafka (or warehouse) bridge for that use only.

## Related ADR

- [ADR 0008: Message wamid dedup in the database](0008-message-wamid-dedup.md): dedup no longer depends on the bus.
- Roadmap phase 8 refers to this decision as "ADR-0008"; that number was taken, so it is ADR 0009.
