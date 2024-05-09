package grpc

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DoomLordor/logger"
)

type Grps interface {
	RegisterServer(grpcServer *grpc.Server)
}

type Server struct {
	config     Config
	logger     *logger.Logger
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(config Config) *Server {
	return &Server{
		config:     config,
		logger:     logger.NewLogger("server-grpc"),
		grpcServer: nil,
	}
}

func (s *Server) Configuration(grps []Grps) error {
	listener, err := net.Listen("tcp", s.config.BindAddress())
	if err != nil {
		return err
	}

	s.listener = listener

	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(),
	)

	for _, imp := range grps {
		imp.RegisterServer(s.grpcServer)
	}

	reflection.Register(s.grpcServer)

	return nil
}

func (s *Server) Start() {
	if !s.Active() {
		return
	}
	s.logger.Info().Msg("Server grpc start")
	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.logger.Err(err).Send()
	}
}

func (s *Server) Active() bool {
	return s.config.Active
}
