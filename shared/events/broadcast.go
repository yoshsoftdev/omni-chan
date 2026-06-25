package events

// Broadcast NATS subjects (JetStream stream: BROADCAST).
const (
	SubjectBroadcastJobCreated      = "broadcast.job.created"
	SubjectBroadcastAudienceResolved = "broadcast.audience.resolved"
	SubjectBroadcastBatchSend       = "broadcast.batch.send"
	SubjectBroadcastBatchRetry      = "broadcast.batch.retry"
	SubjectBroadcastJobCompleted    = "broadcast.job.completed"
)

// JobCreatedPayload starts orchestration for a broadcast job.
type JobCreatedPayload struct {
	JobID      string `json:"job_id"`
	CampaignID string `json:"campaign_id"`
	TeamID     string `json:"team_id"`
	ChannelID  string `json:"channel_id"`
}

// AudienceResolvedPayload signals audience resolution completed.
type AudienceResolvedPayload struct {
	JobID          string `json:"job_id"`
	CampaignID     string `json:"campaign_id"`
	RecipientCount int    `json:"recipient_count"`
}

// BatchSendPayload is the work unit for send workers (no recipient payloads).
type BatchSendPayload struct {
	JobID          string `json:"job_id"`
	BatchID        string `json:"batch_id"`
	CampaignID     string `json:"campaign_id"`
	TeamID         string `json:"team_id"`
	ChannelID      string `json:"channel_id"`
	RecipientCount int    `json:"recipient_count"`
}
