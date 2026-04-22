package utils

import (
	"encoding/json"
)

func MarshalJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func UnmarshalJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
