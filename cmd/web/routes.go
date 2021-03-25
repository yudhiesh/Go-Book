package main

import (
	"net/http"
	"strings"
)

func (app *Application) routes() *http.ServeMux {

	mux := http.NewServeMux()

	// HandleFunc takes in normal functions that are not actually Handlers as
	// they do not have the method ServeHTTP
	mux.HandleFunc("/", app.home)
	// NOTE: If you wanted to turn home into an actual handler you would need to
	// instantiate an interface home and then turn the home function into a
	// ServeHTTP method
	// Then pass it by pointer as below:
	// mux.Handle("/", &home{})
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/snippet/create", app.createSnippet)

	fileServer := http.FileServer(http.Dir(config.StaticDir))

	mux.Handle("/static/", http.StripPrefix("/static", neuter(fileServer)))

	return mux
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
