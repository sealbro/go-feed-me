package graceful

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
)

type ShutdownCloser struct {
	services map[string]io.Closer
	m        sync.Mutex
}

func NewShutdownCloser() *ShutdownCloser {
	return &ShutdownCloser{
		services: make(map[string]io.Closer),
		m:        sync.Mutex{},
	}
}

func (c *ShutdownCloser) Register(closer io.Closer) {
	c.m.Lock()
	c.services[reflect.TypeOf(closer).String()] = closer
	c.m.Unlock()
}

func (c *ShutdownCloser) Close() error {
	builder := strings.Builder{}
	for _, closer := range c.services {
		err := closer.Close()
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v\n", err))
		}
	}

	s := builder.String()
	if len(s) > 0 {
		return fmt.Errorf(s)
	}

	return nil
}
