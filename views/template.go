package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/operas-logicas/lenslocked/context"
	"github.com/operas-logicas/lenslocked/models"
)

type Template struct {
	htmlTpl *template.Template
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any, errs ...error) {
	// Clone the template
	htmlTpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}

	htmlTpl = htmlTpl.Funcs(
		template.FuncMap{
			// Replace csrfField func with gorilla/csrf one
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
			// TODO! Fix this:
			"errors": func() []string {
				var errorMessages []string
				for _, err := range errs {
					errorMessages = append(errorMessages, err.Error())
				}
				return errorMessages
			},
		},
	)
	
	// Set the content type header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create a buffer to store the final result from template execution
	var buf bytes.Buffer 

	// Execute the cloned template	
	err = htmlTpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}

	// Copy buffer to response writer if no errors (this is not efficient for large templates)
	io.Copy(w, &buf)
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
	// Add csrfField placeholder func to template before parsing
	htmlTpl := template.New(pattern[0])
	htmlTpl = htmlTpl.Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (*models.User, error) {
				return nil, fmt.Errorf("currentUser not implemented")
			},
			"errors": func() []string {
				return nil
			},
		},
	)

	// Parse the template
	htmlTpl, err := htmlTpl.ParseFS(fs, pattern...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}

	return Template{
		htmlTpl: htmlTpl,
	}, nil
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}
