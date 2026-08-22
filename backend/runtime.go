package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"example.com/bridge-health-monitor-service/ops"
)

var requestSequence uint64

func serveAddress(address string, handler http.Handler) error {
	return serveHTTP(newEnterpriseServer(address, handler))
}

func serveHTTP(server *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-signals:
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownContext)
	}
}

func newEnterpriseServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           ops.EnterpriseMiddleware(requestIDMiddleware(recoveryMiddleware(deadlineMiddleware(handler)))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

// requestDeadlineHeader lets a caller cap how long the server may spend on a
// single request, expressed in milliseconds.
const requestDeadlineHeader = "X-Request-Deadline-Ms"

// defaultRequestTimeout bounds a request when the caller omits the header.
const defaultRequestTimeout = 5 * time.Second

func deadlineMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeout := requestTimeout(r)
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		// If the request is already past its deadline (client canceled, or the
		// header requested zero/negative time) fail fast instead of running the
		// handler against a dead context.
		if err := ctx.Err(); err != nil {
			writeDeadlineExceeded(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requestTimeout derives the per-request budget from the deadline header. A
// missing or unparsable header falls back to the default. A header of zero or
// a negative value means "already expired" and collapses to a zero duration.
func requestTimeout(r *http.Request) time.Duration {
	raw := strings.TrimSpace(r.Header.Get(requestDeadlineHeader))
	if raw == "" {
		return defaultRequestTimeout
	}
	ms, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return defaultRequestTimeout
	}
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

// writeDeadlineExceeded reports a 504 for a request whose deadline elapsed
// before its handler could complete.
func writeDeadlineExceeded(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Retry-After", "1")
	w.WriteHeader(http.StatusGatewayTimeout)
	_, _ = w.Write([]byte(http.StatusText(http.StatusGatewayTimeout)))
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", atomic.AddUint64(&requestSequence, 1))
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered request_id=%s method=%s path=%s panic=%v", w.Header().Get("X-Request-ID"), r.Method, r.URL.Path, recovered)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
