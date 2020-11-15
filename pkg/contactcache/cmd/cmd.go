package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//NewContactCacheCmd provides a root cmd
func NewContactCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contactcache",
		Short: "Contactcache proxies the contacts endpoint and provides a caching middleware",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(newServeCmd())

	viper.SetConfigName("config")              // name of config file (without extension)
	viper.SetConfigType("yaml")                // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/contactcache/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.contactcache") // call multiple times to add many search paths
	viper.AddConfigPath(".")                   // optionally look for config in the working directory

	return cmd
}
