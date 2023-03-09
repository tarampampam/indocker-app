package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Sender interface {
	// Send sends events to the collector.
	Send(ctx context.Context, userID string, events ...Event) error
}

type Event struct {
	Name       string            // what happened
	Properties map[string]string // additional info
	Timestamp  time.Time         // when it happened
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// MixPanelSender allows sending events to mixpanel.com.
type MixPanelSender struct {
	projectToken string // settings -> projects -> Project Token
	appVersion   string
	client       httpClient
}

var _ Sender = (*MixPanelSender)(nil) // ensure interface is implemented

// NewMixPanelSender creates a new MixPanelSender.
func NewMixPanelSender(projectToken, appVersion string, client ...httpClient) *MixPanelSender {
	if len(client) == 0 {
		client = []httpClient{ // default client
			&http.Client{
				Timeout: time.Second * 30, //nolint:gomnd
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
				},
			},
		}
	}

	return &MixPanelSender{
		projectToken: projectToken,
		appVersion:   appVersion,
		client:       client[0],
	}
}

// Send sends events to mixpanel.com. If event name is empty, it will be skipped.
func (mp *MixPanelSender) Send(ctx context.Context, userID string, events ...Event) error { //nolint:funlen
	if len(events) == 0 {
		return errors.New("empty events")
	} else if len(events) > 50 { //nolint:gomnd // https://bit.ly/3YA4C10
		return errors.New("too many events")
	}

	type mixEvent struct { // https://developer.mixpanel.com/reference/track-event
		Event      string            `json:"event"`
		Properties map[string]string `json:"properties"`
	}

	var (
		mixEvents = make([]mixEvent, 0, len(events))
		os        = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	)

	const (
		tokenKey      = "token"
		appVersionKey = "$app_version_string"
		userIDKey     = "distinct_id"
		osKey         = "$os"
		timeKey       = "time"
	)

	for i := range events {
		if events[i].Name == "" {
			continue
		}

		var props = make(map[string]string, len(events[i].Properties)+4) //nolint:gomnd

		for k, v := range events[i].Properties { // copy properties
			props[k] = v
		}

		// force common properties
		props[tokenKey] = mp.projectToken
		props[userIDKey] = userID
		props[osKey] = os

		if mp.appVersion != "" {
			props[appVersionKey] = mp.appVersion
		}

		if !events[i].Timestamp.IsZero() {
			props[timeKey] = fmt.Sprintf("%d", events[i].Timestamp.Unix())
		}

		mixEvents = append(mixEvents, mixEvent{
			Event:      events[i].Name,
			Properties: props,
		})
	}

	var buf = new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(mixEvents); err != nil {
		return err
	}

	const endpoint = "https://api.mixpanel.com/track?ip=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := mp.client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Trim(string(body), " \r\n\t") != "1" {
		return fmt.Errorf("invalid response: %s", string(body))
	}

	return nil
}
