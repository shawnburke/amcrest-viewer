package cmd

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/shawnburke/amcrest-viewer/cameras"
	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/ftp"
	"github.com/shawnburke/amcrest-viewer/ingest"
	"github.com/shawnburke/amcrest-viewer/storage"
	web "github.com/shawnburke/amcrest-viewer/web"
)

var (
	// Used for flags.

	p = common.Params{}

	rootCmd = &cobra.Command{
		Use:   "amcrest-viewer",
		Short: "A private viewer and storage system for home cameras",
		RunE: func(cmd *cobra.Command, args []string) error {

			app := fx.New(buildGraph(nil))

			app.Run()
			return app.Err()

		},
	}
)

func buildGraph(cfg config.Provider) fx.Option {

	configFunc := yamlConfig

	if cfg != nil {
		configFunc = func(_ *common.Params) (config.Provider, error) {
			return cfg, nil
		}
	}

	return fx.Options(
		fx.Provide(func() *common.Params {
			return &p
		}),
		// basics
		fx.Provide(configFunc),
		fx.Provide(logger),
		fx.Provide(tz),
		fx.Provide(common.NewTime),
		fx.Provide(common.NewEventBus),

		// main modules
		ingest.Module,
		storage.Module,
		cameras.Module,

		// servers
		fx.Provide(ftp.New),
		fx.Provide(web.New),

		// giddyup
		fx.Invoke(register),
	)
}

func tz() *time.Location {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	return loc
}

func logger(cp config.Provider) (*zap.Logger, error) {

	cfg := zap.NewDevelopmentConfig()

	level := zapcore.DebugLevel

	val := cp.Get("log.level")

	if val.HasValue() && level.UnmarshalText([]byte(val.String())) == nil {
		cfg.Level = zap.NewAtomicLevelAt(level)
	}

	return cfg.Build()
}

func register(lifecycle fx.Lifecycle, ftps ftp.FtpServer, web web.HttpServer, logger *zap.Logger) {

	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				errFtp := ftps.Start()

				if errFtp != nil {
					logger.Error("Error starting ftp", zap.Error(errFtp))
					return errFtp
				}

				errWeb := web.Start()
				if errWeb != nil {
					logger.Error("Error starting web", zap.Error(errWeb))
					ftps.Stop()
					return errWeb
				}
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if err := ftps.Stop(); err != nil {
					logger.Error("Error shutting down ftp", zap.Error(err))
				}
				if err := web.Stop(); err != nil {
					logger.Error("Error shutting down web", zap.Error(err))
				}
				return nil
			},
		},
	)
}

func yamlConfig(p *common.Params) (config.Provider, error) {

	configDir := p.GetConfigDir()

	files := []string{path.Join(configDir, "base.yaml")}
	env := os.Getenv("ENVIRONMENT")
	if env != "" {
		files = append(files, path.Join(configDir, env+".yaml"))
	}
	return config.NewYAMLProviderFromFiles(files...)
}

// Execute executes the root command.
func Execute() error {
	rootCmd.PersistentFlags().IntVar(&p.WebPort, "web-port", 9000, "Web server port")
	rootCmd.PersistentFlags().IntVar(&p.FtpPort, "ftp-port", 2121, "FTP server port")
	rootCmd.PersistentFlags().StringVar(&p.ConfigDir, "config-dir", "", "Config dir")
	rootCmd.PersistentFlags().StringVar(&p.DataDir, "data-dir", "", "Data directory root (for files and DB)")
	rootCmd.PersistentFlags().StringVar(&p.FrontendDir, "frontend-dir", "", "Frontend directory root (for web)")

	rootCmd.PersistentFlags().StringVar(&p.Host, "host", "0.0.0.0", "Host address to bind to")
	rootCmd.PersistentFlags().StringVar(&p.FtpPassword, "ftp-password", "admin", "Password to use for FTP")

	return rootCmd.Execute()
}
