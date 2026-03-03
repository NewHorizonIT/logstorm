package encoder

import (
	"encoding/json"

	"github.com/NewHorizonIT/logstorm-ingestion/internal/model"
)

type Encoder interface {
	Encode(model.LogEvent) ([]byte, error)
}

type JSONEncoder struct{}

func (j JSONEncoder) Encode(e model.LogEvent) ([]byte, error) {
	return json.Marshal(e)
}
