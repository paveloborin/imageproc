package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	pkgConfig "github.com/paveloborin/imageproc/pkg/flags"
	grpcapi "github.com/paveloborin/imageproc/proto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conf := &pkgConfig.ServerConfig{}
	parser := flags.NewParser(conf, flags.Default)
	if _, err := parser.Parse(); err != nil {
		log.Fatal().Err(err).Msg("parse env error")
	}

	//logger
	zerolog.MessageFieldName = "MESSAGE"
	zerolog.LevelFieldName = "LEVEL"
	zerolog.ErrorFieldName = "ERROR"
	zerolog.TimestampFieldName = "TIME"
	zerolog.SetGlobalLevel(conf.GetLogLevel())
	log.Logger = log.Output(os.Stderr).With().Str("PROGRAM", "imageproc-server").Caller().Logger()
	log.Info().Msgf("log lever is %s", conf.GetLogLevel())

	grpcServer := grpc.NewServer()
	grpcapi.RegisterImageProcServiceServer(grpcServer, NewServer(conf.TmpDir, conf.PyScript))
	errs := make(chan error, 2)

	go func() {
		addr := fmt.Sprintf(":%d", conf.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to listen at %s", addr)
		}

		log.Info().Msgf("Serving service at %s", addr)
		errs <- grpcServer.Serve(listener)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err := <-errs
	log.Info().Err(err).Msg("service terminated")
}
