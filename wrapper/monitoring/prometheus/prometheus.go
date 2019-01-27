package prometheus

import (
	"context"
	"github.com/micro/go-micro/server"
	"github.com/prometheus/client_golang/prometheus"
)

func NewHandlerWrapper() server.HandlerWrapper {
	opsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "go_micro_requests_total",
			Help: "How many go-miro requests processed, partitioned by method and status",
		},
		[]string{"method", "status"},
	)

	timeCounterHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "go_micro_request_duration_seconds",
			Help: "Service method request time in seconds",
		},
		[]string{"method"},
	)

	timeCounterSummary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "go_micro_upstream_latency_seconds",
			Help: "Service backend method request latencies in seconds",
		},
		[]string{"method"},
	)


	prometheus.MustRegister(opsCounter)
	prometheus.MustRegister(timeCounterHistogram)
	prometheus.MustRegister(timeCounterSummary)

	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := req.Endpoint()

			timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
				timeCounterHistogram.WithLabelValues(name).Observe(v)
				timeCounterSummary.WithLabelValues(name).Observe(v)
			}))
			defer timer.ObserveDuration()

			err := fn(ctx, req, rsp)
			if err == nil {
				opsCounter.WithLabelValues(name, "success").Inc()
			} else {
				opsCounter.WithLabelValues(name, "fail").Inc()
			}

			return err
		}
	}
}
