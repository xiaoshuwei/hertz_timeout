package timeout

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

const (
	defaultTimeout = 5 * time.Second
)

// New wraps a handler and aborts the process of the handler if the timeout is reached
func New(opts ...Option) app.HandlerFunc {
	t := &Timeout{
		timeout:  defaultTimeout,
		handler:  nil,
		response: defaultResponse,
	}

	// Loop through each option
	for _, opt := range opts {
		if opt == nil {
			panic("timeout Option not be nil")
		}

		// Call the option giving the instantiated
		opt(t)
	}

	if t.timeout <= 0 {
		return t.handler
	}

	return func(ctx context.Context, c *app.RequestContext) {
		finish := make(chan struct{}, 1)
		panicChan := make(chan interface{}, 1)

		copyRc := c.Copy()
		copyRc.GetResponse().HijackWriter(NewTimeoutExtWriter())

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			t.handler(ctx, copyRc)
			finish <- struct{}{}
		}()

		select {
		case p := <-panicChan:
			panic(p)

		case <-finish:
			c.Next(ctx)
			copyRc.VisitAllHeaders(func(key, value []byte) {
				c.Header(string(key), string(value))
			})
			if _, err := c.Write(copyRc.GetResponse().Body()); err != nil {
				panic(err)
			}

		case <-time.After(t.timeout):
			c.Abort()
			t.response(ctx, c)
		}
	}
}
