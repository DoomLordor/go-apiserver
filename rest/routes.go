package rest

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

type HandlerFuncRest func(r *http.Request) (any, int, error)

type RouteRest struct {
	Methods     []string
	Pattern     string
	Secure      bool
	Metrics     bool
	HandlerFunc HandlerFuncRest
}

type RoutesRest []*RouteRest
type RouteRestMap map[string]RoutesRest

type HandlerFuncWs func(ctx context.Context, conn *websocket.Conn) (int, error)

type RouteWs struct {
	Pattern     string
	Secure      bool
	HandlerFunc HandlerFuncWs
}

type RoutesWs []*RouteWs
type RouteWsMap map[string]RoutesWs
