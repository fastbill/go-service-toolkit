package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name string `json:"fullName"`
	Age  int    `json:"age"`
}

func TestNewRedis(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err, "error in test setup")
	defer redisServer.Close()

	client, err := NewRedis(redisServer.Host(), redisServer.Port(), "testPrefix")
	assert.NoError(t, err)
	assert.Implements(t, (*Cache)(nil), client)
}

func TestPrefix(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err, "error in test setup")
	defer redisServer.Close()

	client, err := NewRedis(redisServer.Host(), redisServer.Port(), "testPrefix")
	assert.NoError(t, err)
	assert.Equal(t, "testPrefix", client.Prefix())
}

func TestSet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := client.Set("someKey", "someValue", 10*time.Minute)
			assert.NoError(t, err)
			redis.CheckGet(t, "testPrefix:someKey", "someValue")

			// Check expiry date was set
			redis.FastForward(11 * time.Minute)
			assert.False(t, redis.Exists("testPrefix:someKey"))
		})
	})

	t.Run("success, no prefix", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			client.prefix = ""
			err := client.Set("someKey", "someValue", 10*time.Minute)
			assert.NoError(t, err)
			redis.CheckGet(t, "someKey", "someValue")
		})
	})

	t.Run("failure", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			redis.Close()
			err := client.Set("someKey", "someValue", 10*time.Minute)
			assert.Error(t, err)
		})
	})
}

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "someValue")
			assert.NoError(t, err)
			result, err := client.Get("someKey")
			assert.NoError(t, err)
			assert.Equal(t, "someValue", result)
		})
	})

	t.Run("key not found", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			_, err := client.Get("someKey")
			assert.Error(t, ErrNotFound, err)
		})
	})

	t.Run("failure", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "someValue")
			assert.NoError(t, err)
			redis.Close()
			_, err = client.Get("someKey")
			assert.Error(t, err)
		})
	})
}

func TestSetBool(t *testing.T) {
	withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
		err := client.SetBool("someKey", true, 10*time.Minute)
		assert.NoError(t, err, "error in test setup")
		redis.CheckGet(t, "testPrefix:someKey", "true")
	})
}

func TestGetBool(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "true")
			assert.NoError(t, err)
			result, err := client.GetBool("someKey")
			assert.NoError(t, err)
			assert.True(t, result)
		})
	})

	t.Run("parsing error", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "wrong")
			assert.NoError(t, err)
			_, err = client.GetBool("someKey")
			assert.Error(t, err)
		})
	})

	t.Run("value not found", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			_, err := client.GetBool("someKey")
			assert.Equal(t, ErrNotFound, err)
		})
	})
}

func TestSetInt(t *testing.T) {
	withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
		err := client.SetInt("someKey", 100, 10*time.Minute)
		require.NoError(t, err, "error in test setup")
		redis.CheckGet(t, "testPrefix:someKey", "100")
	})
}

func TestGetInt(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "100")
			require.NoError(t, err)
			result, err := client.GetInt("someKey")
			require.NoError(t, err)
			assert.Equal(t, int64(100), result)
		})
	})

	t.Run("parsing error", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "invalid int")
			require.NoError(t, err)
			_, err = client.GetInt("someKey")
			require.Error(t, err)
		})
	})
}

func TestIncr(t *testing.T) {
	withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
		value, err := client.Incr("someKey")
		require.NoError(t, err, "error in test setup")
		assert.Equal(t, int64(1), value)

		value, err = client.Incr("someKey")
		require.NoError(t, err, "error in test setup")
		assert.Equal(t, int64(2), value)

		redis.CheckGet(t, "testPrefix:someKey", "2")
	})
}

func TestSetJSON(t *testing.T) {
	value := testStruct{
		Name: "Kathryn Janeway",
		Age:  42,
	}

	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := client.SetJSON("someKey", value, 10*time.Minute)
			assert.NoError(t, err)
			result, err := redis.Get("testPrefix:someKey")
			assert.NoError(t, err)
			assert.JSONEq(t, `{"fullName":"Kathryn Janeway","age":42}`, result)
		})
	})

	t.Run("invalid value", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := client.SetJSON("someKey", make(chan int), 10*time.Minute)
			assert.Error(t, err)
		})
	})

	t.Run("failure", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			redis.Close()
			err := client.SetJSON("someKey", value, 10*time.Minute)
			assert.Error(t, err)
		})
	})
}

func TestGetJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", `{"fullName":"Kathryn Janeway","age":42}`)
			assert.NoError(t, err)

			result := &testStruct{}
			err = client.GetJSON("someKey", result)
			assert.NoError(t, err)
			assert.Equal(t, "Kathryn Janeway", result.Name)
			assert.Equal(t, 42, result.Age)
		})
	})

	t.Run("invalid value", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", `{"fullName"`)
			assert.NoError(t, err)

			result := &testStruct{}
			err = client.GetJSON("someKey", result)
			assert.Error(t, err)
		})
	})

	t.Run("value not found", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			result := &testStruct{}
			err := client.GetJSON("foo", result)
			assert.Equal(t, ErrNotFound, err)
		})
	})
}

func TestDel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "someValue")
			assert.NoError(t, err)
			assert.True(t, redis.Exists("testPrefix:someKey"))

			err = client.Del("someKey")
			assert.NoError(t, err)
			assert.False(t, redis.Exists("testPrefix:someKey"))
		})
	})
}

func TestTTL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			err := redis.Set("testPrefix:someKey", "100")
			require.NoError(t, err)
			redis.SetTTL("testPrefix:someKey", 1*time.Second)

			result, err := client.TTL("someKey")

			require.NoError(t, err)
			assert.Equal(t, 1*time.Second, result)
		})
	})

	t.Run("not found", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			_, err := client.TTL("someKey")

			require.Error(t, err)
			assert.Equal(t, ErrNotFound, err)
		})
	})

	t.Run("failure", func(t *testing.T) {
		withRedis(t, func(redis *miniredis.Miniredis, client *RedisClient) {
			redis.Close()
			_, err := client.TTL("someKey")

			require.Error(t, err)
			assert.NotEqual(t, ErrNotFound, err)
		})
	})
}

func withRedis(t *testing.T, fn func(redis *miniredis.Miniredis, client *RedisClient)) {
	redis, err := miniredis.Run()
	assert.NoError(t, err, "error in test setup")
	defer redis.Close()

	client, err := NewRedis(redis.Host(), redis.Port(), "testPrefix")
	assert.NoError(t, err)

	fn(redis, client)
}
