package observance

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPrometheusMetrics(t *testing.T) {
	t.Run("full setup", func(t *testing.T) {
		var callCounter uint64
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(&callCounter, 1)
			assert.Contains(t, r.URL.Path, "/metrics/job/test-app/instance/")
			assertBodyContains(t, r, "test_metric", "some_other_metric")
			w.WriteHeader(http.StatusAccepted)
		}))
		defer ts.Close()

		m := NewPrometheusMetrics(ts.URL, "test-app", 5*time.Millisecond, NewTestLogger())

		for i := 0; i < 12; i++ {
			m.Increment("test_metric")
			m.Increment("some_other_metric")
			m.Increment("test_metric")
			time.Sleep(2 * time.Millisecond)
		}
		result := atomic.LoadUint64(&callCounter)
		assert.True(t, result >= 4)
		assert.True(t, result <= 7)
	})
}

func TestMetricTypes(t *testing.T) {
	var m Measurer
	cases := []struct {
		name string
		fn   func(name string)
	}{
		{
			"Increment",
			func(name string) { m.Increment(name) },
		},
		{
			"SetGauge",
			func(name string) { m.SetGauge(name, 123.456) },
		},
		{
			"SetGaugeInt64",
			func(name string) { m.SetGaugeInt64(name, 123) },
		},
		{
			"DurationSince",
			func(name string) { m.DurationSince(name, time.Now()) },
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var callCounter uint64
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddUint64(&callCounter, 1)
				assertBodyContains(t, r, "test_metric")
				w.WriteHeader(http.StatusAccepted)
			}))
			defer ts.Close()
			m = NewPrometheusMetrics(ts.URL, "test-app", 70*time.Millisecond, NewTestLogger())

			test.fn("test_metric")
			time.Sleep(100 * time.Millisecond)
			assert.Equal(t, uint64(1), atomic.LoadUint64(&callCounter))
		})
	}
}

func TestMeasurer(t *testing.T) {
	assert.Implements(t, (*Measurer)(nil), &PrometheusMetrics{})
}

func assertBodyContains(t *testing.T, r *http.Request, expected ...string) {
	bytes, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	for _, value := range expected {
		assert.Contains(t, string(bytes), value)
	}
}
