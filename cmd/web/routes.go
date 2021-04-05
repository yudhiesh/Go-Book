package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *Application) routes() http.Handler {

	// recoverPanic <-> logRequest <-> secureHeaders <-> servemux <-> application handler
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux := pat.New()
	// Order matters here as the "/snippet/create" is valid for GET and POST
	mux.Get("/", http.HandlerFunc(app.home))
	mux.Get("/snippet/create", http.HandlerFunc(app.createSnippetForm))
	mux.Post("/snippet/create", http.HandlerFunc(app.createSnippet))
	mux.Get("/snippet/:id", http.HandlerFunc(app.showSnippet))

	fileServer := http.FileServer(http.Dir(config.StaticDir))

	// Any request that matches the start of "/static/" will be dispatched to
	// the corresponding handler
	mux.Get("/static/", http.StripPrefix("/static", neuter(fileServer)))

	return standardMiddleware.Then(mux)
}
