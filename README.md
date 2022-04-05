# OpenTelemetry Hooks for Twirp

The `oteltwirp` package creates an OpenTelemetry Twirp hook to use in your server. Derived from [twirp-opentracing](https://github.com/twirp-ecosystem/twirp-opentracing) and [otelgrpc](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/google.golang.org/grpc/otelgrpc).

## Installation

`go get -u github.com/chengjiagan/twirp-opentelemetry`

## Server-side usage example

Where you are instantiating your Twirp server:

```go
hooks := NewOpenTracingHooks()
service := haberdasherserver.New()
server := WithTraceContext(haberdasher.NewHaberdasherServer(service, hooks))
log.Fatal(http.ListenAndServe(":8080", server))
```

## Client-side usage example

When instantiating your Twirp client:

```go
client := haberdasher.NewHaberdasherProtobufClient(url, NewTraceHTTPClient(http.DefaultClient))
```