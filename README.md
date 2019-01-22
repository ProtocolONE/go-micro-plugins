# go-micro-plugins

[![Build Status](https://travis-ci.org/ProtocolONE/go-micro-plugins.svg?branch=master)](https://travis-ci.org/ProtocolONE/go-micro-plugins) [![codecov](https://codecov.io/gh/ProtocolONE/go-micro-plugins/branch/master/graph/badge.svg)](https://codecov.io/gh/ProtocolONE/go-micro-plugins)

A repository for go-micro protocol one plugins.

 # Overview
 
 Micro tooling is built on a powerful pluggable architecture. Plugins can be swapped out with zero code changes. This repository contains plugins for all micro related tools. Read on for further info.
 
 ## Contents
 
 Contents of this repository:
 
 | Directory | Description                                          |
 | --------- | ---------------------------------------------------- |
 | Wrappers  | Prometheus                                           |
 
  ## Usage
  
Plugins can be added to go-micro in the following ways. By doing so they'll be available to set via command line args or environment variables.

Import the plugins in a Go program then call service.Init to parse the command line and environment variables.

```go
import (
	"github.com/micro/go-micro"
	prometheus_plugin "github.com/protocolone/go-micro-plugins/wrapper/minitoring/prometheus"
)

func main() {
	service := micro.NewService(
		// Set service name
		micro.Name("my.service"),
		
		// Register wrapper
		micro.WrapHandler(prometheus_plugin.NewHandlerWrapper((*proto.MyServiceInterface)(nil))),		
	)

	// Parse CLI flags
	service.Init()
}
```  