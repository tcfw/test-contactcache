package contactcache

import (
	"net/http"

	"github.com/gorilla/mux"
)

func httpHandler(s *Server) http.Handler {
	r := mux.NewRouter()

	return r
}
