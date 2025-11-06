package middlewares

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
// and tracks the number of bytes writtenBytes to the response.
type responseWriter struct {
	http.ResponseWriter

	wroteHeader  atomic.Bool  // tracks if WriteHeader has been called
	writtenBytes atomic.Int64 // tracks the number of bytes written
	status       int          // stores the HTTP status code
}

// wrapResponseWriter creates a new responseWriter that wraps the provided http.ResponseWriter.
// It initializes the status to http.StatusOK by default.
func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// WriteHeader sets the HTTP status code for the response.
// It ensures that the status code is only set once by using atomic operations.
// Subsequent calls to WriteHeader are ignored.
func (rw *responseWriter) WriteHeader(code int) {
	if swapped := rw.wroteHeader.CompareAndSwap(false, true); !swapped {
		return // already written
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the underlying ResponseWriter and tracks the number of bytes writtenBytes.
// It calls WriteHeader with http.StatusOK if no status code has been set yet.
// The number of bytes writtenBytes is tracked using atomic operations for thread safety.
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.WriteHeader(http.StatusOK) // will be ignored if any other status code was set before
	n, err := rw.ResponseWriter.Write(b)

	if err != nil {
		return n, err
	}

	rw.writtenBytes.Add(int64(n))
	return n, nil
}

// BytesWritten returns the total number of bytes writtenBytes to the response.
func (rw *responseWriter) BytesWritten() int64 {
	return rw.writtenBytes.Load()
}

// Status returns the HTTP status code of the response.
func (rw *responseWriter) Status() int {
	return rw.status
}

// Flush implements the http.Flusher interface to allow flushing buffered data to the client.
// If the underlying ResponseWriter supports Flush, it will be called.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements the http.Hijacker interface to allow taking over the connection.
// If the underlying ResponseWriter supports Hijack, it will be called.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("underlying ResponseWriter does not implement http.Hijacker")
}

// Push implements the http.Pusher interface to support HTTP/2 server push.
// If the underlying ResponseWriter supports Push, it will be called.
func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return errors.New("underlying ResponseWriter does not implement http.Pusher")
}
