package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *Application) routes() http.Handler {

	// recoverPanic <-> logRequest <-> secureHeaders <-> servemux <-> application handler
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// This middleware loads and saves session data to and from the session
	// cookie with every HTTP request and response as appropriate
	dynamicMiddleware := alice.New(app.session.Enable)

	mux := pat.New()
	// Order matters here as the "/snippet/create" is valid for GET and POST
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/snippet/create", dynamicMiddleware.ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))

	fileServer := http.FileServer(http.Dir(config.StaticDir))

	// Any request that matches the start of "/static/" will be dispatched to
	// the corresponding handler
	mux.Get("/static/", http.StripPrefix("/static", neuter(fileServer)))

	return standardMiddleware.Then(mux)
}
