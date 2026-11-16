package nats

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// freePort reserves a TCP port for a cluster route (routes need fixed ports).
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// startCluster runs a 3-node JetStream cluster in-process, like deploy/nats (ADR 0009).
func startCluster(t *testing.T) []*natsserver.Server {
	t.Helper()
	routePorts := []int{freePort(t), freePort(t), freePort(t)}
	var routes []*url.URL
	for _, p := range routePorts {
		routes = append(routes, &url.URL{Scheme: "nats", Host: "127.0.0.1:" + strconv.Itoa(p)})
	}
	servers := make([]*natsserver.Server, 3)
	for i := range servers {
		srv, err := natsserver.NewServer(&natsserver.Options{
			ServerName: fmt.Sprintf("n%d", i+1),
			Host:       "127.0.0.1",
			Port:       -1,
			JetStream:  true,
			StoreDir:   t.TempDir(),
			NoLog:      true,
			NoSigs:     true,
			Cluster:    natsserver.ClusterOpts{Name: "omni", Host: "127.0.0.1", Port: routePorts[i]},
			Routes:     routes,
		})
		if err != nil {
			t.Fatal(err)
		}
		go srv.Start()
		servers[i] = srv
		t.Cleanup(srv.Shutdown)
	}
	for _, srv := range servers {
		if !srv.ReadyForConnections(10 * time.Second) {
			t.Fatal("cluster node not ready")
		}
	}
	return servers
}

func clusterURLs(servers []*natsserver.Server) string {
	urls := ""
	for i, s := range servers {
		if i > 0 {
			urls += ","
		}
		urls += s.ClientURL()
	}
	return urls
}

// retry runs fn until it succeeds or the deadline passes (leader elections take a moment).
func retry(t *testing.T, what string, fn func() error) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		err := fn()
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s: %v", what, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// 8.3/8.4: with Replicas 3, losing the stream leader keeps publish and consume working
// and loses no acknowledged message.
func TestClusterStreamSurvivesLeaderLoss(t *testing.T) {
	servers := startCluster(t)
	nc, err := nats.Connect(clusterURLs(servers), nats.MaxReconnects(-1), nats.ReconnectWait(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(nc.Close)
	js, _ := nc.JetStream(nats.MaxWait(2 * time.Second))

	cfg := testNATSConfig("WEBHOOK_EVENTS")
	cfg.StreamReplicas = 3
	retry(t, "ensure stream", func() error { return EnsureStream(js, WebhookStreamSpec(cfg), quietLog()) })

	info, err := js.StreamInfo("WEBHOOK_EVENTS")
	if err != nil {
		t.Fatal(err)
	}
	if info.Config.Replicas != 3 || info.Cluster == nil || len(info.Cluster.Replicas) != 2 {
		t.Fatalf("want 3 replicas (leader + 2), got config=%d cluster=%+v", info.Config.Replicas, info.Cluster)
	}

	publish := func(i int) {
		retry(t, fmt.Sprintf("publish %d", i), func() error {
			_, err := js.Publish("meta.message.received", []byte(strconv.Itoa(i)), nats.MsgId(strconv.Itoa(i)))
			return err
		})
	}
	for i := 0; i < 50; i++ {
		publish(i)
	}

	// Kill the stream leader; a follower must take over.
	leader := info.Cluster.Leader
	for _, s := range servers {
		if s.Name() == leader {
			s.Shutdown()
		}
	}
	for i := 50; i < 100; i++ {
		publish(i)
	}

	retry(t, "new leader with all messages", func() error {
		info, err := js.StreamInfo("WEBHOOK_EVENTS")
		if err != nil {
			return err
		}
		if info.Cluster.Leader == "" || info.Cluster.Leader == leader {
			return fmt.Errorf("no new leader yet (%q)", info.Cluster.Leader)
		}
		if info.State.Msgs != 100 {
			return fmt.Errorf("messages = %d, want 100", info.State.Msgs)
		}
		return nil
	})

	// A durable consumer created after the failover still reads every message exactly once.
	var sub *nats.Subscription
	retry(t, "subscribe", func() error {
		var err error
		sub, err = js.PullSubscribe("meta.>", "failover-check", nats.BindStream("WEBHOOK_EVENTS"))
		return err
	})
	seen := map[string]bool{}
	retry(t, "consume all", func() error {
		msgs, _ := sub.Fetch(100, nats.MaxWait(time.Second))
		for _, m := range msgs {
			seen[string(m.Data)] = true
			_ = m.Ack()
		}
		if len(seen) != 100 {
			return fmt.Errorf("consumed %d distinct messages, want 100", len(seen))
		}
		return nil
	})
}

// Replicas can be raised on an existing single-replica stream (rolling out the cluster).
func TestEnsureStreamRaisesReplicasOnCluster(t *testing.T) {
	servers := startCluster(t)
	nc, err := nats.Connect(clusterURLs(servers))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(nc.Close)
	js, _ := nc.JetStream(nats.MaxWait(2 * time.Second))

	cfg := testNATSConfig("WEBHOOK_EVENTS")
	retry(t, "create R1", func() error { return EnsureStream(js, WebhookStreamSpec(cfg), quietLog()) })
	cfg.StreamReplicas = 3
	retry(t, "raise to R3", func() error { return EnsureStream(js, WebhookStreamSpec(cfg), quietLog()) })
	retry(t, "replicas in place", func() error {
		info, err := js.StreamInfo("WEBHOOK_EVENTS")
		if err != nil {
			return err
		}
		if info.Config.Replicas != 3 {
			return fmt.Errorf("replicas = %d", info.Config.Replicas)
		}
		return nil
	})
}
