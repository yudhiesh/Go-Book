package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *Application) routes() http.Handler {

	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

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

	// recoverPanic <-> logRequest <-> secureHeaders <-> servemux <-> application handler
	return standardMiddleware.Then(mux)
}
