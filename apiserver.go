package apiserver

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	"github.com/DoomLordor/logger"

	"github.com/DoomLordor/go-apiserver/debug"
	"github.com/DoomLordor/go-apiserver/grpc"
	"github.com/DoomLordor/go-apiserver/rest"
)

var ConfiguratorNotSetup = errors.New("configurator not setup")

type APIServer struct {
	logger      *logger.Logger
	httpServer  *rest.Server
	debugServer *debug.Server
	grpcServer  *grpc.Server
}

func NewServer(config Config) *APIServer {
	return &APIServer{
		logger:      logger.NewLogger("server"),
		httpServer:  rest.NewServer(config.Rest),
		debugServer: debug.NewServer(config.Debug),
		grpcServer:  grpc.NewServer(config.Grpc),
	}
}

func (s *APIServer) Configuration(context context.Context, configurator Configurator) error {
	s.logger.Info().Msg("Server configuration")

	err := s.configuration(context, configurator)
	if err != nil {
		s.logger.Err(err).Send()
	}
	return err
}

func (s *APIServer) configuration(context context.Context, configurator Configurator) error {

	if configurator == nil {
		return ConfiguratorNotSetup
	}

	adapter, err := configurator.Configure(context)
	if err != nil {
		return err
	}

	if s.httpServer.Active() {
		err = s.httpServer.Configuration(adapter.Api, adapter.Auth)
		if err != nil {
			return err
		}
	}

	if s.grpcServer.Active() {
		err = s.grpcServer.Configuration(adapter.Grps)
		if err != nil {
			return err
		}
	}

	if s.debugServer.Active() {
		s.debugServer.Configuration()
	}

	return nil
}

func (s *APIServer) Start() {
	go s.httpServer.Start()
	go s.grpcServer.Start()
	go s.debugServer.Start()
}

func (s *APIServer) stop(ctx context.Context) error {
	errGroup, _ := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return s.httpServer.Stop(ctx)
	})

	errGroup.Go(func() error {
		s.grpcServer.Stop()
		return nil
	})

	errGroup.Go(func() error {
		return s.debugServer.Stop(ctx)
	})

	return errGroup.Wait()
}

func (s *APIServer) Stop(ctx context.Context, shutdown Shutdown) error {
	errStop := s.stop(ctx)
	if errStop != nil {
		s.logger.Err(errStop).Send()
	}

	var errConf error
	if shutdown != nil {
		errConf = shutdown.Stop(ctx)
		if errConf != nil {
			s.logger.Err(errConf).Send()
		}
	}

	s.logger.Info().Msg("Server stop")
	return errors.Join(errConf, errStop)
}
