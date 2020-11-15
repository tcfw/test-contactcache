package main

import (
	"github.com/tcfw/test-contactcache/pkg/contactcache/cmd"
)

func main() {
	cmd := cmd.NewContactCacheCmd()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
