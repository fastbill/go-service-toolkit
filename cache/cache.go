package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// ErrNotFound is returned when the key was not found in the cache.
var ErrNotFound = errors.New("key not found in cache")

// Cache defines basic cache operartions including methods for setting and getting JSON objects.
type Cache interface {
	Prefix() string
	Set(key string, value string, expiration time.Duration) error
	Get(key string) (string, error)
	SetBool(key string, value bool, expiration time.Duration) error
	GetBool(key string) (bool, error)
	SetJSON(key string, value interface{}, expiration time.Duration) error
	GetJSON(key string, result interface{}) error
	Del(key string) error
	Close() error
}

// RedisClient wraps the REDIS client to provide an implementation of the Cache interface.
// It allows defining a prefix that is applied to the key for all operations (optional).
type RedisClient struct {
	prefix string
	Redis  *redis.Client
}

// Prefix returns the prefix string that was defined for the REDIS client.
func (r *RedisClient) Prefix() string {
	return r.prefix
}

// NewRedis creates a new RedisClient.
func NewRedis(redisHost string, redisPort string, prefix string) (*RedisClient, error) {
	redisURL := fmt.Sprintf("%s:%s", redisHost, redisPort)
	opts := redis.Options{
		Addr: redisURL,
	}

	client := redis.NewClient(&opts)
	_, err := client.Ping().Result()
	if err != nil {
		return nil, errors.New("could not ping REDIS")
	}

	redisClient := &RedisClient{
		prefix: prefix,
		Redis:  client,
	}
	return redisClient, nil
}

// Set saves a key value pair to REDIS.
// If the client was set up with a prefix it will be added in front of the key.
// Redis `SET key value [expiration]` command.
// Use expiration for `SETEX`-like behavior. Zero expiration means the key has no expiration time.
func (r *RedisClient) Set(key string, value string, expiration time.Duration) error {
	return r.Redis.Set(r.prefixedKey(key), value, expiration).Err()
}

// Get retrieves a value from REDIS.
// If the client was set up with a prefix it will be added in front of the key.
// If the value was not found ErrNotFound will be returned.
func (r *RedisClient) Get(key string) (string, error) {
	result, err := r.Redis.Get(r.prefixedKey(key)).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	return result, err
}

// SetBool saves a boolean value to REDIS.
// If the client was set up with a prefix it will be added in front of the key.
// Zero expiration means the key has no expiration time.
func (r *RedisClient) SetBool(key string, value bool, expiration time.Duration) error {
	return r.Set(key, strconv.FormatBool(value), expiration)
}

// GetBool retrieves a boolean value from REDIS.
// If the client was set up with a prefix it will be added in front of the key.
func (r *RedisClient) GetBool(key string) (bool, error) {
	result, err := r.Get(key)
	if err == redis.Nil {
		return false, ErrNotFound
	}
	return strconv.ParseBool(result)
}

// SetJSON saves JSON data as string to REDIS.
// If the client was set up with a prefix it will be added in front of the key.
// Zero expiration means the key has no expiration time.
func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(key, string(bytes[:]), expiration)
}

// GetJSON retrieves stringified JSON data from REDIS and parses it into the provided struct.
// If the client was set up with a prefix it will be added in front of the key.
func (r *RedisClient) GetJSON(key string, result interface{}) error {
	resultStr, err := r.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(resultStr), &result)
}

// Del deletes a key value pair from REDIS.
// If the client was set up with a prefix it will be added in front of the key.
func (r *RedisClient) Del(key string) error {
	return r.Redis.Del(r.prefixedKey(key)).Err()
}

// Close closes the connection to the REDIS server.
func (r *RedisClient) Close() error {
	return r.Redis.Close()
}

// prefixedKey adds the prefix in front of the key separated with ":".
// If no prefix was provided for the client than the key is returned as is.
func (r *RedisClient) prefixedKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + ":" + key
}
