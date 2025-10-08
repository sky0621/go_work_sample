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
	scoringhttp "github.com/sky0621/go_work_sample/scoring/internal/http"
	"github.com/sky0621/go_work_sample/scoring/pkg/grading"
)

func main() {
	addr := envOrDefault("SCORING_API_ADDR", ":8091")

	dataPath := envOrDefault("DATA_STORE_PATH", "./data/state.json")
	repo, err := filedb.NewRepository(dataPath, memory.SampleSeed())
	if err != nil {
		log.Fatalf("failed to initialise repository: %v", err)
	}
	assessment := usecase.NewAssessmentService(repo, repo, repo, repo)
	gradingSvc := grading.NewService(assessment)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	scoringhttp.NewHandler(gradingSvc).Register(mux)

	teacherKey := envOrDefault("TEACHER_API_KEY", "teacher-secret")
	authMiddleware := httpmw.APIKey(httpmw.APIKeyConfig{Key: teacherKey, Prefix: "Bearer "})

	server := &http.Server{
		Addr:              addr,
		Handler:           logMiddleware(authMiddleware(mux)),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("scoring-api listening on %s", addr)
		if err := server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("scoring-api shutting down: %s", sig)
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("scoring-api failed: %v", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("scoring-api shutdown error: %v", err)
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
