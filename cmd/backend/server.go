package main

import (
	"context"
	"github.com/h2non/filetype"
	grpcapi "github.com/paveloborin/imageproc/proto"
	"github.com/rs/zerolog/log"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	unsupportedFileMsg = "unsupported file type"
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
		return nil, status.Error(codes.InvalidArgument, unsupportedFileMsg)
	}

	filename := filepath.Join(s.tmpDir, uuid.NewV4().String()+".jpg")
	if err := saveImage(filename, r.File); err != nil {
		log.Error().Err(err).Msg("failed save file to temp dir")
		return nil, status.Error(codes.Unavailable, "failed save file to temp dir")
	}

	defer os.Remove(filename)

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
	if err := ioutil.WriteFile(filename, image, 0644); err != nil {
		return err
	}
	return nil
}
