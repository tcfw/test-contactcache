package contactcache

import "github.com/spf13/viper"

//defaultConfig sets the main default configs
func defaultConfig() {
	viper.SetDefault("cache.address", "127.0.0.1:6379")
}
