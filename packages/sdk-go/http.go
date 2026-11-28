package sdkgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const httpTimeout = 10 * time.Second

type batchPayload struct {
	Events []*logEvent `json:"events"`
}

type httpSender struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

func newHTTPSender(endpoint, apiKey string) *httpSender {
	return &httpSender{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: httpTimeout},
	}
}

func (s *httpSender) send(ctx context.Context, events []*logEvent) error {
	payload, err := json.Marshal(batchPayload{Events: events})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.endpoint+"/api/v1/logs/batch", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
