package telegram

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sublink/internal/testutil"
	"sublink/utils"
)

func TestStopBotCancelsRequestsAndRetryWait(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	connections := testutil.TrackHTTPConnections(server)
	server.Start()
	t.Cleanup(server.Close)
	ctx, cancel := context.WithCancel(context.Background())
	bot := &TelegramBot{ctx: ctx, cancel: cancel, client: utils.NewHTTPClientWithAdapter(nil, time.Minute, nil), stopChan: make(chan struct{})}
	t.Cleanup(bot.Stop)
	requestDone := make(chan error, 1)
	go func() {
		req, err := http.NewRequestWithContext(bot.ctx, http.MethodGet, server.URL, nil)
		if err != nil {
			requestDone <- err
			return
		}
		resp, err := bot.client.Do(req)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		requestDone <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not start")
	}
	retryDone := make(chan bool, 1)
	go func() { retryDone <- bot.waitForPollingRetry(time.Minute) }()
	bot.Stop()
	bot.Stop()
	select {
	case err := <-requestDone:
		if err == nil {
			t.Fatal("request was not canceled")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("request retained after bot stopped")
	}
	select {
	case retried := <-retryDone:
		if retried {
			t.Fatal("stopped bot retried")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("retry wait retained after bot stopped")
	}
	connections.WaitForClosed(t)
}
