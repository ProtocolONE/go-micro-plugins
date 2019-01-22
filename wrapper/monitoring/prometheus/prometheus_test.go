package prometheus

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry/memory"
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/server"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Test interface {
	Method(ctx context.Context, in *TestRequest, opts ...client.CallOption) (*TestResponse, error)
}

type TestRequest struct {
	IsError bool
}
type TestResponse struct{}

type testHandler struct{}

func (t *testHandler) Method(ctx context.Context, req *TestRequest, rsp *TestResponse) error {
	if req.IsError {
		return fmt.Errorf("test error")
	}
	return nil
}

func TestPrometheusMetrics(t *testing.T) {
	// setup
	registry := memory.NewRegistry()
	sel := selector.NewSelector(selector.Registry(registry))
	name := "test"

	c := client.NewClient(client.Selector(sel))
	s := server.NewServer(
		server.Name(name),
		server.Registry(registry),
		server.WrapHandler(NewHandlerWrapper((*Test)(nil))),
	)

	type Test struct {
		*testHandler
	}

	s.Handle(
		s.NewHandler(&Test{new(testHandler)}),
	)

	if err := s.Start(); err != nil {
		t.Fatalf("Unexpected error starting server: %v", err)
	}

	if err := s.Register(); err != nil {
		t.Fatalf("Unexpected error registering server: %v", err)
	}

	req := c.NewRequest(name, "Test.Method", &TestRequest{IsError: false}, client.WithContentType("application/json"))
	rsp := TestResponse{}

	assert.NoError(t, c.Call(context.TODO(), req, &rsp))

	req = c.NewRequest(name, "Test.Method", &TestRequest{IsError: true}, client.WithContentType("application/json"))
	assert.Error(t, c.Call(context.TODO(), req, &rsp))

	list, _ := prometheus.DefaultGatherer.Gather()

	metric := findMetricByName(list, dto.MetricType_SUMMARY, "test_method_request_durations")
	assert.NotNil(t, metric)
	assert.Equal(t, *metric.Metric[0].Summary.SampleCount, uint64(2))

	metric = findMetricByName(list, dto.MetricType_COUNTER, "test_method_requests_total")
	assert.Equal(t, *metric.Metric[0].Label[0].Value, "fail")
	assert.Equal(t, *metric.Metric[0].Counter.Value, float64(1))

	assert.Equal(t, *metric.Metric[1].Label[0].Value, "success")
	assert.Equal(t, *metric.Metric[1].Counter.Value, float64(1))

	s.Deregister()
	s.Stop()
}

func findMetricByName(list []*dto.MetricFamily, tp dto.MetricType, name string) *dto.MetricFamily {
	for _, metric := range list {
		if *metric.Name == name && *metric.Type == tp {
			return metric
		}
	}

	return nil
}
