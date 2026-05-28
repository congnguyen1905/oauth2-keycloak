package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"auth-service/pkg/models"
)

// Storage defines the interface for session storage
type Storage interface {
	Save(ctx context.Context, session *models.UserSession, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (*models.UserSession, error)
	Delete(ctx context.Context, sessionID string) error
	Exists(ctx context.Context, sessionID string) (bool, error)
	UpdateTTL(ctx context.Context, sessionID string, ttl time.Duration) error
	Close() error
}

// RedisStorage implements Storage using Redis
type RedisStorage struct {
	client *redis.Client
	prefix string
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(addr, password string, db int) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	
	return &RedisStorage{
		client: client,
		prefix: "session:",
	}
}

// Save saves a session to Redis
func (r *RedisStorage) Save(ctx context.Context, session *models.UserSession, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	key := r.prefix + session.SessionID
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a session from Redis
func (r *RedisStorage) Get(ctx context.Context, sessionID string) (*models.UserSession, error) {
	key := r.prefix + sessionID
	
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	var session models.UserSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	return &session, nil
}

// Delete removes a session from Redis
func (r *RedisStorage) Delete(ctx context.Context, sessionID string) error {
	key := r.prefix + sessionID
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a session exists in Redis
func (r *RedisStorage) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := r.prefix + sessionID
	
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	
	return result > 0, nil
}

// UpdateTTL updates the TTL of a session
func (r *RedisStorage) UpdateTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := r.prefix + sessionID
	return r.client.Expire(ctx, key, ttl).Err()
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// TestConnection tests the Redis connection
func (r *RedisStorage) TestConnection(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
