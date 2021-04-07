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
	// Does not need to be applied to every route such as the /static/ route
	dynamicMiddleware := alice.New(app.session.Enable, noSurf)

	mux := pat.New()
	// Order matters here as the "/snippet/create" is valid for GET and POST
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	// Add requireAuthentication middlewarte to the routes that require it
	mux.Get("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))

	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))

	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Any request that matches the start of "/static/" will be dispatched to
	// the corresponding handler
	mux.Get("/static/", http.StripPrefix("/static", neuter(fileServer)))

	return standardMiddleware.Then(mux)
}
