package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcapi "github.com/paveloborin/imageproc/proto"
)

type Server struct {
	tmpDir   string
	pyScript string
}

func NewServer(tmpDir string, pyScript string) *Server {
	return &Server{tmpDir: tmpDir, pyScript: pyScript}
}

func (s *Server) Upload(ctx context.Context, r *grpcapi.Request) (*grpcapi.Reply, error) {
	log.Debug().Msgf("image size is %d", len(r.File))
	if !filetype.IsImage(r.File) {
		return nil, status.Error(codes.InvalidArgument, "unsupported file type")
	}

	filename := filepath.Join(s.tmpDir, fmt.Sprintf("%s.jpg", uuid.NewV4().String()))
	if err := saveImage(filename, r.File); err != nil {
		log.Error().Err(err).Msg("failed save file to temp dir")
		return nil, status.Error(codes.Unavailable, "failed save file to temp dir")
	}

	defer func() {
		if err := os.Remove(filename); err != nil {
			log.Error().Err(err).Msg("failed remove file")
		}
	}()

	cmd := exec.Command("python3", s.pyScript, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msg("failed image processing")
		return nil, status.Error(codes.Unavailable, "failed image processing")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error().Err(err).Msg("failed to read image")
		return nil, status.Error(codes.Unavailable, "failed image processing")
	}

	return &grpcapi.Reply{File: data}, nil
}

func saveImage(filename string, image []byte) error {
	log.Debug().Msgf("try to save file %s", filename)
	return ioutil.WriteFile(filename, image, 0644)
}
