package collector_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/daemon/internal/collector"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestMixPanelSender_Send(t *testing.T) {
	t.Parallel()

	var now = time.Now()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "https://x-collect-v1.indocker.app/track?ip=1", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

		j, _ := io.ReadAll(req.Body)
		os := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

		assert.JSONEq(t, `[
	{
		"event":"foo",
		"properties":{
			"$app_version_string":"appVersion321",
			"$os":"`+os+`",
			"distinct_id":"userID",
			"foo1":"bar1",
			"token":"projectToken123"
		}
	},
	{
		"event":"bar",
		"properties":{
			"$app_version_string":"appVersion321",
			"$os":"`+os+`",
			"distinct_id":"userID",
			"foo2":"bar2",
			"token":"projectToken123",
			"time":"`+fmt.Sprintf("%d", now.Unix())+`"
		}
	}
]`, string(j))

		return &http.Response{
			Body:       io.NopCloser(bytes.NewReader([]byte("1"))),
			StatusCode: http.StatusOK,
		}, nil
	}

	var sender = collector.NewMixPanelSender("projectToken123", "appVersion321", httpMock)

	err := sender.Send(context.Background(), "userID", collector.Event{
		Name:       "foo",
		Properties: map[string]string{"foo1": "bar1"},
	}, collector.Event{
		Name:       "bar",
		Properties: map[string]string{"foo2": "bar2"},
		Timestamp:  now,
	})

	assert.NoError(t, err)
}
