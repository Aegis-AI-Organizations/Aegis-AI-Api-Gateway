package testutils

import (
	"net/http/httptest"
)

// CloseNotifierRecorder is a ResponseRecorder that implements the CloseNotifier interface.
type CloseNotifierRecorder struct {
	*httptest.ResponseRecorder
	Closed chan bool
}

// CloseNotify returns a channel that receives at most a single value when the client connection has gone away.
func (c *CloseNotifierRecorder) CloseNotify() <-chan bool {
	return c.Closed
}

// NewCloseNotifierRecorder returns an initialized CloseNotifierRecorder.
func NewCloseNotifierRecorder() *CloseNotifierRecorder {
	return &CloseNotifierRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		Closed:           make(chan bool, 1),
	}
}
