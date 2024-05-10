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
