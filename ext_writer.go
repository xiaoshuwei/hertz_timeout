package timeout

import (
	"bytes"
	"runtime"
	"sync"

	"github.com/cloudwego/hertz/pkg/network"
)

type timeoutWriter struct {
	tmp []byte
	Buf *bytes.Buffer
}

var timeoutReaderPool sync.Pool

func init() {
	timeoutReaderPool = sync.Pool{
		New: func() interface{} {
			return &timeoutWriter{
				Buf: &bytes.Buffer{},
			}
		},
	}
}

// Write
func (c *timeoutWriter) Write(p []byte) (n int, err error) {
	c.tmp = p
	return len(p), nil
}

func (c *timeoutWriter) Flush() error {
	_, err := c.Buf.Write(c.tmp)
	return err
}

// Finalize
func (c *timeoutWriter) Finalize() error {
	return nil
}

func (c *timeoutWriter) release() {
	c.tmp = nil
	c.Buf.Reset()
	timeoutReaderPool.Put(c)
}

func NewTimeoutExtWriter() network.ExtWriter {
	extWriter := timeoutReaderPool.Get().(*timeoutWriter)
	runtime.SetFinalizer(extWriter, (*timeoutWriter).release)
	return extWriter
}
