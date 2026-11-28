package sdkgo

import "time"

// Field is a structured key-value attribute attached to a log event.
type Field struct {
	Key   string
	Value string
}

// F constructs a Field. Shorthand for Field{Key: key, Value: value}.
//
//	client.Info("payment processed", sdkgo.F("order_id", "123"), sdkgo.F("amount", "99.00"))
func F(key, value string) Field {
	return Field{Key: key, Value: value}
}

// logEvent mirrors the JSON schema of the API's IngestLogInput.
type logEvent struct {
	Timestamp   time.Time         `json:"timestamp"`
	Environment string            `json:"environment,omitempty"`
	Level       string            `json:"level"`
	Service     string            `json:"service,omitempty"`
	Message     string            `json:"message"`
	TraceID     string            `json:"trace_id,omitempty"`
	SpanID      string            `json:"span_id,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

func (c *Client) log(level, msg string, fields []Field) {
	e := &logEvent{
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Message:     msg,
		Service:     c.cfg.Service,
		Environment: c.cfg.Environment,
	}
	if len(fields) > 0 {
		e.Attributes = make(map[string]string, len(fields))
		for _, f := range fields {
			e.Attributes[f.Key] = f.Value
		}
	}
	c.enqueue(e)
}

// Debug emits a log event at the debug level.
func (c *Client) Debug(msg string, fields ...Field) { c.log("debug", msg, fields) }

// Info emits a log event at the info level.
func (c *Client) Info(msg string, fields ...Field) { c.log("info", msg, fields) }

// Warn emits a log event at the warn level.
func (c *Client) Warn(msg string, fields ...Field) { c.log("warn", msg, fields) }

// Error emits a log event at the error level.
func (c *Client) Error(msg string, fields ...Field) { c.log("error", msg, fields) }

// Fatal emits a log event at the fatal level.
// Note: unlike standard loggers, this does NOT call os.Exit — it only records
// the event. Call os.Exit yourself if needed.
func (c *Client) Fatal(msg string, fields ...Field) { c.log("fatal", msg, fields) }
