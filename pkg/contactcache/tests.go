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

//setupTestServer helper for setting up a mock backend and in-mem redis server with a single text response
func setupTestServer(t *testing.T, resp string) (*Server, *int, func(), *miniredis.Miniredis) {
	var beReqCount int

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beReqCount++
		fmt.Fprintln(w, resp)
	})

	srv, closer, cache := setupTestServerHandleFunc(t, handler)

	return srv, &beReqCount, closer, cache
}

//setupTestServer helper for setting up a mock backend and in-mem redis server passing through a
//handler for complex/multiple response (e.g. via mutex)
func setupTestServerHandleFunc(t *testing.T, handler http.HandlerFunc) (*Server, func(), *miniredis.Miniredis) {
	//Spin up backend
	ts := httptest.NewServer(handler)

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

	return srv, closer, s
}
