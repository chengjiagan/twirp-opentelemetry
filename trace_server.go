package oteltwirp

import (
	"context"
	"net/http"
	"strconv"

	"github.com/twitchtv/twirp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const rpcEventName = "message"

type TraceServerHooks struct {
	tracer              trace.Tracer
	includeClientErrors bool
}

// NewOpenTelemetryHooks provides a twirp.ServerHooks struct which records
// OpenTelemetry spans.
func NewOpenTelemetryHooks(opts ...Option) *twirp.ServerHooks {
	traceHooks := &TraceServerHooks{}
	c := newConfig(opts)
	traceHooks.configure(c)

	return traceHooks.TwirpHooks()
}

func (t *TraceServerHooks) configure(c *config) {
	t.tracer = c.tracerProvider.Tracer(instrumentationName)
	t.includeClientErrors = c.includeClientErrors
}

func (t *TraceServerHooks) TwirpHooks() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestReceived: t.startTraceSpan,
		RequestRouted:   t.handleRequestRouted,
		ResponseSent:    t.finishTrace,
		Error:           t.handleError,
	}
}

func (t *TraceServerHooks) startTraceSpan(ctx context.Context) (context.Context, error) {
	ctx, span := t.tracer.Start(ctx, rpcEventName, trace.WithSpanKind(trace.SpanKindServer))
	span.AddEvent(rpcEventName, trace.WithAttributes(RPCMessageTypeReceived))
	return ctx, nil
}

// handleRequestRouted sets the operation name and attributes because we won't
// know what it is until the RequestRouted hook.
func (t *TraceServerHooks) handleRequestRouted(ctx context.Context) (context.Context, error) {
	remoteAddr := ctx.Value(keyRemoteAddr).(string)
	name, attr := spanInfo(ctx, remoteAddr)
	span := trace.SpanFromContext(ctx)
	span.SetName(name)
	span.SetAttributes(attr...)
	return ctx, nil
}

func (t *TraceServerHooks) finishTrace(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	status, haveStatus := twirp.StatusCode(ctx)
	code, err := strconv.Atoi(status)
	if haveStatus && err != nil {
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(code)...)
	}
	span.AddEvent(rpcEventName, trace.WithAttributes(RPCMessageTypeSent))
	span.End()
}

func (t *TraceServerHooks) handleError(ctx context.Context, err twirp.Error) context.Context {
	span := trace.SpanFromContext(ctx)
	statusCode := twirp.ServerHTTPStatusFromErrorCode(err.Code())
	if t.includeClientErrors || statusCode >= 500 {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return ctx
}

// WithTraceContext wraps the handler and extracts the span context from request
// headers to attach to the context for connecting client and server calls.
func WithTraceContext(handler http.Handler, opts ...Option) http.Handler {
	cfg := newConfig(opts)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := cfg.propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx = context.WithValue(ctx, keyRemoteAddr, r.RemoteAddr)
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}
