package main

import (
	"context"
	"net"

	"go_practice/go_course/7_async_logger/service"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func StartMyMicroservice(ctx context.Context, listenAddr, ACLData string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("Listen TCP port")
	}

	srv := grpc.NewServer()
	service.RegisterAdminServer(srv, newAdminService())
	service.RegisterBizServer(srv, newBizService())

	go func() {
		<-ctx.Done()

		log.Info().Msg("Graceful stopping gRPC server...")
		srv.GracefulStop()
	}()

	go func() {
		log.Info().Msg("Starting gRPC server...")
		if err := srv.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Serve gRPC server")
		}
	}()

	return nil
}
