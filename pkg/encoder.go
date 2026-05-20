package pkg

import "encoding/json"

func JsonToBytes(v any) ([]byte, error) {
	return json.Marshal(v)
}

func BytesToJson(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
