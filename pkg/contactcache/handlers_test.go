package contactcache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthChec(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	}))
	defer ts.Close()

	beURL, _ := url.Parse(ts.URL)
	be := httputil.NewSingleHostReverseProxy(beURL)

	srv := &Server{
		be: be,
	}
	handler := srv.httpHandler()

	//No Key
	req, _ := http.NewRequest("GET", "https://anywhere.local/", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Result().StatusCode)

	//Add API key
	req.Header.Add(apiKeyHeader, "1234")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.NotEqual(t, 200, w.Result().StatusCode)
}

func TestPrefixKey(t *testing.T) {
	srv := &Server{}

	req, _ := http.NewRequest("GET", "https://anywhere.local/v1/contact", nil)
	req.Header.Add(apiKeyHeader, "1234")

	key := srv.prefixKey(req, "4321")
	assert.Equal(t, "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4:4321", key)
}

func TestIsPersonKey(t *testing.T) {
	srv := &Server{}

	//Email should be false
	t1 := srv.isPersonKey("test@example.com")
	assert.False(t, t1)

	t2 := srv.isPersonKey("person_9EAF39E4-9AEC-4134-964A-D9D8D54162E7")
	assert.True(t, t2)
}
