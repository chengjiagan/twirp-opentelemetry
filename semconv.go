package oteltwirp

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const instrumentationName = "github.com/chengjiagan/twirp-opentelemetry"

// Semantic conventions for attribute keys for twirp.
const (
	// Name of message transmitted or received.
	RPCNameKey = attribute.Key("name")

	// Type of message transmitted or received.
	RPCMessageTypeKey = attribute.Key("message.type")
)

// Semantic conventions for common RPC attributes.
var (
	// Semantic convention for gRPC as the remoting system.
	RPCSystemTwirp = semconv.RPCSystemKey.String("twrip")

	// Semantic convention for a message named message.
	RPCNameMessage = RPCNameKey.String("message")

	// Semantic conventions for RPC message types.
	RPCMessageTypeSent     = RPCMessageTypeKey.String("SENT")
	RPCMessageTypeReceived = RPCMessageTypeKey.String("RECEIVED")
)
