package debug

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/DoomLordor/logger"
)

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
		WriteTimeout: time.Second * 100,
		ReadTimeout:  time.Second * 100,
		IdleTimeout:  time.Second * 100,
		Handler:      router,
	}
	return &Server{
		router:     router,
		httpServer: httpServer,
		logger:     logger.NewLogger("rest-server"),
	}
}

func (s *Server) Configuration() {
	s.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	router := s.router.PathPrefix("/debug/pprof").Subrouter()
	router.HandleFunc("/", pprof.Index)
	router.HandleFunc("/cmdline", pprof.Cmdline)
	router.HandleFunc("/profile", pprof.Profile)
	router.HandleFunc("/symbol", pprof.Symbol)
	router.Handle("/goroutine", pprof.Handler("goroutine"))
	router.Handle("/heap", pprof.Handler("heap"))
	router.Handle("/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/block", pprof.Handler("block"))
}

func (s *Server) Start() {
	if !s.Active() {
		return
	}
	s.logger.Info().Msg("Server debug start")
	if err := s.httpServer.ListenAndServe(); err != nil {
		s.logger.Err(err).Send()
	}
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Active() bool {
	return s.config.Debug
}
