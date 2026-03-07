package pkg

import "encoding/json"

func JsonToBytes(logs any) []byte {
	logUmarsal, err := json.Marshal(logs)
	if err != nil {
		return nil
	}
	return logUmarsal
}

func BytesToJson(data []byte, logs any) error {
	err := json.Unmarshal(data, logs)
	if err != nil {
		return err
	}
	return nil
}
