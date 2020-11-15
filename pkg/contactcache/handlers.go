package contactcache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
)

const (
	apiKeyHeader = "autopilotapikey"
	noAPIKey     = `{"error":"Bad Request", "message": "No autopilotapikey header provided."}`

	cacheTTL = 60 * time.Minute
)

func (s *Server) httpHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/v1/contact/{idOrEmail}", s.handleGetContact).Methods(http.MethodGet)
	r.HandleFunc("/v1/contact/{idOrEmail}", s.handleDeleteContact).Methods(http.MethodDelete)
	r.HandleFunc("/v1/contact", s.handleUpsertContact).Methods(http.MethodPost)
	r.HandleFunc("/v1/contacts", s.handleListContact).Methods(http.MethodGet)
	r.HandleFunc("/v1/contacts/{bookmark}", s.handleListContact).Methods(http.MethodGet)

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
	//Cache the response from the backend server

	apiKey := r.Request.Header.Get(apiKeyHeader)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	//Get contact
	if strings.Index(r.Request.URL.Path, "/v1/contact/") == 0 && r.Request.Method == http.MethodGet {
		s.cacheContact(apiKey, string(b))
	}

	//Upsert contact
	if r.Request.URL.Path == "/v1/contact" && r.Request.Method == http.MethodPost {
		s.cacheContact(apiKey, string(b))
	}

	//List contacts
	if strings.Index(r.Request.URL.Path, "/v1/contacts") == 0 && r.Request.Method == http.MethodGet {
		s.cacheList(r.Request, apiKey, string(b))
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	return nil
}

func (s *Server) cacheContact(apiKey string, body string) error {
	//New ctx since outside of response routine
	ctx := context.Background()

	bulk := gjson.Get(body, "contacts").Array()

	if len(bulk) != 0 {
		for _, contact := range bulk {
			s.cacheContact(apiKey, contact.Raw)
		}

		return nil
	}

	email := gjson.Get(body, "Email").String()
	id := gjson.Get(body, "contact_id").String()

	cacheKey := s.prefixKey(apiKey, email)

	err := s.cache.Set(ctx, cacheKey, body, cacheTTL)
	if err != nil {
		s.log.WithError(err).Error("failed to set contact key")
		return err
	}

	//Add contact/person id alias
	err = s.cache.Set(ctx, s.prefixKey(apiKey, id), cacheKey, cacheTTL)
	if err != nil {
		s.log.WithError(err).Error("failed to set contact key")
		return err
	}

	return nil
}

func (s *Server) cacheList(r *http.Request, apiKey, body string) error {
	//New ctx since outside of response routine
	ctx := context.Background()

	var bookmark string

	if strings.Contains(r.URL.Path, "/contacts/person_") {
		//is Bookmarked
		bookmark = r.URL.Path[len("/v1/contacts/"):]
	}

	cacheKey := s.prefixKey(apiKey, fmt.Sprintf("lists:%s", bookmark))

	err := s.cache.Set(ctx, cacheKey, body, cacheTTL)
	if err != nil {
		s.log.WithError(err).Error("failed to set contact key")
		return err
	}

	return nil
}

func (s *Server) handleGetContact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idOrEmail := vars["idOrEmail"]

	apiKey := r.Header.Get(apiKeyHeader)

	var cacheKey string
	var val string
	var err error

	if idOrEmail == "" {
		goto passthrough
	}

	//Check if is a person key or email
	if s.isPersonKey(idOrEmail) {
		//Find the contact key for email
		realKey, err := s.cache.Get(r.Context(), s.prefixKey(apiKey, idOrEmail))
		if realKey == "" || err != nil {
			//passthrough
			goto passthrough
		}

		cacheKey = realKey
	} else {
		cacheKey = s.prefixKey(apiKey, idOrEmail)
	}

	val, err = s.cache.Get(r.Context(), cacheKey)
	if err != nil || val == "" {
		goto passthrough
	}

	w.Header().Add("content-type", "application/json")
	w.Write([]byte(val))

	return

passthrough:
	s.be.ServeHTTP(w, r)
}

func (s *Server) handleUpsertContact(w http.ResponseWriter, r *http.Request) {
	//Invalidate existing

	apiKey := r.Header.Get(apiKeyHeader)

	vars := mux.Vars(r)
	idOrEmail := vars["idOrEmail"]

	if err := s.invalidateContact(r.Context(), apiKey, idOrEmail); err != nil {
		s.log.WithError(err).Error("failed to invalidate contact cache")
	}

	//passthrough to be cached
	s.be.ServeHTTP(w, r)
}

func (s *Server) handleDeleteContact(w http.ResponseWriter, r *http.Request) {
	//passthrough
	s.be.ServeHTTP(w, r)

	//Invalidate cache

	apiKey := r.Header.Get(apiKeyHeader)

	vars := mux.Vars(r)
	idOrEmail := vars["idOrEmail"]

	if err := s.invalidateContact(r.Context(), apiKey, idOrEmail); err != nil {
		s.log.WithError(err).Error("failed to invalidate contact cache")
	}

}

func (s *Server) invalidateContact(ctx context.Context, apiKey string, idOrEmail string) error {
	//Check if is a person key or email
	var cacheKey string
	if s.isPersonKey(idOrEmail) {
		//Find the contact key for email
		realKey, err := s.cache.Get(ctx, s.prefixKey(apiKey, idOrEmail))
		if realKey == "" || err != nil {
			//passthrough
			return err
		}

		cacheKey = realKey
	} else {
		cacheKey = s.prefixKey(apiKey, idOrEmail)
	}

	s.cache.Delete(ctx, cacheKey)

	//Invalidate lists
	s.cache.Delete(ctx, s.prefixKey(apiKey, "lists:*"))

	return nil
}

func (s *Server) handleListContact(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get(apiKeyHeader)

	vars := mux.Vars(r)
	bookmark := vars["bookmark"]

	//TODO(tcfw) validate bookmark format

	cacheKey := s.prefixKey(apiKey, fmt.Sprintf("lists:%s", bookmark))
	val, err := s.cache.Get(r.Context(), cacheKey)
	if val == "" || err != nil {
		goto passthrough
	}

	w.Header().Add("content-type", "application/json")
	w.Write([]byte(val))

	return

	//passthrough
passthrough:
	s.be.ServeHTTP(w, r)
}

func (s *Server) prefixKey(apiKey string, key string) string {
	return fmt.Sprintf("%x:contact:%s", sha256.Sum256([]byte(apiKey)), key)
}

func (s *Server) isPersonKey(key string) bool {
	return strings.Contains(key, "person_") && !strings.Contains(key, "@")
}
