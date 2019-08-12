package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	pkgConfig "github.com/paveloborin/imageproc/pkg/flags"
	grpcapi "github.com/paveloborin/imageproc/proto"
)

func main() {
	conf := &pkgConfig.ServerConfig{}
	parser := flags.NewParser(conf, flags.Default)
	if _, err := parser.Parse(); err != nil {
		log.Fatal().Err(err).Msg("parse env error")
	}

	zerolog.MessageFieldName = "MESSAGE"
	zerolog.LevelFieldName = "LEVEL"
	zerolog.ErrorFieldName = "ERROR"
	zerolog.TimestampFieldName = "TIME"
	zerolog.SetGlobalLevel(conf.GetLogLevel())
	log.Logger = log.Output(os.Stderr).With().Str("PROGRAM", "imageproc-server").Logger()
	log.Info().Msgf("log lever is %s", conf.GetLogLevel())

	if err := os.MkdirAll(conf.TmpDir, os.ModePerm); err != nil {
		log.Panic().Err(err).Msgf("failed create temporary dir")
	}

	addr := fmt.Sprintf(":%d", conf.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panic().Err(err).Msgf("failed to listen at %s", addr)
	}

	gracefulShChan := make(chan interface{})

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)

		select {
		case s := <-c:
			log.Info().Msgf("os signal received %v", s)
			close(gracefulShChan)
		case <-gracefulShChan:
			return
		}
	}()

	log.Info().Msgf("starting service at %s", addr)
	server := grpc.NewServer()
	grpcapi.RegisterImageProcServiceServer(server, NewServer(conf.TmpDir, conf.PyScript))

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Error().Err(err).Msg("listener error")
		}

		select {
		case <-gracefulShChan:
			return
		default:
			close(gracefulShChan)
		}
	}()

	<-gracefulShChan
	server.GracefulStop()
	log.Info().Msg("service terminated")
}
