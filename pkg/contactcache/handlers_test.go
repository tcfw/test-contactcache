package contactcache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestPrefixKey(t *testing.T) {
	srv := &Server{}

	req, _ := http.NewRequest("GET", "https://anywhere.local/v1/contact", nil)
	req.Header.Add(apiKeyHeader, "1234")

	key := srv.prefixKey("1234", "4321")
	assert.Equal(t, "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4:contact:4321", key)
}

func TestIsPersonKey(t *testing.T) {
	srv := &Server{}

	//Email should be false
	t1 := srv.isPersonKey("test@example.com")
	assert.False(t, t1)

	t2 := srv.isPersonKey("person_9EAF39E4-9AEC-4134-964A-D9D8D54162E7")
	assert.True(t, t2)
}

func TestAuthCheck(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestHandleGetContact(t *testing.T) {
	var beReqCount int

	contact := `{
"contact_id": "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23",
"FirstName": "Chris",
"LastName": "Sharkey"
"Email": "chris@autopilothq.com"
}`

	//Spin up backend
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beReqCount++
		fmt.Fprintln(w, contact)
	}))
	defer ts.Close()

	//Spin up cache
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()
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

	handler := srv.httpHandler()
	w := httptest.NewRecorder()

	apiKey := "1234"

	req, err := http.NewRequest("GET", "https://anywhere.local/v1/contact/chris@autopilothq.com", nil)
	if err != nil {
		t.Fatal(t, err)
	}
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Check backend was called
	assert.Equal(t, 1, beReqCount)

	//Check cache entry made
	mainCacheKey := srv.prefixKey(apiKey, "chris@autopilothq.com")
	aliasCacheKey := srv.prefixKey(apiKey, "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23")

	//Check matching
	ce, _ := s.Get(mainCacheKey)
	assert.Equal(t, contact+"\n", ce)

	//check alias
	ce, _ = s.Get(aliasCacheKey)
	assert.Equal(t, mainCacheKey, ce)

	//Run again to use cached response
	handler.ServeHTTP(w, req)
	assert.Equal(t, 1, beReqCount)

}
