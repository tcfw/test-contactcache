package contactcache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/spf13/viper"
)

func setupTestServer(t *testing.T, resp string) (*Server, *int, func(), *miniredis.Miniredis) {
	var beReqCount int

	//Spin up backend
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beReqCount++
		fmt.Fprintln(w, resp)
	}))

	//Spin up cache
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	//Set config to test redis
	viper.Set("cache.address", s.Addr())

	cache, err := NewRedisCache()
	if err != nil {
		t.Fatal(err)
	}

	beURL, _ := url.Parse(ts.URL)
	be := httputil.NewSingleHostReverseProxy(beURL)

	srv := &Server{
		be:    be,
		cache: cache,
	}
	be.ModifyResponse = srv.handleProxyResponse

	closer := func() {
		ts.Close()
		s.Close()
	}

	return srv, &beReqCount, closer, s
}
