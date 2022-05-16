package observance

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyWithRequest(t *testing.T) {
	config := Config{
		AppName:  "testApp",
		LogLevel: "debug",
		LoggedHeaders: map[string]string{
			"Fastbill-Outer-RequestId": "requestId",
			"Fastbill-AccountId":       "accountId",
		},
	}
	obs, err := NewObs(config)
	assert.NoError(t, err)
	capture := bytes.Buffer{}
	obs.Logger.SetOutput(&capture)

	t.Run("adds request properties to all log messages", func(t *testing.T) {
		r := httptest.NewRequest("POST", "http://example.com/test?foo=bar", nil)
		r.Header.Set("Fastbill-Outer-RequestId", "testRequestId")
		r.Header.Set("Fastbill-AccountId", "testAccountId")

		reqObs := obs.CopyWithRequest(r)
		reqObs.Logger.Debug("testLogMessage2öüä")
		got := capture.String()

		assert.Contains(t, got, `"msg":"testLogMessage2öüä"`)
		assert.Contains(t, got, `"requestId":"testRequestId"`)
		assert.Contains(t, got, `"accountId":"testAccountId"`)
		assert.Contains(t, got, `"method":"POST"`)
		assert.Contains(t, got, `"url":"http://example.com/test?foo=bar"`)

		capture = bytes.Buffer{}
		reqObs.Logger.Error("someOtherMsg")

		got = capture.String()
		assert.Contains(t, got, `"msg":"someOtherMsg"`)
		assert.Contains(t, got, `"requestId":"testRequestId"`)
		assert.Contains(t, got, `"accountId":"testAccountId"`)
		assert.Contains(t, got, `"method":"POST"`)
		assert.Contains(t, got, `"url":"http://example.com/test?foo=bar"`)
	})

	t.Run("old properties do not persist for new requests", func(t *testing.T) {
		capture = bytes.Buffer{}
		r := httptest.NewRequest("GET", "http://example.com", nil)

		reqObs := obs.CopyWithRequest(r)
		reqObs.Logger.Error("some message")
		got := capture.String()
		assert.Contains(t, got, `"msg":"some message"`)
		assert.NotContains(t, got, "accountId")
	})

	t.Run("main observance is not affected", func(t *testing.T) {
		capture = bytes.Buffer{}
		obs.Logger.Error("some message")
		got := capture.String()
		assert.NotContains(t, got, "url")
		assert.Contains(t, got, `"msg":"some message"`)
	})

}
