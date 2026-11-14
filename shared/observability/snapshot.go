package observability

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// snapshotCollector exposes an existing in-process counter set (a Snapshot()
// map) at scrape time, so call sites keep their current Inc* methods.
type snapshotCollector struct {
	namespace string
	snapshot  func() map[string]int64
}

// NewSnapshotCollector exports each key as <namespace>_<key>: keys ending in
// _total are counters, everything else a gauge. It is an unchecked collector
// because the key set is only known at scrape time.
func NewSnapshotCollector(namespace string, snapshot func() map[string]int64) prometheus.Collector {
	return snapshotCollector{namespace: namespace, snapshot: snapshot}
}

func (snapshotCollector) Describe(chan<- *prometheus.Desc) {}

func (c snapshotCollector) Collect(ch chan<- prometheus.Metric) {
	for key, value := range c.snapshot() {
		kind := prometheus.GaugeValue
		if strings.HasSuffix(key, "_total") {
			kind = prometheus.CounterValue
		}
		desc := prometheus.NewDesc(c.namespace+"_"+key, "In-process counter "+key+".", nil, nil)
		ch <- prometheus.MustNewConstMetric(desc, kind, float64(value))
	}
}

// labeledCounterCollector exports one counter per snapshot key, with the key as
// a label value (e.g. failures by code). Keys must come from a small fixed set.
type labeledCounterCollector struct {
	desc     *prometheus.Desc
	snapshot func() map[string]int64
}

// NewLabeledCounterCollector exports name{label="<key>"} = value for each key.
func NewLabeledCounterCollector(name, help, label string, snapshot func() map[string]int64) prometheus.Collector {
	return labeledCounterCollector{desc: prometheus.NewDesc(name, help, []string{label}, nil), snapshot: snapshot}
}

func (c labeledCounterCollector) Describe(ch chan<- *prometheus.Desc) { ch <- c.desc }

func (c labeledCounterCollector) Collect(ch chan<- prometheus.Metric) {
	for key, value := range c.snapshot() {
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.CounterValue, float64(value), key)
	}
}
