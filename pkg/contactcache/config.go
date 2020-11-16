package contactcache

import "github.com/spf13/viper"

//defaultConfig sets the main default configs
func defaultConfig() {
	viper.SetDefault("cache.address", "localhost:6379s")
}
