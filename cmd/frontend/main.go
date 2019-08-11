package main

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	pkgConfig "github.com/paveloborin/imageproc/pkg/flags"
	grpcapi "github.com/paveloborin/imageproc/proto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func main() {
	conf := &pkgConfig.ClientConfig{}
	parser := flags.NewParser(conf, flags.Default)
	if _, err := parser.Parse(); err != nil {
		log.Fatal().Err(err).Msg("parse env error")
	}

	zerolog.MessageFieldName = "MESSAGE"
	zerolog.LevelFieldName = "LEVEL"
	zerolog.ErrorFieldName = "ERROR"
	zerolog.SetGlobalLevel(conf.GetLogLevel())
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Logger()
	log.Info().Msgf("log lever is %s", conf.GetLogLevel())

	conn := connectToGRPCService(conf.ServerHost, conf.ServerPort)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close the cloud GRPC service connection")
		}
	}()
	grpcClient := grpcapi.NewImageProcServiceClient(conn)

	files, err := ioutil.ReadDir(conf.InputPath)
	if err != nil {
		log.Panic().Err(err).Msgf("failed read dir %s", conf.InputPath)
	}

	log.Info().Msgf("input dir: %s", conf.InputPath)
	log.Info().Msgf("output dir: %s", conf.OutputPath)

	for _, im := range files {
		log.Debug().Msgf("start processing image with name: %s", im.Name())
		filename := filepath.Join(conf.InputPath, im.Name())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Error().Err(err).Msg("failed to read image")
			continue
		}

		file, err := uploadImage(grpcClient, data)
		if err != nil {
			log.Error().Err(err).Msg("failed to upload image")
			continue
		}

		if err := saveImage(conf.OutputPath, im.Name(), file); err != nil {
			log.Error().Err(err).Msg("failed save image")
			continue
		}
	}

}

func connectToGRPCService(host string, port int) *grpc.ClientConn {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to grpc service at %s", address)
		panic(err)
	}
	return conn
}

func uploadImage(client grpcapi.ImageProcServiceClient, image []byte) ([]byte, error) {
	req := &grpcapi.Request{File: image}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	reply, err := client.Upload(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply.File, nil
}

func saveImage(path, name string, image []byte) error {
	log.Debug().Msgf("try to save file %s", name)
	filename := filepath.Join(path, name)
	if err := ioutil.WriteFile(filename, image, 0644); err != nil {
		return err
	}
	return nil
}
