package main

import (
	"html/template"
	"path/filepath"
	"time"

	"yudhiesh/snippetbox/pkg/forms"
	"yudhiesh/snippetbox/pkg/models"
)

// Define a templateData type to act as the holding structure for
// any dynamic data that we want to pass to our HTML templates.
// At the moment it only contains one field, but we'll add more
// to it as the build progresses.
type templateData struct {
	CurrentYear int
	Flash       string
	Form        *forms.Form
	Snippet     *models.Snippet
	Snippets    []*models.Snippet
}

// Returns a human readable formatted string of the time.Time object
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable
// Acts as a lookup between the names of our custom template functions and the
// functions themselves
// NOTE: Custom template functions like humanDate must return a single
// value(excluding the error)!
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Map to act as a cache
	cache := map[string]*template.Template{}

	// filepath.Glob() gets us a slice of all filepaths with the extension
	// '.page.tmpl'
	// this essentially gives us a slice of all the 'page' template for the
	// application
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	// Loop  through the pages
	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl') from the full file path
		// and assign it to the name variable.
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before
		// you call the ParseFiles() method.
		// This means we have to use template.New() to create an empty template
		// set, use the Funcs() method to register the template.FuncMap, and
		// then parse the file as normal.
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	return cache, nil
}
