package cachemock

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// Cache is a mock implementation of the cache.Cache interface.
type Cache struct {
	mock.Mock
}

// Close is a mock implementation of cache.Cache#Close.
func (m *Cache) Close() error {
	args := m.Called()

	return args.Error(0)
}

// Del is a mock implementation of cache.Cache#Del.
func (m *Cache) Del(key string) error {
	args := m.Called(key)

	return args.Error(0)
}

// Get is a mock implementation of cache.Cache#Get.
func (m *Cache) Get(key string) (string, error) {
	args := m.Called(key)

	return args.String(0), args.Error(1)
}

// GetBool is a mock implementation of cache.Cache#GetBool.
func (m *Cache) GetBool(key string) (bool, error) {
	args := m.Called(key)

	return args.Bool(0), args.Error(1)
}

// GetJSON is a mock implementation of cache.Cache#GetJSON.
func (m *Cache) GetJSON(key string, result interface{}) error {
	args := m.Called(key, result)

	return args.Error(0)
}

// Prefix is a mock implementation of cache.Cache#Prefix.
func (m *Cache) Prefix() string {
	args := m.Called()

	return args.String(0)
}

// Set is a mock implementation of cache.Cache#Set.
func (m *Cache) Set(key string, value string, expiration time.Duration) error {
	args := m.Called(key, value, expiration)

	return args.Error(0)
}

// SetBool is a mock implementation of cache.Cache#SetBool.
func (m *Cache) SetBool(key string, value bool, expiration time.Duration) error {
	args := m.Called(key, value, expiration)

	return args.Error(0)
}

// SetJSON is a mock implementation of cache.Cache#SetJSON.
func (m *Cache) SetJSON(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)

	return args.Error(0)
}
