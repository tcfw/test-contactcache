package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tcfw/test-contactcache/pkg/contactcache"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					// Config file not found; ignore error
				} else {
					panic(err)
				}
			}

			srv, err := contactcache.NewServer()
			if err != nil {
				panic(err)
			}
			if err := srv.Start(); err != nil {
				panic(err)
			}
		},
	}

	cmd.Flags().String("tls-key", "", "TLS key")
	cmd.Flags().String("tls-cert", "", "TLS cert")
	cmd.Flags().StringP("listen", "l", ":443", "Listening address")
	cmd.Flags().String("metrics-listen", ":9102", "Metrics listening address")
	cmd.Flags().StringP("target", "t", "https://api2.autopilothq.com", "Metrics listening address")

	viper.BindPFlag("tls.key", cmd.Flags().Lookup("tls-key"))
	viper.BindPFlag("tls.cert", cmd.Flags().Lookup("tls-cert"))
	viper.BindPFlag("address", cmd.Flags().Lookup("listen"))
	viper.BindPFlag("backend.address", cmd.Flags().Lookup("target"))
	viper.BindPFlag("metrics.address", cmd.Flags().Lookup("metrics-listen"))

	return cmd
}
