package natsclient

import (
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestConsumerGroupNaming(t *testing.T) {
	got, err := ConsumerGroup("thing-events")
	if err != nil {
		t.Fatal(err)
	}
	if got != "registry-thing-events" {
		t.Fatalf("got %q", got)
	}

	got, err = ConsumerGroup("registry-thing-events")
	if err != nil {
		t.Fatal(err)
	}
	if got != "registry-thing-events" {
		t.Fatalf("already prefixed: got %q", got)
	}

	if _, err := ConsumerGroup("dispatch-thing-events"); err == nil {
		t.Fatal("dispatch prefix should be rejected")
	}
	if _, err := ConsumerGroup(""); err == nil {
		t.Fatal("empty name should be rejected")
	}
	if _, err := ConsumerGroup("Thing_Events"); err == nil {
		t.Fatal("uppercase and underscore should be rejected")
	}
}

func TestPublishAndConsumeCoreAndJetStream(t *testing.T) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://127.0.0.1:4224"
	}
	c := Connect(url)
	defer c.Close()
	deadline := time.Now().Add(3 * time.Second)
	for !c.Connected() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if !c.Connected() {
		t.Skipf("nats is not reachable at %s", url)
	}

	coreSubject := "registry.core.reg05"
	coreBody := []byte("core-" + time.Now().Format(time.RFC3339Nano))
	coreGot := make(chan []byte, 1)
	coreSub, err := c.SubscribeCore(coreSubject, "reg05-core", func(msg *nats.Msg) {
		coreGot <- append([]byte(nil), msg.Data...)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer coreSub.Unsubscribe()
	if coreSub.Queue != "registry-reg05-core" {
		t.Fatalf("core queue group = %q", coreSub.Queue)
	}
	if err := c.PublishCore(coreSubject, coreBody); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-coreGot:
		if string(got) != string(coreBody) {
			t.Fatalf("core body = %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("core subscriber did not receive the message")
	}

	jsSubject := "registry.thing.created"
	jsBody := []byte("js-" + time.Now().Format(time.RFC3339Nano))
	jsGot := make(chan []byte, 1)
	jsSub, err := c.SubscribeJetStream(jsSubject, "reg05-durable", func(msg *nats.Msg) {
		jsGot <- append([]byte(nil), msg.Data...)
		_ = msg.Ack()
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		c.mu.RLock()
		js := c.js
		c.mu.RUnlock()
		if js != nil {
			_ = js.DeleteConsumer(Stream, "registry-reg05-durable")
		}
		_ = jsSub.Unsubscribe()
	}()
	info, err := jsSub.ConsumerInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "registry-reg05-durable" {
		t.Fatalf("durable consumer = %q", info.Name)
	}
	if err := c.PublishJetStream(jsSubject, jsBody); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-jsGot:
		if string(got) != string(jsBody) {
			t.Fatalf("jetstream body = %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("jetstream consumer did not receive the message")
	}
}
