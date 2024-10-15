package timeout

import (
	"runtime"
	"sync"

	"github.com/cloudwego/hertz/pkg/network"
)

type timeoutBodyWriter struct {
	w network.Writer
}

var timeoutReaderPool sync.Pool

func init() {
	timeoutReaderPool = sync.Pool{
		New: func() interface{} {
			return &timeoutBodyWriter{}
		},
	}
}

// Write
func (c *timeoutBodyWriter) Write(p []byte) (n int, err error) {
	return c.w.WriteBinary(p)
}

func (c *timeoutBodyWriter) Flush() error {
	return c.w.Flush()
}

// Finalize
func (c *timeoutBodyWriter) Finalize() error {
	return nil
}

func (c *timeoutBodyWriter) release() {
	c.w = nil
	timeoutReaderPool.Put(c)
}

func NewTimeoutBodyWriter(w network.Writer) network.ExtWriter {
	extWriter := timeoutReaderPool.Get().(*timeoutBodyWriter)
	extWriter.w = w
	runtime.SetFinalizer(extWriter, (*timeoutBodyWriter).release)
	return extWriter
}
