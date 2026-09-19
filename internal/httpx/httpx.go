// Package httpx holds the HTTP client shared by every outbound request.
//
// http.DefaultClient has no timeout, so a server that accepts the connection
// and never answers hangs the process forever. This tool runs from cron and
// holds the bbolt lock for the length of a run, so a hung request does not
// just lose one run: it blocks every run that follows until it is killed.
package httpx

import (
	"net"
	"net/http"
	"time"
)

// RequestTimeout bounds a whole exchange: connect, TLS handshake, request,
// response headers and body read. It has to fit the largest thing we fetch,
// an og:image of up to maxDownloadSize, while still being short enough that a
// stalled host cannot outlive the cron interval.
const RequestTimeout = 30 * time.Second

const (
	dialTimeout           = 10 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	responseHeaderTimeout = 15 * time.Second
)

// Client is the shared client. Reuse it rather than building a new one per
// call, so connections are pooled and no call site can forget the timeout.
var Client = newClient(RequestTimeout)

func newClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   dialTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
		},
	}
}

// Get issues a GET through the shared client.
func Get(url string) (*http.Response, error) {
	return Client.Get(url)
}
