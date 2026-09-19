package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientHasTimeout(t *testing.T) {
	if Client.Timeout == 0 {
		t.Fatal("shared client has no timeout")
	}
}

// A server that accepts the connection and never writes must not hang the
// caller: the client has to give up on its own.
func TestStalledServerTimesOut(t *testing.T) {
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	// Close waits for the handler to return, so the handler must be released
	// first: defers run last-registered-first.
	defer server.Close()
	defer close(release)

	client := newClient(100 * time.Millisecond)

	done := make(chan error, 1)
	go func() {
		resp, err := client.Get(server.URL)
		if err == nil {
			resp.Body.Close()
		}
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a timeout error, got none")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("request did not time out")
	}
}
