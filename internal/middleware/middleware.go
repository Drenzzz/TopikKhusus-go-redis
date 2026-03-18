package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	rds "github.com/redis/go-redis/v9"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

type Middleware func(http.Handler) http.Handler

type TrackPayload struct {
	RequestID       string    `json:"request_id"`
	Endpoint        string    `json:"endpoint"`
	Method          string    `json:"method"`
	StatusCode      int       `json:"status_code"`
	ExecutionTimeMS int64     `json:"execution_time_ms"`
	Timestamp       time.Time `json:"timestamp"`
}

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}

	return wrapped
}

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.NewString()
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			recorder := newStatusRecorder(w)

			next.ServeHTTP(recorder, r)

			duration := time.Since(startedAt)
			log.Printf("method=%s path=%s status=%d duration_ms=%d", r.Method, r.URL.Path, recorder.StatusCode(), duration.Milliseconds())
		})
	}
}

func Tracker(track func(ctx context.Context, payload TrackPayload) error) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			recorder := newStatusRecorder(w)

			next.ServeHTTP(recorder, r)

			if track == nil {
				return
			}

			payload := TrackPayload{
				RequestID:       GetRequestID(r.Context()),
				Endpoint:        r.URL.Path,
				Method:          r.Method,
				StatusCode:      recorder.StatusCode(),
				ExecutionTimeMS: time.Since(startedAt).Milliseconds(),
				Timestamp:       time.Now().UTC(),
			}

			if err := track(r.Context(), payload); err != nil {
				log.Printf("tracker save failed: %v", err)
			}
		})
	}
}

func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					log.Printf("panic recovered: %v", recovered)
					writeJSONError(w, http.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimit(client *rds.Client, requestsPerMinute int, timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if client == nil || requestsPerMinute <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			operationCtx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			minuteWindow := time.Now().UTC().Format("200601021504")
			clientIP := requestIP(r.RemoteAddr)
			key := fmt.Sprintf("ratelimit:%s:%s", clientIP, minuteWindow)

			count, err := client.Incr(operationCtx, key).Result()
			if err != nil {
				log.Printf("rate limit check failed: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				if expireErr := client.Expire(operationCtx, key, time.Minute).Err(); expireErr != nil {
					log.Printf("rate limit expire set failed: %v", expireErr)
				}
			}

			if count > int64(requestsPerMinute) {
				writeJSONError(w, http.StatusTooManyRequests, "too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetRequestID(ctx context.Context) string {
	value, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return ""
	}

	return value
}

type statusRecorder struct {
	writer     http.ResponseWriter
	statusCode int
}

func newStatusRecorder(writer http.ResponseWriter) *statusRecorder {
	return &statusRecorder{writer: writer, statusCode: http.StatusOK}
}

func (r *statusRecorder) Header() http.Header {
	return r.writer.Header()
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.writer.WriteHeader(statusCode)
}

func (r *statusRecorder) StatusCode() int {
	return r.statusCode
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	payload := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode json error: %v", err), http.StatusInternalServerError)
	}
}

func requestIP(remoteAddress string) string {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		host = remoteAddress
	}

	if host == "" {
		return "unknown"
	}

	return host
}
