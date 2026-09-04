// Package chattrace keeps short-lived, safe-to-display conversation tool traces.
// It deliberately has no MySQL fallback: tool execution payloads are transient
// operational data, not durable conversation history.
package chattrace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultTTL = time.Hour
	keyPrefix  = "ppt:chat-trace:"
)

// Event contains only the user-visible execution trail. Preview is already
// sanitized by the caller and must never contain model prompts or raw tool data.
type Event struct {
	ID        uint64         `json:"id"`
	SegmentID string         `json:"segment_id,omitempty"`
	Type      string         `json:"type"`
	Phase     string         `json:"phase,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	Detail    string         `json:"detail,omitempty"`
	Error     string         `json:"error,omitempty"`
	Preview   map[string]any `json:"preview,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type Store interface {
	Append(context.Context, string, Event) error
	Close() error
}

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewFromEnv connects only when CHAT_TRACE_REDIS_ADDR is configured. An empty
// setting disables persistence rather than accidentally writing execution data
// to another store. CHAT_TRACE_REDIS_TTL accepts a Go duration (default 1h).
func NewFromEnv(ctx context.Context) (Store, error) {
	addr := strings.TrimSpace(os.Getenv("CHAT_TRACE_REDIS_ADDR"))
	if addr == "" {
		return nil, nil
	}
	dbIndex := 0
	if raw := strings.TrimSpace(os.Getenv("CHAT_TRACE_REDIS_DB")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return nil, fmt.Errorf("invalid CHAT_TRACE_REDIS_DB")
		}
		dbIndex = value
	}
	ttl := defaultTTL
	if raw := strings.TrimSpace(os.Getenv("CHAT_TRACE_REDIS_TTL")); raw != "" {
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 {
			return nil, fmt.Errorf("invalid CHAT_TRACE_REDIS_TTL")
		}
		ttl = value
	}
	client := redis.NewClient(&redis.Options{Addr: addr, Password: os.Getenv("CHAT_TRACE_REDIS_PASSWORD"), DB: dbIndex})
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("chat trace redis unavailable: %w", err)
	}
	return &RedisStore{client: client, ttl: ttl}, nil
}

func (s *RedisStore) Append(ctx context.Context, taskID string, event Event) error {
	if s == nil || s.client == nil || strings.TrimSpace(taskID) == "" {
		return nil
	}
	event.CreatedAt = event.CreatedAt.UTC()
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	key := keyPrefix + taskID
	pipe := s.client.TxPipeline()
	pipe.RPush(ctx, key, payload)
	pipe.Expire(ctx, key, s.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}
