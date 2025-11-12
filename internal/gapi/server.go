package gapi

import (
	"fmt"
	"net"

	"172.21.5.249/airtrans/at-flight-authorization/internal/config"
	"172.21.5.249/airtrans/at-flight-authorization/internal/service"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Config  config.ServiceConfig
	Service *service.Service
}

func NewServer(cfg config.ServiceConfig, svc *service.Service) *Server {
	s := &Server{
		Config:  cfg,
		Service: svc,
	}
	return s
}

func (s *Server) Start(errs chan error) {
	grpcLogger := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		GrpcLogger,
		grpc_prometheus.UnaryServerInterceptor,
		otelgrpc.UnaryServerInterceptor(),
	))
	// embedded logger to grpc server
	grpcServer := grpc.NewServer(grpcLogger)
	// Register pb user service with grpc server
	reflection.Register(grpcServer)
	grpcAddress := fmt.Sprintf("%v:%d", s.Config.GrpcConfig.GrpcHost, s.Config.GrpcConfig.GrpcPort)
	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create grpc listener")
		errs <- err
	}
	log.Info().Msgf("start grpc server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot start grpc server")
		errs <- err
	}
}
