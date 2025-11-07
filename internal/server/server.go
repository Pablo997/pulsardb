package server

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/Pablo997/pulsardb/internal/config"
	"github.com/Pablo997/pulsardb/pkg/storage"
)

// Server represents the PulsarDB HTTP server
type Server struct {
	config  *config.Config
	storage *storage.Engine
	router  *mux.Router
	server  *http.Server
	
	// Metrics (atomic operations, no mutex needed)
	startTime     time.Time
	pointsWritten int64 // accessed via atomic
	queriesServed int64 // accessed via atomic
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	engine, err := storage.NewEngine(&cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage engine: %w", err)
	}

	s := &Server{
		config:    cfg,
		storage:   engine,
		router:    mux.NewRouter(),
		startTime: time.Now(),
	}

	s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Address, cfg.HTTP.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.storage.Close(); err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}

	return s.server.Shutdown(ctx)
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	// Write endpoint
	s.router.HandleFunc("/write", s.handleWrite).Methods("POST")
	
	// Query endpoint
	s.router.HandleFunc("/query", s.handleQuery).Methods("POST")
	
	// Metrics endpoint
	s.router.HandleFunc("/metrics", s.handleMetrics).Methods("GET")
}

// incrementPointsWritten atomically increments the points written counter
func (s *Server) incrementPointsWritten(count int64) {
	atomic.AddInt64(&s.pointsWritten, count)
}

// incrementQueriesServed atomically increments the queries served counter
func (s *Server) incrementQueriesServed() {
	atomic.AddInt64(&s.queriesServed, 1)
}

// getMetrics returns current metrics (thread-safe via atomic loads)
func (s *Server) getMetrics() (int64, int64, int64) {
	points := atomic.LoadInt64(&s.pointsWritten)
	queries := atomic.LoadInt64(&s.queriesServed)
	uptime := int64(time.Since(s.startTime).Seconds())
	return points, queries, uptime
}

