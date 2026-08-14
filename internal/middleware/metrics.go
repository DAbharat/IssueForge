package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type Middleware struct {
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec
	httpResponseSize     *prometheus.HistogramVec
}

func NewMetricsMiddleware(reg prometheus.Registerer) *Middleware {
	m := &Middleware{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "issueforge",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed.",
			},
			[]string{"method", "route", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "issueforge",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency histogram in seconds.",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "route"},
		),
		httpRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "issueforge",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Current number of HTTP requests being served.",
			},
			[]string{"method"},
		),
		httpResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "issueforge",
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "HTTP response size in bytes.",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
			},
			[]string{"method", "route"},
		),
	}

	if reg != nil {
		reg.MustRegister(m.httpRequestsTotal, m.httpRequestDuration, m.httpRequestsInFlight, m.httpResponseSize)
	}
	return m
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	wroteHeader  bool
	bytesWritten int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.statusCode = statusCode
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += n
	return n, err
}

func (m *Middleware) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		m.httpRequestsInFlight.WithLabelValues(r.Method).Inc()
		defer m.httpRequestsInFlight.WithLabelValues(r.Method).Dec()

		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		route := "unmatched"
		if currentRoute := mux.CurrentRoute(r); currentRoute != nil {
			if path, err := currentRoute.GetPathTemplate(); err == nil {
				route = path
			}
		}

		status := strconv.Itoa(recorder.statusCode)

		m.httpRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
		m.httpRequestDuration.WithLabelValues(r.Method, route).Observe(float64(time.Since(start).Seconds()))
		m.httpResponseSize.WithLabelValues(r.Method, route).Observe(float64(recorder.bytesWritten))
	})
}
