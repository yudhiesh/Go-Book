package main

import (
	"log"
	"net/http"
)

// NOTE:
// You can use http.Handle() and http.HandleFunc() to register routes without
// declaring a servemux
// This allows the code for creating the routes to be a bit shorter
// These functions register their routes with something called the
// DefaultServeMux which is initialized by default and stored in a net/http
// global variable
// As it is a global variable, any package can access it and register a route,
// this includes any third-party packages that your application imports
// Through this someone could expose a malicious handler to the web.
// tl;dr its a security issue

// Define a home handler function which writes a byte slice containing "Hello
// from Snippetbox" as the response body
func home(w http.ResponseWriter, r *http.Request) {
	// To overcome the catch-all nature of the "/" for the home path check if it
	// doesn't match the specific path and if it does not throw an error
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Hello from home"))
}

func showSnippet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a specific snippet"))
}

func createSnippet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new snippet"))
}

func main() {
	// Use the http.NewServeMux() function to initialize a new servemux, then
	// register the home function as the handler fo the "/" URL pattern

	// Longer URL patterns always take precedence over shorter ones
	// So if there are multiple patterns which match a request, it will always
	// dispatch the request to the handler corresponding to the longest pattern.
	mux := http.NewServeMux()
	// Subtree path → ends with a trailing slash
	// Acts as a catch-all as it essentially means match a single slash,
	// followed by anything(or nothing at all)
	mux.HandleFunc("/", home)
	// Fixed path → ends with no slash
	mux.HandleFunc("/snippet", showSnippet)
	mux.HandleFunc("/snippet/create", createSnippet)

	// Use the http.ListenAndServe() function to start a new web server.
	// We pass in two parameters: the TCP network address to listen on
	// and the servemux we just created
	// If http.ListenAndServe() returns an error
	// log.Fatal will log the error message and exit
	// Any error returned by http.ListenAndServe() is always non-nil
	log.Println("Starting server on :4000")
	// TCP network address should be in the format "host:port"
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)

}
