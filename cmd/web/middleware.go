package main

import (
	"fmt"
	"net/http"
	"strings"
)

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a defered function which will always be run in the event of a
		// panic as Go unwinds the stack
		defer func() {
			// Use the builtin recover() to check if there has been a panic or
			// not
			if err := recover(); err != nil {
				// If there is a panic then close the connection
				w.Header().Set("Connection", "close")
				// Return a 500 Internal Server response
				// Normalize the error to create a new error object containing
				// the defautl textual representation of the interface{}
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Adds two headers: X-Frame-Options: deny and X-XSS-Protection: 1; mode=block
// to every response
// These headers help prevent XSS and Clickjacking attacks
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		// If you have a return statement here before the next() function then
		// the chain of execution will stop and control will flow back upstream.
		// Any code here will execute on the way down the chain.
		next.ServeHTTP(w, r)
		// Any code here will execute on the way back up the chain.
	})
}

// Disable http.FileServer Directory Listings
// NOTE: Prevent navigable directory listings and disable it all together
func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *Application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		w.Header().Set("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
