package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Server struct {
	handler http.Handler
	started chan string

	network       string
	serverAddress string
}

func New(network string, serverAddress string, handler http.Handler) *Server {
	return &Server{
		network:       network,
		serverAddress: serverAddress,
		handler:       handler,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	listener, err := s.getListener()
	if err != nil {
		return err
	}

	if listener == nil {
		return nil
	}

	defer listener.Close()

	srv := &http.Server{
		Handler:        s.handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Starting HTTP server on", listener.Addr())

	// testing
	if s.started != nil {
		select {
		case <-ctx.Done():
		case s.started <- listener.Addr().String():
		}
	}

	serveError := make(chan error, 1)
	go func() {
		select {
		case serveError <- srv.Serve(listener):
		case <-ctx.Done():
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Shutting down HTTP server")
	case err := <-serveError:
		fmt.Println("HTTP server error:", err)
	}

	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := srv.Shutdown(timeout); err == timeout.Err() {
		srv.Close()
	}

	return err
}

func (s *Server) getListener() (net.Listener, error) {
	l, err := net.Listen(s.network, s.serverAddress)
	if err != nil {
		return nil, err
	}

	return l, nil
}
