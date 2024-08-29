package config

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/fmartingr/notion2ical/internal/notion"
	"github.com/sethvargo/go-envconfig"
)

// readDotEnv reads the configuration from variables in a .env file (only for contributing)
func readDotEnv(logger *slog.Logger) map[string]string {
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
		logger.Error("error reading dotenv", slog.String("err", err.Error()))
	}

	return result
}

type Config struct {
	Hostname string `env:"HOSTNAME,required"`
	LogLevel string `env:"LOG_LEVEL,default=info"`
	Http     struct {
		Enabled        bool   `env:"HTTP_ENABLED,default=True"`
		Port           int    `env:"HTTP_PORT,default=8080"`
		PublicHostname string `env:"HTTP_PUBLIC_HOSTNAME,required"`
	}
	Branding struct {
		ThanksMessage      string `env:"BRANDING_THANKS_MESSAGE"`
		FooterExtraMessage string `env:"BRANDING_FOOTER_EXTRA"`
	}
	Notion struct {
		IntegrationToken string               `env:"NOTION_INTEGRATION_TOKEN,required"`
		MaxPagination    int                  `env:"NOTION_MAX_PAGINATION,default=2"`
		Client           *notion.NotionClient // Must be manually set up
	}
	Routes struct {
		Calendar struct {
			CacheExpiration   time.Duration `env:"ROUTES_CACHE_EXPIRATION,default=24h"`
			CacheControl      bool          `env:"ROUTES_CACHE_CONTROL,default=true"`
			LimiterMaxRequest int           `env:"ROUTES_CALENDAR_LIMITER_MAX_REQUESTS,default=2"`
			LimiterExpiration time.Duration `env:"ROUTES_CALENDAR_LIMITER_DURATION,default=1s"`
		}
		Static struct {
			Path   string        `env:"ROUTES_STATIC_PATH,default=/static"`
			MaxAge time.Duration `env:"ROUTES_STATIC_MAX_AGE,default=720h"`
		}
		System struct {
			Path string `env:"ROUTES_SYSTEM_PATH,default=/system"`
		}
	}
}

func ParseServerConfiguration(ctx context.Context, logger *slog.Logger) (*Config, error) {
	var cfg Config

	lookuper := envconfig.MultiLookuper(
		envconfig.MapLookuper(map[string]string{"HOSTNAME": os.Getenv("HOSTNAME")}),
		envconfig.MapLookuper(readDotEnv(logger)),
		envconfig.PrefixLookuper("NOTION2ICAL_", envconfig.OsLookuper()),
		envconfig.OsLookuper(),
	)
	if err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   &cfg,
		Lookuper: lookuper,
	}); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	return &cfg, nil
}
