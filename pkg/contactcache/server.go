package contactcache

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//NewServer creates a new instance of the middleware
func NewServer() (*Server, error) {
	srv := &Server{
		log: logrus.New(),
	}

	//New up a redis endpoint
	cache, err := NewRedisCache()
	if err != nil {
		return nil, err
	}
	srv.cache = cache

	//Parse backend
	backendAddress := viper.GetString("backend.address")
	if backendAddress == "" {
		return nil, fmt.Errorf("no backend endpoint provided")
	}

	backend, err := url.Parse(backendAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse backend URL: %s", err)
	}

	//Set backend reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(backend)
	srv.be = proxy

	return srv, nil
}

//Server primary content server
type Server struct {
	log   *logrus.Logger
	be    *httputil.ReverseProxy
	cache Cacher
}

//Start starts serving https requests
func (s *Server) Start() error {
	tlsConfig, err := s.tlsConfig()
	if err != nil {
		s.log.Fatalf("TLS config failed: %s", err)
		return err
	}

	handler := s.httpHandler()
	handler = handlers.LoggingHandler(s.log.Out, handler)

	httpSrv := http.Server{
		Addr:      viper.Get("address").(string),
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	cert := viper.Get("tls.cert").(string)
	key := viper.Get("tls.key").(string)

	s.log.Info("Starting HTTPs endpoint")

	return httpSrv.ListenAndServeTLS(cert, key)
}
