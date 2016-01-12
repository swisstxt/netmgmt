package logger

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/negroni"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{log.New(os.Stdout, "", 0)}
}

type logmgs struct {
	Timestamp string        `json:"timestamp"`
	Status    int           `json:"status"`
	Method    string        `json:"method"`
	Request   string        `json:"request"`
	Latency   time.Duration `json:"latency"`
}

func (l *Logger) ServeHTTP(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(res, req)
	end := time.Now()

	r := res.(negroni.ResponseWriter)
	log := &logmgs{
		Timestamp: end.Format("2006/01/02-15:04:05.000"),
		Status:    r.Status(),
		Latency:   end.Sub(start),
		Method:    req.Method,
		Request:   req.URL.Path,
	}

	b, _ := json.Marshal(log)
	l.Println(string(b))
}
