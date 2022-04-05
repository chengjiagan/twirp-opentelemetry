package oteltwirp

import (
	"io"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// HTTPClient as an interface that models *http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TraceHTTPClient wraps a provided http.Client and tracer for instrumenting
// requests.
type TraceHTTPClient struct {
	client              HTTPClient
	tracer              trace.Tracer
	propagator          propagation.TextMapPropagator
	includeClientErrors bool
}

func NewTraceHTTPClient(client HTTPClient, opts ...Option) *TraceHTTPClient {
	if client == nil {
		client = http.DefaultClient
	}

	cfg := newConfig(opts)
	c := &TraceHTTPClient{client: client}
	c.configure(cfg)
	return c
}

func (c *TraceHTTPClient) configure(cfg *config) {
	c.includeClientErrors = cfg.includeClientErrors
	c.tracer = cfg.tracerProvider.Tracer(instrumentationName)
	c.propagator = cfg.propagator
}

// Do injects the tracing headers into the tracer and updates the headers before
// making the actual request.
func (c *TraceHTTPClient) Do(r *http.Request) (*http.Response, error) {
	ctx := r.Context()
	name, attr := spanInfo(ctx, r.RemoteAddr)
	attr = append(attr, semconv.HTTPClientAttributesFromHTTPRequest(r)...)
	ctx, span := c.tracer.Start(
		ctx,
		name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attr...),
	)

	c.propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))
	r = r.WithContext(ctx)

	span.AddEvent(rpcEventName, trace.WithAttributes(RPCMessageTypeSent))
	resp, err := c.client.Do(r)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return resp, err
	}

	// Check for error codes greater than 400 if withUserErr is set and codes
	// greater than 500 if not, and mark the span as an error if appropriate.
	if resp.StatusCode >= 400 && c.includeClientErrors || resp.StatusCode >= 500 {
		span.SetStatus(codes.Error, "")
	}
	span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)

	// We want to track when the body is closed, meaning the server is done with
	// the response.
	resp.Body = closer{
		ReadCloser: resp.Body,
		span:       span,
	}
	return resp, nil
}

type closer struct {
	io.ReadCloser
	span trace.Span
}

func (c closer) Close() error {
	err := c.ReadCloser.Close()
	c.span.End()
	return err
}
