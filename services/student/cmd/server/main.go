package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sky0621/go_work_sample/core/pkg/httpmw"
	"github.com/sky0621/go_work_sample/core/pkg/memory"
	"github.com/sky0621/go_work_sample/core/pkg/storage/filedb"
	"github.com/sky0621/go_work_sample/core/pkg/usecase"
	studenthttp "github.com/sky0621/go_work_sample/student/internal/http"
)

func main() {
	addr := envOrDefault("STUDENT_API_ADDR", ":8081")

	dataPath := envOrDefault("DATA_STORE_PATH", "./data/state.json")
	repo, err := filedb.NewRepository(dataPath, memory.SampleSeed())
	if err != nil {
		log.Fatalf("failed to initialise repository: %v", err)
	}
	assessment := usecase.NewAssessmentService(repo, repo, repo, repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	studenthttp.NewHandler(assessment).Register(mux)

	studentKey := envOrDefault("STUDENT_API_KEY", "student-secret")
	authMiddleware := httpmw.APIKey(httpmw.APIKeyConfig{Key: studentKey, Prefix: "Bearer "})

	server := &http.Server{
		Addr:              addr,
		Handler:           logMiddleware(authMiddleware(mux)),
		ReadTimeout:       3 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      6 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("student-api listening on %s", addr)
		if err := server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("student-api shutting down: %s", sig)
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("student-api failed: %v", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("student-api shutdown error: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
