package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const verificationKeyPrefix = "ppt:email-verification:"

// verificationCodeStore holds short-lived email verification codes. It is
// intentionally separate from the durable account database.
type verificationCodeStore interface {
	Issue(ctx context.Context, email, code string, ttl time.Duration, maxAttempts int) (bool, error)
	Consume(ctx context.Context, email, code string) (bool, error)
	Close() error
}

type redisVerificationCodeStore struct {
	client *redis.Client
}

var (
	verificationStoreMu sync.RWMutex
	verificationStore   verificationCodeStore
)

var issueVerificationCodeScript = redis.NewScript(`
local attempts = redis.call('GET', KEYS[2])
if attempts and tonumber(attempts) >= tonumber(ARGV[3]) then
  return 0
end
local count = redis.call('INCR', KEYS[2])
if count == 1 then
  redis.call('EXPIRE', KEYS[2], ARGV[2])
end
redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])
return 1
`)

var consumeVerificationCodeScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  redis.call('DEL', KEYS[1])
  return 1
end
return 0
`)

// InitVerificationStoreFromEnv initializes the Redis-only verification-code
// store. EMAIL_VERIFICATION_REDIS_ADDR is required in web mode; password and
// database index use the matching EMAIL_VERIFICATION_REDIS_* variables.
func InitVerificationStoreFromEnv(ctx context.Context) error {
	store, err := newRedisVerificationCodeStoreFromEnv(ctx)
	if err != nil {
		return err
	}
	verificationStoreMu.Lock()
	previous := verificationStore
	verificationStore = store
	verificationStoreMu.Unlock()
	if previous != nil {
		_ = previous.Close()
	}
	return nil
}

// CloseVerificationStore releases the Redis connection during shutdown.
func CloseVerificationStore() error {
	verificationStoreMu.Lock()
	store := verificationStore
	verificationStore = nil
	verificationStoreMu.Unlock()
	if store == nil {
		return nil
	}
	return store.Close()
}

func newRedisVerificationCodeStoreFromEnv(ctx context.Context) (*redisVerificationCodeStore, error) {
	addr := strings.TrimSpace(os.Getenv("EMAIL_VERIFICATION_REDIS_ADDR"))
	if addr == "" {
		return nil, errors.New("EMAIL_VERIFICATION_REDIS_ADDR is not configured")
	}
	dbIndex := 0
	if raw := strings.TrimSpace(os.Getenv("EMAIL_VERIFICATION_REDIS_DB")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return nil, errors.New("invalid EMAIL_VERIFICATION_REDIS_DB")
		}
		dbIndex = value
	}
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("EMAIL_VERIFICATION_REDIS_PASSWORD"),
		DB:           dbIndex,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("email verification redis unavailable: %w", err)
	}
	return &redisVerificationCodeStore{client: client}, nil
}

func (s *redisVerificationCodeStore) Issue(ctx context.Context, email, code string, ttl time.Duration, maxAttempts int) (bool, error) {
	result, err := issueVerificationCodeScript.Run(ctx, s.client, []string{verificationCodeKey(email), verificationAttemptsKey(email)}, code, int(ttl.Seconds()), maxAttempts).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (s *redisVerificationCodeStore) Consume(ctx context.Context, email, code string) (bool, error) {
	result, err := consumeVerificationCodeScript.Run(ctx, s.client, []string{verificationCodeKey(email)}, code).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (s *redisVerificationCodeStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

func currentVerificationStore() verificationCodeStore {
	verificationStoreMu.RLock()
	defer verificationStoreMu.RUnlock()
	return verificationStore
}

func verificationCodeKey(email string) string {
	return verificationKeyPrefix + "code:" + verificationEmailDigest(email)
}

func verificationAttemptsKey(email string) string {
	return verificationKeyPrefix + "attempts:" + verificationEmailDigest(email)
}

func verificationEmailDigest(email string) string {
	digest := sha256.Sum256([]byte(email))
	return fmt.Sprintf("%x", digest[:])
}

// Ensure the concrete Redis store keeps the shutdown contract explicit.
var _ io.Closer = (*redisVerificationCodeStore)(nil)
