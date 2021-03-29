package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Writes an error message and stack trace to the errorLog
func (app *Application) serverError(w http.ResponseWriter, err error) {
	// debug.Stack() is used to get a stack trace for the current goroutine and
	// append it to log message
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// Report the file name and the stack trace one step back in the stack trace
	// using errorLog.Output() where the frame depth is set to 2
	app.errorLog.Output(2, trace)
	// http.StatusText() returns a human-readable format string of the http
	// server error
	// http.StatusText(400) → "Bad Request"
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Sends a specific status code and corresponding description to the user
func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// Convenience wrapper for sending a 404 Not Found response to the user
func (app *Application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *Application) render(w http.ResponseWriter, r *http.Request, name string, tf *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	err := ts.Execute(w, tf)
	if err != nil {
		app.serverError(w, err)
	}
}
