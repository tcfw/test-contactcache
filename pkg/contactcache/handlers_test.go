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
	contact := `{
"contact_id": "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23",
"FirstName": "Chris",
"LastName": "Sharkey"
"Email": "chris@autopilothq.com"
}`

	srv, beReqCount, close, s := setupTestServer(t, contact)
	defer close()

	handler := srv.httpHandler()
	w := httptest.NewRecorder()

	apiKey := "1234"

	req, err := http.NewRequest("GET", "https://anywhere.local/v1/contact/chris@autopilothq.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Check backend was called
	assert.Equal(t, 1, *beReqCount)

	//Check cache entry made
	mainCacheKey := srv.prefixKey(apiKey, "chris@autopilothq.com")
	aliasCacheKey := srv.prefixKey(apiKey, "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23")

	//Check matching
	ce, _ := s.Get(mainCacheKey)
	assert.Contains(t, ce, contact)

	//check alias
	ce, _ = s.Get(aliasCacheKey)
	assert.Equal(t, mainCacheKey, ce)

	//Run again to use cached response
	handler.ServeHTTP(w, req)
	assert.Equal(t, 1, *beReqCount)

}

func TestHandleDeleteContact(t *testing.T) {
	contact := `{
"contact_id": "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23",
"FirstName": "Chris",
"LastName": "Sharkey"
"Email": "chris@autopilothq.com"
}`

	srv, beReqCount, close, s := setupTestServer(t, contact)
	defer close()

	handler := srv.httpHandler()
	w := httptest.NewRecorder()

	apiKey := "1234"

	cacheKey := srv.prefixKey(apiKey, "chris@autopilothq.com")

	s.Set(cacheKey, contact)

	req, err := http.NewRequest("DELETE", "https://anywhere.local/v1/contact/chris@autopilothq.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Backend should have been called
	assert.Equal(t, 1, *beReqCount)

	//Key should have been removed
	val, _ := s.Get(cacheKey)
	assert.Equal(t, "", val)
}

func TestHandleUpsertContact(t *testing.T) {
	contact := `{
"contact_id": "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23",
"FirstName": "Chris",
"LastName": "Sharkey"
"Email": "chris@autopilothq.com"
}`

	srv, beReqCount, close, s := setupTestServer(t, contact)
	defer close()

	handler := srv.httpHandler()
	w := httptest.NewRecorder()

	apiKey := "1234"

	req, err := http.NewRequest("POST", "https://anywhere.local/v1/contact", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Backend should have been called
	assert.Equal(t, 1, *beReqCount)

	//Cache value should exist
	cacheKey := srv.prefixKey(apiKey, "chris@autopilothq.com")
	val, _ := s.Get(cacheKey)
	assert.NotEqual(t, "", val)

	req, _ = http.NewRequest("GET", "https://anywhere.local/v1/contact/person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23", nil)
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Backend should NOT have been called
	assert.Equal(t, 1, *beReqCount)
}

func TestBulkUpsertContacts(t *testing.T) {
	contact := `{
  "contacts": [
    {
      "FirstName": "Slarty",
      "LastName": "Bartfast",
      "Email": "test@slarty.com",
      "custom": {
        "string--Test--Field": "This is a test"
      }
    },
    {
      "FirstName": "Jerry",
      "LastName": "Seinfeld",
      "Email": "jerry@seinfeld.com"
    },
    {
      "FirstName": "Elaine",
      "LastName": "Benes",
      "Email": "elaine@seinfeld.com"
    }
  ]
}`

	srv, beReqCount, close, s := setupTestServer(t, contact)
	defer close()

	handler := srv.httpHandler()
	w := httptest.NewRecorder()

	apiKey := "1234"

	req, err := http.NewRequest("POST", "https://anywhere.local/v1/contact", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Backend should have been called
	assert.Equal(t, 1, *beReqCount)

	contactEmail := "jerry@seinfeld.com"

	//Cache value should exist
	cacheKey := srv.prefixKey(apiKey, contactEmail)
	val, _ := s.Get(cacheKey)
	assert.NotEqual(t, "", val)

	req, _ = http.NewRequest("GET", "https://anywhere.local/v1/contact/"+contactEmail, nil)
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Backend should NOT have been called
	assert.Equal(t, 1, *beReqCount)
}

func TestHandleListContact(t *testing.T) {
	contactList := `{
		"contacts": [
			{
				"contact_id": "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23",
				"FirstName": "Chris",
				"LastName": "Sharkey"
				"Email": "chris@autopilothq.com"
			}
		],
		"total_contacts": 1
	}`

	srv, beReqCount, close, _ := setupTestServer(t, contactList)
	defer close()

	handler := srv.httpHandler()
	w := httptest.NewRecorder()

	apiKey := "1234"

	req, err := http.NewRequest("GET", "https://anywhere.local/v1/contacts", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(apiKeyHeader, apiKey)

	handler.ServeHTTP(w, req)

	//Backend should have been called
	assert.Equal(t, 1, *beReqCount)

	handler.ServeHTTP(w, req)

	//Backend should NOT have been called
	assert.Equal(t, 1, *beReqCount)
}

func TestHandleUpsertThenList(t *testing.T) {
	contactList := `{
		"contacts": [
		],
		"total_contacts": 1
	}`

	contact := `{
		"contact_id": "person_AP2-9cbf7ac0-eec5-11e4-87bc-6df09cc44d23",
		"FirstName": "Chris",
		"LastName": "Sharkey"
		"Email": "chris@autopilothq.com"
	}`

	srv, beReqCount, close, s := setupTestServer(t, contact)
	defer close()

	apiKey := "1234"

	//Setup existing cache
	listCacheKey := srv.prefixKey(apiKey, "lists:")
	s.Set(listCacheKey, contactList)

	req, err := http.NewRequest("POST", "https://anywhere.local/v1/contact", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(apiKeyHeader, apiKey)

	handler := srv.httpHandler()
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	//Backend should have been called
	assert.Equal(t, 1, *beReqCount)

	//Cache should have invalidated
	val, _ := s.Get(listCacheKey)
	assert.Equal(t, "", val)
}
