package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"yudhiesh/snippetbox/pkg/models"

	"github.com/justinas/nosurf"
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
		// If the user is not authenticated then redirect them to the login
		// page
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		// Else set the Cache-Control: no-store header so that pages require
		// authentication are not stored in the users browser cache
		w.Header().Set("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

// Authenticates the user middleware
// When the user is not authenticated and not active pass the unchanged request
// to the next handler in the chain.
// When the user is authenticated and active, create a copy of the request with
// a contextIsAuthenticated key and true value stored in the request context.
// Then pass this copy of the context to the request.
func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the authenticated user ID value exists in the session
		// If there isn't one then just continue the chain as normal
		exists := app.session.Exists(r, "authenticatedUserID")
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		// Fetch the details from the current user in the database
		// If not matching record or the user is not active(deactivated their
		// account) then remove the authenticatedID value from their session
		user, err := app.users.Get(app.session.GetInt(r, "authenticatedUserID"))
		if errors.Is(err, models.ErrNoRecord) || !user.Active {
			app.session.Remove(r, "authenticatedUserID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}

		// Otherwise we know the user is authenticated and active
		// So we add in the contextIsAuthenticated value of true to the context
		// to a copy of the request
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
