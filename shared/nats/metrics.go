package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
)

// DLQSubjectFilters match every dead-letter subject on the webhook stream.
var DLQSubjectFilters = []string{"meta.dlq.>", "meta.consumer.dlq.>"}

// JetStreamCollector reads consumer lag and dead-letter depth from the server
// at scrape time, so the numbers are true for the whole cluster, not one process.
type JetStreamCollector struct {
	js        nats.JetStreamContext
	stream    string
	consumers []string

	pending, ackPending, redelivered, dlq, up, bytes, maxBytes *prometheus.Desc
}

// NewJetStreamCollector watches the given durable consumers of stream.
func NewJetStreamCollector(js nats.JetStreamContext, stream string, consumers ...string) *JetStreamCollector {
	labels := []string{"stream", "consumer"}
	return &JetStreamCollector{
		js: js, stream: stream, consumers: consumers,
		pending:     prometheus.NewDesc("jetstream_consumer_pending_messages", "Messages not yet delivered to the consumer (lag).", labels, nil),
		ackPending:  prometheus.NewDesc("jetstream_consumer_ack_pending_messages", "Messages delivered but not yet acknowledged.", labels, nil),
		redelivered: prometheus.NewDesc("jetstream_consumer_redelivered_messages", "Messages currently being redelivered.", labels, nil),
		dlq:         prometheus.NewDesc("jetstream_dlq_messages", "Messages stored on a dead-letter subject.", []string{"stream", "subject"}, nil),
		up:          prometheus.NewDesc("jetstream_up", "1 when the stream could be read at scrape time.", []string{"stream"}, nil),
		bytes:       prometheus.NewDesc("jetstream_stream_bytes", "Bytes stored in the stream.", []string{"stream"}, nil),
		maxBytes:    prometheus.NewDesc("jetstream_stream_max_bytes", "Configured stream size limit (only when one is set).", []string{"stream"}, nil),
	}
}

func (c *JetStreamCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range []*prometheus.Desc{c.pending, c.ackPending, c.redelivered, c.dlq, c.up, c.bytes, c.maxBytes} {
		ch <- d
	}
}

func (c *JetStreamCollector) Collect(ch chan<- prometheus.Metric) {
	up := 1.0
	if info, err := c.js.StreamInfo(c.stream); err == nil {
		ch <- prometheus.MustNewConstMetric(c.bytes, prometheus.GaugeValue, float64(info.State.Bytes), c.stream)
		if info.Config.MaxBytes > 0 {
			ch <- prometheus.MustNewConstMetric(c.maxBytes, prometheus.GaugeValue, float64(info.Config.MaxBytes), c.stream)
		}
	} else {
		up = 0
	}
	for _, name := range c.consumers {
		info, err := c.js.ConsumerInfo(c.stream, name)
		if err != nil {
			continue // consumer not created yet (no subscriber started)
		}
		ch <- prometheus.MustNewConstMetric(c.pending, prometheus.GaugeValue, float64(info.NumPending), c.stream, name)
		ch <- prometheus.MustNewConstMetric(c.ackPending, prometheus.GaugeValue, float64(info.NumAckPending), c.stream, name)
		ch <- prometheus.MustNewConstMetric(c.redelivered, prometheus.GaugeValue, float64(info.NumRedelivered), c.stream, name)
	}
	for _, filter := range DLQSubjectFilters {
		info, err := c.js.StreamInfo(c.stream, &nats.StreamInfoRequest{SubjectsFilter: filter})
		if err != nil {
			up = 0
			continue
		}
		for subject, n := range info.State.Subjects {
			ch <- prometheus.MustNewConstMetric(c.dlq, prometheus.GaugeValue, float64(n), c.stream, subject)
		}
	}
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up, c.stream)
}
