package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

type blockingShutdownServer struct {
	serveStarted    chan struct{}
	shutdownStarted chan struct{}
	finishShutdown  chan struct{}
}

func newBlockingShutdownServer() *blockingShutdownServer {
	return &blockingShutdownServer{
		serveStarted:    make(chan struct{}),
		shutdownStarted: make(chan struct{}),
		finishShutdown:  make(chan struct{}),
	}
}

func (s *blockingShutdownServer) ListenAndServe() error {
	close(s.serveStarted)
	<-s.shutdownStarted
	return http.ErrServerClosed
}

func (s *blockingShutdownServer) Shutdown(ctx context.Context) error {
	close(s.shutdownStarted)
	select {
	case <-s.finishShutdown:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestServeHTTPWaitsForShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := newBlockingShutdownServer()
	done := make(chan error, 1)

	go func() {
		done <- serveHTTP(ctx, server, time.Second)
	}()

	<-server.serveStarted
	cancel()
	<-server.shutdownStarted

	select {
	case err := <-done:
		t.Fatalf("serveHTTP returned before shutdown completed: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	close(server.finishShutdown)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serveHTTP() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("serveHTTP did not return after shutdown completed")
	}
}
