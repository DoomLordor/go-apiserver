package apiserver

import (
	"context"

	"github.com/DoomLordor/go-apiserver/grpc"
	"github.com/DoomLordor/go-apiserver/rest"
)

type Adapter struct {
	Auth rest.AuthFunc
	Api  []rest.Api
	Grps []grpc.Grps
}

type Configurator interface {
	Configure(ctx context.Context) (*Adapter, error)
}
