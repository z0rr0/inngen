// Package web provides web services.
package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/z0rr0/inngen/web/handlers"
	"github.com/z0rr0/inngen/web/limiter"
	"github.com/z0rr0/inngen/web/middlewares"
)

const (
	// maxHeaderBytes is the maximum size of HTTP headers.
	maxHeaderBytes = 1 << 20 // 1 MB

	rateLimit        = 2.0
	rateLimitBurst   = 10.0
	rateLimitTimeout = 10 * time.Second

	readTimeout  = 2 * time.Second
	writeTimeout = 3 * time.Second
)

// Server represents an HTTP server.
type Server struct {
	srv         *http.Server
	waitLimiter chan struct{}
	stopped     chan struct{}
}

// NewAPIServer creates a new HTTP server with the given configuration.
func NewAPIServer(ctxStop context.Context, address string) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.GetHandler)
	mux.HandleFunc("POST /", handlers.PostHandler)

	rateLimiter := limiter.NewRateLimiter[string](
		rateLimit,
		rateLimitBurst,
		time.Second,
		nil,
	)

	// reversed order of middlewares wrapping, 1st will be called last
	wrappedMux := middlewares.BasicAuthMiddleware(mux, nil)
	wrappedMux = middlewares.RateLimiterMiddleware(wrappedMux, rateLimiter)
	wrappedMux = middlewares.LoggingMiddleware(wrappedMux)
	wrappedMux = middlewares.ErrorHandlingMiddleware(wrappedMux)

	return &Server{
		srv: &http.Server{
			Addr:           address,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			MaxHeaderBytes: maxHeaderBytes,
			Handler:        wrappedMux,
		},
		waitLimiter: rateLimiter.Cleanup(ctxStop, rateLimitTimeout),
		stopped:     make(chan struct{}),
	}
}

// Run starts the server and listens for incoming requests.
func (s *Server) Run() {
	go func() {
		const delay = 15 * time.Second
		err := s.srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			ticker := time.NewTicker(delay)
			defer ticker.Stop()
			// infinite loop to print error every 15 seconds until stopped
			for {
				select {
				case <-s.stopped:
					slog.Info("stopping error print")
					return
				case <-ticker.C:
					slog.Error("failed to start HTTP server", "error", err)
				}
			}
		}
	}()
}

// ShutdownHTTP gracefully shuts down the server with a timeout.
func (s *Server) ShutdownHTTP(ctx context.Context) error {
	close(s.stopped)
	return s.srv.Shutdown(ctx)
}

// GracefulShutdown gracefully shuts down other services.
func (s *Server) GracefulShutdown() {
	<-s.waitLimiter
}
