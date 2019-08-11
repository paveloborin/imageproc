package flags

import (
	"strings"

	"github.com/rs/zerolog"
)

type ServerConfig struct {
	Port     int    `long:"grpc-port" env:"GRPC_PORT" required:"true"`
	TmpDir   string `long:"tmp-dir" env:"TMP_DIR" required:"true"`
	PyScript string `long:"py-script" env:"PY_SCRIPT" required:"true"`
	config
}

type ClientConfig struct {
	InputPath  string `long:"input-path" env:"INPUT_PATH" required:"true"`
	OutputPath string `long:"output-path" env:"OUTPUT_PATH" required:"true"`
	ServerPort int    `long:"grpc-port" env:"GRPC_PORT" required:"true"`
	ServerHost string `long:"grpc-host" env:"GRPC_HOST" required:"true"`
	config
}

type config struct {
	LogLevel string `long:"log-level" env:"MF_LOG_LEVEL" required:"false" default:"info"`
}

func (c *config) GetLogLevel() zerolog.Level {
	l, err := zerolog.ParseLevel(strings.ToLower(c.LogLevel))
	if err != nil || l == zerolog.NoLevel {
		l = zerolog.InfoLevel
	}

	return l
}
