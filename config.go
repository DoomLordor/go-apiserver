package apiserver

import (
	"github.com/DoomLordor/go-apiserver/debug"
	"github.com/DoomLordor/go-apiserver/grpc"
	"github.com/DoomLordor/go-apiserver/rest"
)

type Config struct {
	Rest  rest.Config
	Debug debug.Config
	Grpc  grpc.Config
}

type JaegerConfig struct {
	JaegerGRPCAddr string `env:"JAEGER_GRPC_ADDR" envDefault:"localhost:4317"`
	ServiceName    string `env:"JAEGER_SERVICE_NAME" envDefault:""`
}
