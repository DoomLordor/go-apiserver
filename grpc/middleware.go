package grpc

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/DoomLordor/logger"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
)

type Middlewares struct {
	logger           *logger.Logger
	metricsCollector *grpcprom.ServerMetrics
}

func NewMiddlewares(logger *logger.Logger) *Middlewares {
	return &Middlewares{
		logger:           logger,
		metricsCollector: grpcprom.NewServerMetrics(),
	}
}

func (m *Middlewares) RecoveryMiddleware() recovery.Option {
	grpcPanicRecoveryHandler := func(p any) (err error) {
		m.logger.Err(errors.New("panic")).Msg(string(debug.Stack()))
		return status.Errorf(codes.Internal, "%s", p)
	}

	return recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)
}

func (m *Middlewares) TimeMiddleware() grpc.UnaryServerInterceptor {
	f := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now().UnixMilli()
		m.logger.Info().Str("full_method", info.FullMethod).Msg("Start")

		resp, err := handler(ctx, req)
		statusErr, _ := status.FromError(err)
		end := time.Now().UnixMilli() - start
		m.logger.Info().
			Str("full_method", info.FullMethod).
			Uint64("code", uint64(statusErr.Code())).
			Int64("response_time", end).
			Msg("End")

		return resp, err
	}
	return f
}

func (m *Middlewares) LoggingMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		if err != nil {
			m.logging(info.FullMethod, err)
			return nil, err
		}

		return resp, nil
	}
}

func (m *Middlewares) LoggingStreamMiddleware() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)

		if err != nil {
			m.logging(info.FullMethod, err)
			return err
		}

		return nil
	}
}

func (m *Middlewares) logging(fullMethod string, err error) {
	statusErr, _ := status.FromError(err)
	massageField := ""
	switch statusErr.Code() {
	case codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.PermissionDenied,
		codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unauthenticated:
		massageField = "warning"
	case codes.Unknown, codes.DeadlineExceeded, codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss:
		massageField = "error"
	}

	if massageField != "" {
		m.logger.Warn().Str("full_method", fullMethod).Str(massageField, statusErr.Message()).Send()
	}
}

func (m *Middlewares) MetricsMiddleware() (*grpcprom.ServerMetrics, error) {

	return m.metricsCollector, nil
}
