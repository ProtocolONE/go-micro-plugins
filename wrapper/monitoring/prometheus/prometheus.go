package prometheus

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/server"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
)

type counters struct {
	opsCounter  *prometheus.CounterVec
	timeCounter prometheus.Summary
}

func NewHandlerWrapper(eps ...interface{}) server.HandlerWrapper {
	endpoints := make(map[string]counters)

	for _, ep := range eps {
		fooType := reflect.TypeOf(ep).Elem()
		interfaceName := fooType.Name()

		for i := 0; i < fooType.NumMethod(); i++ {
			methodName := fooType.Method(i).Name

			endpointName := interfaceName + "." + methodName
			counterName := strings.ToLower(interfaceName + "_" + methodName)

			endpoints[endpointName] = counters{
				opsCounter: prometheus.NewCounterVec(
					prometheus.CounterOpts{
						Name: counterName + "_requests_total",
						Help: fmt.Sprintf("How many %s requests processed, partitioned by status", endpointName),
					},
					[]string{"status"},
				),
				timeCounter: prometheus.NewSummary(
					prometheus.SummaryOpts{
						Name: counterName + "_request_durations",
						Help: fmt.Sprintf("%s requests latencies in seconds", endpointName),
					},
				),
			}

			prometheus.MustRegister(endpoints[endpointName].opsCounter)
			prometheus.MustRegister(endpoints[endpointName].timeCounter)
		}
	}

	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			val, ok := endpoints[req.Endpoint()]
			if !ok {
				return fn(ctx, req, rsp)
			}

			timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
				us := v * 1000000 // make microseconds
				val.timeCounter.Observe(us)
			}))
			defer timer.ObserveDuration()

			err := fn(ctx, req, rsp)
			if err == nil {
				val.opsCounter.WithLabelValues("success").Inc()
			} else {
				val.opsCounter.WithLabelValues("fail").Inc()
			}

			return err
		}
	}
}
