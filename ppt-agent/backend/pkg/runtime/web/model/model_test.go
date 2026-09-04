package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestModelCredentialCannotLeakThroughJSON(t *testing.T) {
	data, err := json.Marshal(ModelCredential{Provider: "deepseek", APIKey: "secret-value"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret-value") || strings.Contains(string(data), "api_key") {
		t.Fatalf("credential serialized secret: %s", data)
	}
}
