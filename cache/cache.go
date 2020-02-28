package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

// ErrNotFound is returned when the key was not found in the cache.
var ErrNotFound = errors.New("key not found in cache")

// Cache defines basic cache operations including methods for setting and getting JSON objects.
type Cache interface {
	Prefix() string
	Set(key string, value string, expiration time.Duration) error
	Get(key string) (string, error)
	SetBool(key string, value bool, expiration time.Duration) error
	GetBool(key string) (bool, error)
	SetInt(key string, value int64, expiration time.Duration) error
	GetInt(key string) (int64, error)
	Incr(key string) (int64, error)
	SetJSON(key string, value interface{}, expiration time.Duration) error
	GetJSON(key string, result interface{}) error
	Del(key string) error
	Close() error
	TTL(key string) (time.Duration, error)
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
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(result)
}

// SetInt saves an integer value to REDIS.
// If the client was set up with a prefix it will be added in front of the key.
// Zero expiration means the key has no expiration time.
func (r *RedisClient) SetInt(key string, value int64, expiration time.Duration) error {
	return r.Set(key, strconv.FormatInt(value, 10), expiration)
}

// GetInt retrieves an integer value from REDIS.
// If the client was set up with a prefix it will be added in front of the key.
func (r *RedisClient) GetInt(key string) (int64, error) {
	result, err := r.Get(key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(result, 10, 64)
}

// Incr increments a value in REDIS.
// If the client was set up with a prefix it will be added in front of the key.
// It returns the new (incremented) value.
func (r *RedisClient) Incr(key string) (int64, error) {
	return r.Redis.Incr(r.prefixedKey(key)).Result()
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

// TTL returns remaining time to live of the given key found in REDIS.
// If the key doesn't exist, it returns ErrNotFound.
func (r *RedisClient) TTL(key string) (time.Duration, error) {
	result, err := r.Redis.TTL(key).Result()
	if err != nil {
		return 0, err
	}

	if result == -2 {
		return 0, ErrNotFound
	}

	return result, nil
}

// prefixedKey adds the prefix in front of the key separated with ":".
// If no prefix was provided for the client than the key is returned as is.
func (r *RedisClient) prefixedKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + ":" + key
}
