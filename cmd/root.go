package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
				fx.Provide(zap.NewDevelopment),
				fx.Invoke(register),
				fx.Invoke(ftp.New),
				fx.Invoke(web.New),
			)
			app.Run()
			return app.Err()

		},
	}
)

func register(lifecycle fx.Lifecycle, ftps ftp.FtpServer) {
   
    lifecycle.Append(
        fx.Hook{
            OnStart: func(context.Context) error {
                if err := ftps.Start(); err != nil {
					return err
				}
            },
            OnStop: func(ctx context.Context) error {
               if ftps != nil {
				   ftps.Stop()
			   }
            }
        }
    )
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().IntVar(&p.WebPort,"web-port", 9000, "Web server port")
	rootCmd.PersistentFlags().IntVar(&p.FtpPort, "ftp-port", 2121, "FTP server port")

	rootCmd.PersistentFlags().StringVar(&p.Host, "host", "0.0.0.0", "Host address to bind to")
	rootCmd.PersistentFlags().StringVar(&p.FtpPassword, "ftp-password", "admin", "Password to use for FTP")
	
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}