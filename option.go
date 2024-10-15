package timeout

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"time"
)

// Option for timeout
type Option func(*Timeout)

// WithTimeout set timeout
func WithTimeout(timeout time.Duration) Option {
	return func(t *Timeout) {
		t.timeout = timeout
	}
}

// WithHandler add gin handler
func WithHandler(h app.HandlerFunc) Option {
	return func(t *Timeout) {
		t.handler = h
	}
}

// WithResponse add gin handler
func WithResponse(h app.HandlerFunc) Option {
	return func(t *Timeout) {
		t.response = h
	}
}

func defaultResponse(ctx context.Context, c *app.RequestContext) {
	c.String(http.StatusRequestTimeout, http.StatusText(http.StatusRequestTimeout))
}

// Timeout struct
type Timeout struct {
	timeout  time.Duration
	handler  app.HandlerFunc
	response app.HandlerFunc
}
