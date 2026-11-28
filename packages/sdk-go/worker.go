package sdkgo

import (
	"context"
	"fmt"
	"os"
	"time"
)

func (c *Client) worker() {
	defer c.wg.Done()

	var batch []*logEvent
	ticker := time.NewTicker(c.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-c.ch:
			batch = append(batch, event)
			if len(batch) >= c.cfg.BatchSize {
				c.flush(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				c.flush(batch)
				batch = batch[:0]
			}

		case <-c.done:
			// drain remaining events before exit
			for len(c.ch) > 0 {
				batch = append(batch, <-c.ch)
			}
			if len(batch) > 0 {
				c.flush(batch)
			}
			return
		}
	}
}

func (c *Client) flush(batch []*logEvent) {
	if err := c.sender.send(context.Background(), batch); err != nil {
		fmt.Fprintf(os.Stderr, "logstorm: flush failed: %v\n", err)
	}
}
