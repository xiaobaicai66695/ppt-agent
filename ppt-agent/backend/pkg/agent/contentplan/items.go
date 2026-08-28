package contentplan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

var titleKeys = []string{"title", "label", "name", "key", "heading"}
var bodyKeys = []string{"text", "value", "content", "detail", "body"}

// DecodeItems accepts the common shapes produced by models and normalizes
// every meaningful value into the string list consumed by slide generators.
func DecodeItems(data []byte) ([]string, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return nil, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode items: %w", err)
	}

	values, ok := value.([]any)
	if !ok {
		values = []any{value}
	}
	result := make([]string, 0, len(values))
	for _, item := range values {
		if normalized := normalizeValue(item); normalized != "" {
			result = append(result, normalized)
		}
	}
	return result, nil
}

func normalizeValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return fmt.Sprintf("%g", typed)
	case bool:
		return fmt.Sprintf("%t", typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, nested := range typed {
			if text := normalizeValue(nested); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " / ")
	case map[string]any:
		return normalizeObject(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func normalizeObject(value map[string]any) string {
	title := firstObjectValue(value, titleKeys)
	body := firstObjectValue(value, bodyKeys)
	if title != "" && body != "" && title != body {
		return title + ": " + body
	}
	if title != "" {
		return title
	}
	if body != "" {
		return body
	}

	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if text := normalizeValue(value[key]); text != "" {
			parts = append(parts, key+": "+text)
		}
	}
	return strings.Join(parts, "; ")
}

func firstObjectValue(value map[string]any, keys []string) string {
	for _, key := range keys {
		if field, ok := value[key]; ok {
			if text := normalizeValue(field); text != "" {
				return text
			}
		}
	}
	return ""
}
