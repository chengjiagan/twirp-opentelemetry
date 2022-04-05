package oteltwirp

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/twitchtv/twirp"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type oteltwirpKey string

const keyRemoteAddr oteltwirpKey = "OteltwirpRemoteAddr"

// hostIPNamePort extracts the IP address, name and (optional) port from hostWithPort.
// It handles both IPv4 and IPv6 addresses. If the host portion is not recognized
// as a valid IPv4 or IPv6 address, the `ip` result will be empty and the
// host portion will instead be returned in `name`.
func hostIPNamePort(hostWithPort string) (ip string, name string, port int) {
	var (
		hostPart, portPart string
		parsedPort         uint64
		err                error
	)
	if hostPart, portPart, err = net.SplitHostPort(hostWithPort); err != nil {
		hostPart, portPart = hostWithPort, ""
	}
	if parsedIP := net.ParseIP(hostPart); parsedIP != nil {
		ip = parsedIP.String()
	} else {
		name = hostPart
	}
	if parsedPort, err = strconv.ParseUint(portPart, 10, 16); err == nil {
		port = int(parsedPort)
	}
	return
}

// spanInfo returns a span name and all appropriate attributes from the context
func spanInfo(ctx context.Context, addr string) (string, []attribute.KeyValue) {
	packageName, _ := twirp.PackageName(ctx)
	serviceName, _ := twirp.ServiceName(ctx)
	methodName, _ := twirp.MethodName(ctx)
	attr := []attribute.KeyValue{
		semconv.RPCServiceKey.String(serviceName),
		semconv.RPCMethodKey.String(methodName),
		semconv.NetTransportTCP,
	}
	spanName := fmt.Sprintf("%s.%s/%s", packageName, serviceName, methodName)

	ip, name, port := hostIPNamePort(addr)
	if ip != "" {
		attr = append(attr, semconv.NetPeerIPKey.String(ip))
	}
	if name != "" {
		attr = append(attr, semconv.NetPeerNameKey.String(name))
	}
	if port != 0 {
		attr = append(attr, semconv.NetPeerPortKey.Int(port))
	}

	return spanName, attr
}
