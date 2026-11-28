package sdkgo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	sdkgo "github.com/NewHorizonIT/logstorm-sdk-go"
)

// receivedBatch mirrors the JSON sent by the SDK.
type receivedBatch struct {
	Events []struct {
		Level       string            `json:"level"`
		Message     string            `json:"message"`
		Service     string            `json:"service"`
		Environment string            `json:"environment"`
		Attributes  map[string]string `json:"attributes"`
	} `json:"events"`
}

func newTestServer(t *testing.T, calls *atomic.Int32, batches *[]receivedBatch) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var b receivedBatch
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			t.Errorf("decode body: %v", err)
		}
		*batches = append(*batches, b)
		w.WriteHeader(http.StatusAccepted)
	}))
}

func mustNew(t *testing.T, apiKey string, opts ...sdkgo.Option) *sdkgo.Client {
	t.Helper()
	c, err := sdkgo.New(apiKey, opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNew_defaults(t *testing.T) {
	client := mustNew(t, "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestNew_errorOnEmptyAPIKey(t *testing.T) {
	_, err := sdkgo.New("")
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
}

func TestNew_errorOnInvalidBatchSize(t *testing.T) {
	cases := []int{0, -1, 1001}
	for _, n := range cases {
		_, err := sdkgo.New("key", sdkgo.WithBatchSize(n))
		if err == nil {
			t.Errorf("expected error for batch size %d", n)
		}
	}
}

func TestNew_errorOnInvalidFlushInterval(t *testing.T) {
	cases := []time.Duration{50 * time.Millisecond, 2 * time.Hour}
	for _, d := range cases {
		_, err := sdkgo.New("key", sdkgo.WithFlushInterval(d))
		if err == nil {
			t.Errorf("expected error for flush interval %s", d)
		}
	}
}

func TestNew_errorOnInvalidChannelSize(t *testing.T) {
	cases := []int{0, -1, 100_001}
	for _, n := range cases {
		_, err := sdkgo.New("key", sdkgo.WithChannelSize(n))
		if err == nil {
			t.Errorf("expected error for channel size %d", n)
		}
	}
}

func TestClient_flushesOnBatchSize(t *testing.T) {
	var calls atomic.Int32
	var batches []receivedBatch

	srv := newTestServer(t, &calls, &batches)
	defer srv.Close()

	client := mustNew(t, "key",
		sdkgo.WithEndpoint(srv.URL),
		sdkgo.WithBatchSize(3),
		sdkgo.WithFlushInterval(10*time.Second),
	)

	client.Info("msg 1")
	client.Warn("msg 2")
	client.Error("msg 3")

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client.Shutdown(ctx)

	if calls.Load() == 0 {
		t.Fatal("expected at least one HTTP call")
	}
	total := 0
	for _, b := range batches {
		total += len(b.Events)
	}
	if total != 3 {
		t.Fatalf("expected 3 events total, got %d", total)
	}
}

func TestClient_flushesOnTicker(t *testing.T) {
	var calls atomic.Int32
	var batches []receivedBatch

	srv := newTestServer(t, &calls, &batches)
	defer srv.Close()

	client := mustNew(t, "key",
		sdkgo.WithEndpoint(srv.URL),
		sdkgo.WithBatchSize(100),
		sdkgo.WithFlushInterval(100*time.Millisecond),
	)

	client.Info("ticker event")

	time.Sleep(300 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client.Shutdown(ctx)

	if calls.Load() == 0 {
		t.Fatal("expected ticker to trigger a flush")
	}
}

func TestClient_shutdownDrainsChannel(t *testing.T) {
	var calls atomic.Int32
	var batches []receivedBatch

	srv := newTestServer(t, &calls, &batches)
	defer srv.Close()

	client := mustNew(t, "key",
		sdkgo.WithEndpoint(srv.URL),
		sdkgo.WithBatchSize(100),
		sdkgo.WithFlushInterval(10*time.Second),
	)

	for i := 0; i < 5; i++ {
		client.Debug("shutdown drain test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	total := 0
	for _, b := range batches {
		total += len(b.Events)
	}
	if total != 5 {
		t.Fatalf("expected 5 events drained on shutdown, got %d", total)
	}
}

func TestClient_defaultServiceAndEnvironment(t *testing.T) {
	var calls atomic.Int32
	var batches []receivedBatch

	srv := newTestServer(t, &calls, &batches)
	defer srv.Close()

	client := mustNew(t, "key",
		sdkgo.WithEndpoint(srv.URL),
		sdkgo.WithService("payments"),
		sdkgo.WithEnvironment("staging"),
		sdkgo.WithBatchSize(1),
	)

	client.Info("test")
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client.Shutdown(ctx)

	if len(batches) == 0 || len(batches[0].Events) == 0 {
		t.Fatal("no events received")
	}
	e := batches[0].Events[0]
	if e.Service != "payments" {
		t.Errorf("service: got %q, want %q", e.Service, "payments")
	}
	if e.Environment != "staging" {
		t.Errorf("environment: got %q, want %q", e.Environment, "staging")
	}
}

func TestClient_enqueueDropsWhenFull(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	client := mustNew(t, "key",
		sdkgo.WithEndpoint(srv.URL),
		sdkgo.WithChannelSize(1),
		sdkgo.WithBatchSize(1),
		sdkgo.WithFlushInterval(10*time.Second),
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			client.Info("flood")
		}
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("enqueue blocked the caller")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	client.Shutdown(ctx)
}
