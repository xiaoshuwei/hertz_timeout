package timeout

import (
	"context"
	"github.com/cloudwego/hertz/pkg/network"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

var bufPool *BufferPool

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

	bufPool = &BufferPool{}

	return func(ctx context.Context, c *app.RequestContext) {
		finish := make(chan struct{}, 1)
		panicChan := make(chan interface{}, 1)

		network.NewWriter(NewWriter())
		c.GetResponse().HijackWriter(NewTimeoutBodyWriter())
		copyRc := c.Copy()

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
			tw.FreeBuffer()
			c.Writer = w
			panic(p)

		case <-finish:
			c.Next()
			tw.mu.Lock()
			defer tw.mu.Unlock()
			dst := tw.ResponseWriter.Header()
			for k, vv := range tw.Header() {
				dst[k] = vv
			}

			if _, err := tw.ResponseWriter.Write(buffer.Bytes()); err != nil {
				panic(err)
			}
			tw.FreeBuffer()
			bufPool.Put(buffer)

		case <-time.After(t.timeout):
			c.Abort()
			tw.mu.Lock()
			defer tw.mu.Unlock()
			tw.timeout = true
			tw.FreeBuffer()
			bufPool.Put(buffer)

			c.Writer = w
			t.response(c)
			c.Writer = tw
		}
	}
}
