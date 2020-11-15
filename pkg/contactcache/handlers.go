package contactcache

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	apiKeyHeader = "autopilotapikey"
	noAPIKey     = `{"error":"Bad Request", "message": "No autopilotapikey header provided."}`
)

func (s *Server) httpHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/v1/contact/{idOrEmail}", s.handleGetContact).Methods("GET")
	r.HandleFunc("/v1/contact/{idOrEmail}", s.handleDeleteContact).Methods("DELETE")
	r.HandleFunc("/v1/contact", s.handleUpsertContact).Methods("POST")
	r.HandleFunc("/v1/contacts", s.handleListContact).Methods("GET")

	//Passthrough all over requests
	r.PathPrefix("/").HandlerFunc(s.be.ServeHTTP)

	//Check for API key
	r.Use(s.authCheck)

	return r
}

func (s *Server) authCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//Basic check for API key existance
		//TODO(tcfw) pass to api key validator if no already checked by another middleware
		if r.Header.Get(apiKeyHeader) == "" {
			w.Header().Add("content-type", "application/json")
			http.Error(w, noAPIKey, http.StatusBadRequest)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (s *Server) handleProxyResponse(r *http.Response) error {

	return nil
}

func (s *Server) handleGetContact(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) handleUpsertContact(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) handleDeleteContact(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) handleListContact(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) prefixKey(r *http.Request, key string) string {
	apiKey := r.Header.Get(apiKeyHeader)
	return fmt.Sprintf("%x:%s", sha256.Sum256([]byte(apiKey)), key)
}

func (s *Server) isPersonKey(key string) bool {
	return strings.Contains(key, "person_") && !strings.Contains(key, "@")
}
