package middlewares

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockFlusher implements http.Flusher for testing
type mockFlusher struct {
	http.ResponseWriter
	flushed bool
}

func (m *mockFlusher) Flush() {
	m.flushed = true
}

// mockHijacker implements http.Hijacker for testing
type mockHijacker struct {
	http.ResponseWriter
	hijacked bool
}

func (m *mockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	m.hijacked = true
	return &mockConn{}, bufio.NewReadWriter(bufio.NewReader(nil), bufio.NewWriter(nil)), nil
}

// mockPusher implements http.Pusher for testing
type mockPusher struct {
	http.ResponseWriter
	pushed     bool
	pushedPath string
	pushedOpts *http.PushOptions
}

func (m *mockPusher) Push(target string, opts *http.PushOptions) error {
	m.pushed = true
	m.pushedPath = target
	m.pushedOpts = opts
	return nil
}

// mockFailingPusher implements http.Pusher that fails
type mockFailingPusher struct {
	http.ResponseWriter
}

func (m *mockFailingPusher) Push(_ string, _ *http.PushOptions) error {
	return errors.New("push failed")
}

// mockFailingHijacker implements http.Hijacker that fails
type mockFailingHijacker struct {
	http.ResponseWriter
}

func (m *mockFailingHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack failed")
}

// mockConn implements net.Conn for testing
type mockConn struct{}

func (m *mockConn) Read(_ []byte) (n int, err error)   { return 0, io.EOF }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *mockConn) SetDeadline(_ time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(_ time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(_ time.Time) error { return nil }

func TestResponseWriter(t *testing.T) {
	tests := []struct {
		name         string
		writeStatus  int
		writeBody    string
		expectStatus int
		expectBody   string
	}{
		{
			name:         "explicit status",
			writeStatus:  http.StatusCreated,
			writeBody:    "test body",
			expectStatus: http.StatusCreated,
			expectBody:   "test body",
		},
		{
			name:         "default status",
			writeBody:    "test body",
			expectStatus: http.StatusOK,
			expectBody:   "test body",
		},
		{
			name:         "multiple writes",
			writeStatus:  http.StatusOK,
			writeBody:    "test body",
			expectStatus: http.StatusOK,
			expectBody:   "test body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			wrapped := wrapResponseWriter(rec)

			if tc.writeStatus != 0 {
				wrapped.WriteHeader(tc.writeStatus)
				// 2nd call should be ignored
				wrapped.WriteHeader(http.StatusTeapot)
			}

			if tc.writeBody != "" {
				_, _ = wrapped.Write([]byte(tc.writeBody))
			}

			result := rec.Result()
			defer func() {
				if err := result.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			if result.StatusCode != tc.expectStatus {
				t.Errorf("got status %d, want %d", result.StatusCode, tc.expectStatus)
			}

			body := rec.Body.String()
			if body != tc.expectBody {
				t.Errorf("got body %q, want %q", body, tc.expectBody)
			}

			// count writtenBytes bytes
			if n, m := int64(len(tc.expectBody)), wrapped.BytesWritten(); m != n {
				t.Errorf("got writtenBytes bytes %d, want %d", m, n)
			}
		})
	}
}

// TestResponseWriterStatus tests the Status method of responseWriter
func TestResponseWriterStatus(t *testing.T) {
	tests := []struct {
		name           string
		writeStatus    int
		expectedStatus int
	}{
		{
			name:           "default status",
			writeStatus:    0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom status",
			writeStatus:    http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "error status",
			writeStatus:    http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			wrapped := wrapResponseWriter(rec)

			if tc.writeStatus != 0 {
				wrapped.WriteHeader(tc.writeStatus)
			}

			if wrapped.Status() != tc.expectedStatus {
				t.Errorf("Status() = %d, want %d", wrapped.Status(), tc.expectedStatus)
			}
		})
	}
}

// TestResponseWriterFlush tests the Flush method of responseWriter
func TestResponseWriterFlush(t *testing.T) {
	tests := []struct {
		name         string
		supportFlush bool
	}{
		{
			name:         "with flusher",
			supportFlush: true,
		},
		{
			name:         "without flusher",
			supportFlush: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var rw http.ResponseWriter
			var mockFlush *mockFlusher

			if tc.supportFlush {
				mockFlush = &mockFlusher{ResponseWriter: httptest.NewRecorder()}
				rw = mockFlush
			} else {
				rw = httptest.NewRecorder() // doesn't implement Flush
			}

			wrapped := wrapResponseWriter(rw)
			wrapped.Flush()

			if tc.supportFlush {
				if !mockFlush.flushed {
					t.Error("Flush() should call the underlying Flush method")
				}
			} else {
				// No assertion needed, just ensuring no panic occurs
				// Call wrapped.Flush() again to ensure no panic
				wrapped.Flush()
			}
		})
	}
}

// TestResponseWriterHijack tests the Hijack method of responseWriter
func TestResponseWriterHijack(t *testing.T) {
	tests := []struct {
		name          string
		hijacker      http.ResponseWriter
		expectSuccess bool
		expectError   string
	}{
		{
			name:          "successful hijack",
			hijacker:      &mockHijacker{ResponseWriter: httptest.NewRecorder()},
			expectSuccess: true,
		},
		{
			name:          "failed hijack",
			hijacker:      &mockFailingHijacker{ResponseWriter: httptest.NewRecorder()},
			expectSuccess: false,
			expectError:   "hijack failed",
		},
		{
			name:          "not a hijacker",
			hijacker:      httptest.NewRecorder(),
			expectSuccess: false,
			expectError:   "underlying ResponseWriter does not implement http.Hijacker",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := wrapResponseWriter(tc.hijacker)
			conn, rw, err := wrapped.Hijack()

			if tc.expectSuccess {
				if err != nil {
					t.Errorf("Hijack() error = %v, want nil", err)
				}
				if conn == nil {
					t.Error("Hijack() conn = nil, want non-nil")
				}
				if rw == nil {
					t.Error("Hijack() rw = nil, want non-nil")
				}
				if mh, ok := tc.hijacker.(*mockHijacker); ok {
					if !mh.hijacked {
						t.Error("Hijack() should call the underlying Hijack method")
					}
				}
			} else {
				if err == nil {
					t.Errorf("Hijack() error = nil, want error containing %q", tc.expectError)
				} else if !strings.Contains(err.Error(), tc.expectError) {
					t.Errorf("Hijack() error = %q, want error containing %q", err.Error(), tc.expectError)
				}
				if conn != nil {
					t.Errorf("Hijack() conn = %v, want nil", conn)
				}
				if rw != nil {
					t.Errorf("Hijack() rw = %v, want nil", rw)
				}
			}
		})
	}
}

// TestResponseWriterPush tests the Push method of responseWriter
func TestResponseWriterPush(t *testing.T) {
	tests := []struct {
		name          string
		pusher        http.ResponseWriter
		target        string
		opts          *http.PushOptions
		expectSuccess bool
		expectError   string
	}{
		{
			name:          "successful push",
			pusher:        &mockPusher{ResponseWriter: httptest.NewRecorder()},
			target:        "/style.css",
			opts:          &http.PushOptions{Method: "GET"},
			expectSuccess: true,
		},
		{
			name:          "failed push",
			pusher:        &mockFailingPusher{ResponseWriter: httptest.NewRecorder()},
			target:        "/script.js",
			expectSuccess: false,
			expectError:   "push failed",
		},
		{
			name:          "not a pusher",
			pusher:        httptest.NewRecorder(),
			target:        "/image.png",
			expectSuccess: false,
			expectError:   "underlying ResponseWriter does not implement http.Pusher",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := wrapResponseWriter(tc.pusher)
			err := wrapped.Push(tc.target, tc.opts)

			if tc.expectSuccess {
				if err != nil {
					t.Errorf("Push() error = %v, want nil", err)
				}
				if mp, ok := tc.pusher.(*mockPusher); ok {
					if !mp.pushed {
						t.Error("Push() should call the underlying Push method")
					}
					if mp.pushedPath != tc.target {
						t.Errorf("Push() pushedPath = %q, want %q", mp.pushedPath, tc.target)
					}
					if mp.pushedOpts != tc.opts {
						t.Errorf("Push() pushedOpts = %v, want %v", mp.pushedOpts, tc.opts)
					}
				}
			} else {
				if err == nil {
					t.Errorf("Push() error = nil, want error containing %q", tc.expectError)
				} else if !strings.Contains(err.Error(), tc.expectError) {
					t.Errorf("Push() error = %q, want error containing %q", err.Error(), tc.expectError)
				}
			}
		})
	}
}
