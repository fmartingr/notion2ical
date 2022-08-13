package server

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

// readDotEnv reads the configuration from variables in a .env file (only for contributing)
func readDotEnv(logger *zap.Logger) map[string]string {
	file, err := os.Open(".env")
	if err != nil {
		return nil
	}
	defer file.Close()

	result := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		keyval := strings.SplitN(line, "=", 2)
		result[keyval[0]] = keyval[1]
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal("error reading dotenv", zap.Error(err))
	}

	return result
}

type ServerConfig struct {
	Hostname string `env:"HOSTNAME,required"`
	Http     struct {
		Enabled bool `env:"HTTP_ENABLED,default=True"`
		Port    int  `env:"HTTP_PORT,default=8080"`
	}
	LogLevel string `env:"LOG_LEVEL,default=info"`
	Notion   struct {
		IntegrationToken string `env:"NOTION_INTEGRATION_TOKEN"`
	}
}

func ParseServerConfiguration(ctx context.Context, logger *zap.Logger) *ServerConfig {
	var cfg ServerConfig

	lookuper := envconfig.MultiLookuper(
		envconfig.MapLookuper(map[string]string{"HOSTNAME": os.Getenv("HOSTNAME")}),
		envconfig.MapLookuper(readDotEnv(logger)),
		envconfig.PrefixLookuper("NOTION2ICAL_", envconfig.OsLookuper()),
		envconfig.OsLookuper(),
	)
	if err := envconfig.ProcessWith(ctx, &cfg, lookuper); err != nil {
		logger.Fatal("Error parsing configuration: %s", zap.Error(err))
	}

	return &cfg
}
