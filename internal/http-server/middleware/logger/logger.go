package logger

import (
	"fmt"
	"net/http"
	"time"
)

type Logger struct {
	w      http.ResponseWriter
	status int
}

func (rl *Logger) LogRequestInfo(method, path, remoteAddr, userAgent string, duration time.Duration) {
	fmt.Printf("request completed - method: %s, path: %s, remote_addr: %s, user_agent: %s, duration: %s, status_code: %d\n",
		method, path, remoteAddr, userAgent, duration.String(), rl.status)
}

func (rl *Logger) Header() http.Header {
	return rl.w.Header()
}

func CreateObjectLogger(w http.ResponseWriter) *Logger {
	return &Logger{
		w:      w,
		status: http.StatusOK,
	}
}

func (rl *Logger) Write(b []byte) (int, error) {
	return rl.w.Write(b)
}

func (rl *Logger) WriteHeader(statusCode int) {
	rl.status = statusCode
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl := CreateObjectLogger(w)
		t1 := time.Now()

		defer func() {
			rl.LogRequestInfo(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent(), time.Since(t1))
		}()

		next.ServeHTTP(rl, r)
	})
}
