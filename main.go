package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

const maxRetries = 3

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

// ...

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	serviceName := "dice"
	serviceVersion := "0.1.0"
	otelShutdown, err := setupOTelSDK(ctx, serviceName, serviceVersion)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// Create a WaitGroup to wait for the server to finish processing existing requests.
	var wg sync.WaitGroup
	wg.Add(1)

	// Start a goroutine to gracefully shut down the server.
	go func() {
		defer wg.Done()

		// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
		err = srv.Shutdown(context.Background())
		if err != nil {
			log.Printf("Error during server shutdown: %v\n", err)
		}
	}()

	// Wait for the server to finish processing existing requests.
	wg.Wait()

	return
}

// ...

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// Register handlers.
	handleFunc("/rolldice", withRetry(rolldice, maxRetries))

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

func withRetry(fn func(http.ResponseWriter, *http.Request), maxRetries int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a span for the retry logic.
		_, retrySpan := otel.Tracer("main").Start(r.Context(), "retry-span")
		defer retrySpan.End()

		var lastError error

		// Initial retries
		for retryCounter := 0; retryCounter < maxRetries; retryCounter++ {
			// Create a span for each retry attempt.
			_, attemptSpan := otel.Tracer("main").Start(r.Context(), "retry-attempt-span")
			defer attemptSpan.End()

			lastError = recoverFromPanic(func() {
				fn(w, r)
			})

			if lastError == nil {
				return // Function succeeded, return response
			}

			log.Printf("Retry %d failed with error: %v\n", retryCounter+1, lastError)
			retrySpan.RecordError(lastError)

			time.Sleep(time.Second) // Add a delay before retrying
		}

		// Log a message when reaching max retries
		log.Println("Max retries reached, giving up.")
		return
	}
}

func recoverFromPanic(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("panic occurred")
		}
	}()
	fn()
	return
}
