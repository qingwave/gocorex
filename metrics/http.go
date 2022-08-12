package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(opt *stateOption)

func WithNamespace(namespace string) Option {
	return func(opt *stateOption) {
		opt.namespace = namespace
	}
}

func WithRegistry(registry Registry) Option {
	return func(opt *stateOption) {
		opt.registry = registry
	}
}

func newStateOption() *stateOption {
	return &stateOption{
		namespace: "",
		subsystem: "http",
	}
}

type stateOption struct {
	namespace string
	subsystem string
	registry  Registry
}

type Registry interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type HttpState struct {
	option *stateOption

	requestsTotal    *prometheus.CounterVec
	inflightRequests *prometheus.GaugeVec
	requestsDuration *prometheus.HistogramVec
	requestSize      *prometheus.HistogramVec
	responseSize     *prometheus.HistogramVec
}

func NewHttpState(opts ...Option) *HttpState {
	option := newStateOption()
	for _, opt := range opts {
		opt(option)
	}

	state := &HttpState{
		option: option,
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: option.namespace,
				Subsystem: option.subsystem,
				Name:      "requests_total",
				Help:      "Number of get requests.",
			},
			[]string{"method", "path", "code"},
		),
		inflightRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: option.namespace,
				Subsystem: option.subsystem,
				Name:      "inflight_requests",
				Help:      "Status of HTTP response.",
			},
			[]string{"method", "path"},
		),
		requestsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: option.namespace,
				Subsystem: option.subsystem,
				Name:      "requests_duration_seconds",
				Help:      "Duration of HTTP requests.",
			},
			[]string{"method", "path"},
		),
		requestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: option.namespace,
				Subsystem: option.subsystem,
				Name:      "request_size_bytes",
				Help:      "Size of HTTP requests.",
			},
			[]string{"method", "path"},
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: option.namespace,
				Subsystem: option.subsystem,
				Name:      "response_size_bytes",
				Help:      "Size of HTTP responses.",
			},
			[]string{"method", "path"},
		),
	}

	state.register()

	return state
}

func (s *HttpState) register() {
	collector := []prometheus.Collector{
		s.requestsTotal,
		s.inflightRequests,
		s.requestsDuration,
		s.requestSize,
		s.responseSize,
	}

	registerer := prometheus.DefaultRegisterer
	if s.option.registry != nil {
		registerer = s.option.registry
	}

	for _, c := range collector {
		registerer.MustRegister(c)
	}
}

func (s *HttpState) Handler() http.Handler {
	if s.option.registry != nil {
		return promhttp.HandlerFor(s.option.registry, promhttp.HandlerOpts{})
	}
	return promhttp.Handler()
}

func (s *HttpState) WrapMetrics(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		s.inflightRequests.WithLabelValues(method, path).Inc()
		s.requestSize.WithLabelValues(method, path).Observe(float64(r.ContentLength))

		delegate := &ResponseWriterDelegator{ResponseWriter: w}

		defer func() {
			duration := time.Since(start).Seconds()
			code := codeToString(delegate.Status())
			size := float64(delegate.ContentLength())

			s.inflightRequests.WithLabelValues(method, path).Dec()
			s.requestsTotal.WithLabelValues(method, path, code).Inc()
			s.requestsDuration.WithLabelValues(method, path).Observe(duration)
			s.responseSize.WithLabelValues(method, path).Observe(size)
		}()

		h.ServeHTTP(delegate, r)
	})
}

var _ http.ResponseWriter = (*ResponseWriterDelegator)(nil)

// ResponseWriterDelegator interface wraps http.ResponseWriter to additionally record content-length, status-code, etc.
type ResponseWriterDelegator struct {
	http.ResponseWriter

	status      int
	written     int64
	wroteHeader bool
}

func (r *ResponseWriterDelegator) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *ResponseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *ResponseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func (r *ResponseWriterDelegator) Status() int {
	return r.status
}

func (r *ResponseWriterDelegator) ContentLength() int {
	return int(r.written)
}

// Small optimization over Itoa
func codeToString(s int) string {
	switch s {
	case 100:
		return "100"
	case 101:
		return "101"

	case 200:
		return "200"
	case 201:
		return "201"
	case 202:
		return "202"
	case 203:
		return "203"
	case 204:
		return "204"
	case 205:
		return "205"
	case 206:
		return "206"

	case 300:
		return "300"
	case 301:
		return "301"
	case 302:
		return "302"
	case 304:
		return "304"
	case 305:
		return "305"
	case 307:
		return "307"

	case 400:
		return "400"
	case 401:
		return "401"
	case 402:
		return "402"
	case 403:
		return "403"
	case 404:
		return "404"
	case 405:
		return "405"
	case 406:
		return "406"
	case 407:
		return "407"
	case 408:
		return "408"
	case 409:
		return "409"
	case 410:
		return "410"
	case 411:
		return "411"
	case 412:
		return "412"
	case 413:
		return "413"
	case 414:
		return "414"
	case 415:
		return "415"
	case 416:
		return "416"
	case 417:
		return "417"
	case 418:
		return "418"

	case 500:
		return "500"
	case 501:
		return "501"
	case 502:
		return "502"
	case 503:
		return "503"
	case 504:
		return "504"
	case 505:
		return "505"

	case 428:
		return "428"
	case 429:
		return "429"
	case 431:
		return "431"
	case 511:
		return "511"

	default:
		return strconv.Itoa(s)
	}
}
