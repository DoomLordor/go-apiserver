package grpc

import (
	"google.golang.org/grpc/metadata"
)

// metadataCarrier type for using MD as open telemetry TextMapCarrier
type metadataCarrier metadata.MD

func (mc metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(metadata.MD(mc)))
	for key := range metadata.MD(mc) {
		keys = append(keys, key)
	}
	return keys
}

func (mc metadataCarrier) Get(key string) string {
	if values := metadata.MD(mc).Get(key); len(values) > 0 {
		return values[0]
	}
	return ""
}

func (mc metadataCarrier) Set(key string, value string) {
	metadata.MD(mc).Append(key, value)
}
