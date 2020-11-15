package contactcache

import (
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//NewServer creates a new instance of the middleware
func NewServer() (*Server, error) {
	srv := &Server{
		log: logrus.New(),
	}

	return srv, nil
}

//Server primary content server
type Server struct {
	log   *logrus.Logger
	be    httputil.ReverseProxy
	cache Cacher
}

//Start starts serving https requests
func (s *Server) Start() error {
	tlsConfig, err := s.tlsConfig()
	if err != nil {
		s.log.Fatalf("TLS config failed: %s", err)
		return err
	}

	handler := httpHandler(s)

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
