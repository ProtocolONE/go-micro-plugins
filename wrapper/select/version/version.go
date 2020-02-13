package version

import (
	"context"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/selector"
	"github.com/micro/go-micro/registry"
	"sort"
)

// NewClientWrapper is a wrapper which selects only latest versions of services
func NewClientWrapper(version string, fallback string) client.Wrapper {
	return func(c client.Client) client.Client {
		return &versionWrapper{
			Client:  c,
			Version: version,
			FallbackVersion: fallback,
		}
	}
}

const LatestFallback = "latest"

type versionWrapper struct {
	client.Client
	Version         string
	FallbackVersion string
}

func (w *versionWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	nOpts := append(opts, client.WithSelectOption(selector.WithFilter(FilterVersionWithFallback(w.Version, w.FallbackVersion))))
	return w.Client.Call(ctx, req, rsp, nOpts...)
}

// FilterVersionWithFallback is a version based Select Filter which will
// only return services with the version specified or with fallback version.
func FilterVersionWithFallback(version string, fallback string) selector.Filter {
	return func(old []*registry.Service) []*registry.Service {
		var services []*registry.Service

		for _, service := range old {
			if service.Version == version {
				services = append(services, service)
			}
		}

		if len(services) == 0 {
			if fallback == LatestFallback || len(fallback) == 0 {
				versions := make(map[string]*registry.Service)
				for _, sn := range old {
					v, ok := versions[sn.Name+sn.Version]
					if !ok {
						versions[sn.Name+sn.Version] = sn
						continue
					}
					// append to service:version nodes
					v.Nodes = append(v.Nodes, sn.Nodes...)
				}

				for _, service := range versions {
					services = append(services, service)
				}

				// sort the services
				sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
			} else {
				for _, service := range old {
					if service.Version == fallback {
						services = append(services, service)
					}
				}
			}
		}

		return services
	}
}