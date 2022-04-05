package oteltwirp

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// config is a group of options for this instrumentation.
type config struct {
	propagator          propagation.TextMapPropagator
	tracerProvider      trace.TracerProvider
	includeClientErrors bool
}

// Option applies an option value for a config.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// newConfig returns a config configured with all the passed Options.
func newConfig(opts []Option) *config {
	c := &config{
		propagator:          otel.GetTextMapPropagator(),
		tracerProvider:      otel.GetTracerProvider(),
		includeClientErrors: true,
	}
	for _, o := range opts {
		o.apply(c)
	}
	return c
}

// WithPropagators returns an Option to use the Propagators when extracting
// and injecting trace context from requests.
func WithPropagators(p propagation.TextMapPropagator) Option {
	return optionFunc(func(c *config) {
		if p != nil {
			c.propagator = p
		}
	})
}

// WithTracerProvider returns an Option to use the TracerProvider when
// creating a Tracer.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return optionFunc(func(c *config) {
		if tp != nil {
			c.tracerProvider = tp
		}
	})
}

// IncludeClientErrors, if set, will report client errors (4xx) as errors in the server span.
// If not set, only 5xx status will be reported as erroneous.
func IncludeClientErrors(include bool) Option {
	return optionFunc(func(c *config) {
		c.includeClientErrors = include
	})
}
