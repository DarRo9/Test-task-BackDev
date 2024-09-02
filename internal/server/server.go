package server

import (
	"context"
	"net/http"

	"github.com/DarRo9/Test-task-BackDev/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func CreateObject(config *config.Config, router http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         config.HTTPServer.Address,
			Handler:      router,
			ReadTimeout:  config.HTTPServer.Timeout,
			WriteTimeout: config.HTTPServer.Timeout,
			IdleTimeout:  config.HTTPServer.IdleTimeout,
		},
	}
}
