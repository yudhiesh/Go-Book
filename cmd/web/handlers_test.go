package main

import (
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	// Shutdown the sever when the test is over
	defer ts.Close()

	statusCode, _, body := ts.get(t, "/ping")
	if statusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, statusCode)
	}
	if string(body) != "OK" {
		t.Errorf("want body to equal %q; got %q", "OK", string(body))
	}

}
