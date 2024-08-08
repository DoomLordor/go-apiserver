package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/logger"
)

type Api interface {
	RegistrationRest() RouteRestMap
	RegistrationWs() RouteWsMap
}

type Server struct {
	config     Config
	router     *mux.Router
	httpServer *http.Server
	logger     *logger.Logger
}

func NewServer(config Config) *Server {
	router := mux.NewRouter()
	httpServer := &http.Server{
		Addr:         config.BindAddress(),
		WriteTimeout: config.WriteTimeout,
		ReadTimeout:  config.ReadTimeout,
		IdleTimeout:  config.IdleTimeout,
		Handler:      router,
	}
	return &Server{
		config:     config,
		router:     router,
		httpServer: httpServer,
		logger:     logger.NewLogger("rest-server"),
	}
}

func (s *Server) Configuration(api []Api, authFunc AuthFunc, tracer trace.Tracer) error {
	s.logger.Info().Msg("Router configuration")
	metrics, err := NewPrometheusService()
	if err != nil {
		return err
	}

	m := NewMiddlewares(authFunc, logger.NewLogger("middlewares-rest"), tracer)
	s.router.Use(m.RecoveryMiddleware)
	routerRest := s.router.PathPrefix("/api/v1").Subrouter()
	routerRest.Use(m.CommonMiddleware)
	routerRest.Use(m.TimeMiddleware)

	routerWs := s.router.PathPrefix("/ws").Subrouter()
	routerWs.Use(m.LoggingMiddleware)

	for _, a := range api {
		routeMap := a.RegistrationRest()
		for prefix, routes := range routeMap {
			sub := routerRest.PathPrefix(prefix).Subrouter()

			for _, route := range routes {
				handlerFunc := m.TracingMiddleware(route.HandlerFunc)
				handler := m.HandleWrapper(handlerFunc)
				if route.Secure {
					handler = m.TokenMiddleware(handler)
				}

				r := sub.Path(route.Pattern)
				if route.Metrics {
					path, _ := r.GetPathTemplate()
					handler = metrics.RequestMetricsMiddleware(path, handler)
				}

				r.Handler(handler).Methods(route.Methods...)
			}
		}

		routeMapWs := a.RegistrationWs()
		for prefix, routes := range routeMapWs {
			sub := routerWs.PathPrefix(prefix).Subrouter()

			for _, route := range routes {
				handler := m.HandleWsWrapper(route.HandlerFunc)
				if route.Secure {
					handler = m.TokenMiddleware(handler)
				}
				sub.Handle(route.Pattern, handler).Methods(http.MethodGet)
			}
		}
	}

	routerRest.Handle("", m.HandleWrapper(s.urls)).Methods(http.MethodGet)
	s.router.NotFoundHandler = s.router.NewRoute().HandlerFunc(notFound).GetHandler()

	return nil
}

func (s *Server) urls(_ *http.Request) (any, int, error) {
	res := make(map[string][]string, 10)

	_ = s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		template, _ := route.GetPathTemplate()
		met, _ := route.GetMethods()
		if len(met) == 0 {
			return nil
		}
		methods, ok := res[template]
		if !ok {
			methods = make([]string, 0, len(met))
		}
		res[template] = append(methods, met...)
		return nil
	})
	return res, http.StatusOK, nil
}

func (s *Server) Start() {
	if !s.Active() {
		return
	}
	s.logger.Info().Msg("Server rest start")
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatal().Err(err).Send()
	}
}

func (s *Server) stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.stop(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		} else {
			s.logger.Err(err).Send()
		}
	}
	s.logger.Info().Msg("Server stop")
	return err
}

func (s *Server) Active() bool {
	return s.config.Active
}
