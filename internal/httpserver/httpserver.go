// Package httpserver wraps http.Server with shared timeouts and graceful
// shutdown so every service stops accepting slowloris-style connections and
// drains in-flight requests on SIGINT/SIGTERM.
package httpserver

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 60 * time.Second
	// Handlers that fan out to the WCL API can legitimately run past the
	// 120s service-to-service client timeout; keep the write timeout above it.
	writeTimeout    = 300 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 30 * time.Second
)

// ListenAndServe serves handler on addr until the process receives
// SIGINT/SIGTERM, then drains in-flight requests. Returns nil on a clean
// shutdown.
func ListenAndServe(name, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		log.Printf("%s: received %s, shutting down", name, sig)
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}
}
