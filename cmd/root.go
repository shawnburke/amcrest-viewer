package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/ftp"
	"github.com/shawnburke/amcrest-viewer/web"
)

var (
	// Used for flags.

	p = common.Params{}

	rootCmd = &cobra.Command{
		Use:   "amcrest-viewer",
		Short: "A private viewer and storage system for home cameras",
		RunE: func(cmd *cobra.Command, args []string) error {

			app := fx.New(
				fx.Provide(func() *common.Params {
					return &p
				}),
				fx.Provide(yamlConfig),
				fx.Provide(common.NewConfigAuth),
				fx.Provide(zap.NewDevelopment),
				fx.Provide(ftp.New),
				fx.Provide(web.New),
				fx.Invoke(register),
			)
			app.Run()
			return app.Err()

		},
	}
)

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

func yamlConfig() (config.Provider, error) {

	return config.NewYAMLProviderFromFiles("./config/base.yaml")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().IntVar(&p.WebPort, "web-port", 9000, "Web server port")
	rootCmd.PersistentFlags().IntVar(&p.FtpPort, "ftp-port", 2121, "FTP server port")

	rootCmd.PersistentFlags().StringVar(&p.Host, "host", "0.0.0.0", "Host address to bind to")
	rootCmd.PersistentFlags().StringVar(&p.FtpPassword, "ftp-password", "admin", "Password to use for FTP")

}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}
