package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"
	"yudhiesh/snippetbox/pkg/models/mock"

	"github.com/golangcollege/sessions"
)

// Creates a new instance of the Application struct and adds in the loggers
// needed for testing the middleware
func newTestApplication(t *testing.T) *Application {
	templateCache, err := newTemplateCache("./../../ui/html/")
	if err != nil {
		t.Fatal(err)
	}

	session := sessions.New([]byte("3dSm5MnygFHh7XidAtbskXrjbwfoJcbJ"))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	return &Application{
		// Logger is needed by every middleware
		// Without these two there would be a panic
		errorLog:      log.New(ioutil.Discard, "", 0),
		infoLog:       log.New(ioutil.Discard, "", 0),
		session:       session,
		snippets:      &mock.SnippetModel{},
		templateCache: templateCache,
		users:         &mock.UserModel{},
	}
}

// Custom testServer which anonymously embeds a httptest.Server instance
type testServer struct {
	*httptest.Server
}

// Initializes and returns a new instance of the custom testServer type
func newTestServer(t *testing.T, h http.Handler) *testServer {
	// Creates a new test server passing in the value returned by our
	// app.routes() method as the handler for the server.
	// use httptest.NewServer() if testing http request
	ts := httptest.NewTLSServer(h)
	// Store any cookies sent in a HTTPS response, so that we can include them
	// in any subsequent requests back to the test server.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar
	// Don't automatically follow redirects, instead return the first HTTPS
	// response sent by our server so that we can test the reponse for that
	// specific request.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

// Get method to the custom testServer type. This makes a GET request to the
// given urlPath and returns the StatusCode, Header, and the body
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	return rs.StatusCode, rs.Header, body
}
